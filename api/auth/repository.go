package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-sql-driver/mysql"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
	"github.com/zulfikarrosadi/code_roast/user"
)

type (
	RepositoryImpl struct {
		*slog.Logger
		*sql.DB
	}

	// these data is populated by system
	authentication struct {
		id           string
		refreshToken string
		lastLogin    int64
		remoteIP     string
		agent        string
		userId       string
	}

	publicUserData struct {
		id       string
		fullname string
		email    string
		roles    []user.Roles
	}
)

const (
	DUPLICATE_CONSTRAINT_ERROR = 1062
)

func NewUserRepository(logger *slog.Logger, db *sql.DB) *RepositoryImpl {
	return &RepositoryImpl{
		Logger: logger,
		DB:     db,
	}
}

func (repo *RepositoryImpl) findRefreshToken(ctx context.Context, token string) (publicUserData, error) {
	newPublicUserData := new(publicUserData)
	tx, err := repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return publicUserData{}, fmt.Errorf("repository: transaction begin error: %w", err)
	}
	defer func() {
		// handle panic in extreamely rare case condition e.g driver fails
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()
	err = tx.QueryRowContext(
		ctx,
		`
		SELECT u.email , u.id, u.fullname
		FROM authentication AS a
		JOIN users AS u
		ON a.user_id = u.id
		WHERE refresh_token = ?
		`,
		token,
	).Scan(&newPublicUserData.email, &newPublicUserData.id, &newPublicUserData.fullname)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return publicUserData{}, errors.New("refresh token not found")
		}
		return publicUserData{}, fmt.Errorf("repository: db query scan failed, %w", err)
	}
	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT r.id as id, r.name as role
		FROM users AS u
		JOIN user_roles AS ur
		ON u.id = ur.user_id
		JOIN roles AS r
		ON ur.role_id = r.id
		WHERE u.id = ?;
		`,
		newPublicUserData.id,
	)
	if err != nil {
		return publicUserData{}, fmt.Errorf("repository: role lookup fail %w", err)
	}
	defer rows.Close()

	userRoles := []user.Roles{}
	for rows.Next() {
		role := user.Roles{}
		rows.Scan(&role.Id, &role.Name)
		userRoles = append(userRoles, role)
	}
	newPublicUserData.roles = userRoles
	err = tx.Commit()
	if err != nil {
		return publicUserData{}, fmt.Errorf("failed to commit transaction %w", err)
	}
	return *newPublicUserData, nil
}

// this method is not only find user by email, but inserting user auth details in db at once
func (repo *RepositoryImpl) loginByEmail(ctx context.Context, email string, auth authentication) (user.User, error) {
	userFromDb := new(user.User)

	tx, err := repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return user.User{}, fmt.Errorf("repository: transaction begin error: %w", err)
	}
	defer func() {
		// handle panic in extreamely rare case condition e.g driver fails
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	err = tx.QueryRowContext(
		ctx,
		"SELECT id, fullname, password, email FROM users WHERE email = ?",
		email,
	).Scan(&userFromDb.Id, &userFromDb.Fullname, &userFromDb.Password, &userFromDb.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// we use apperror to make it easier to directly handle this case
			return user.User{}, apperror.New(http.StatusBadRequest, "email or password is incorrect", err)
		}
		return user.User{}, fmt.Errorf("repository: db query scan failed, %w", err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO authentication (id, refresh_token, last_login, remote_ip, agent, user_id) VALUES(?,?,?,?,?,?)",
		auth.id,
		auth.refreshToken,
		auth.lastLogin,
		auth.remoteIP,
		auth.agent,
		userFromDb.Id,
	)
	if err != nil {
		return user.User{}, fmt.Errorf("repository: insert new user auth credentials failed %w", err)
	}
	rows, err := tx.QueryContext(
		ctx,
		`
			SELECT r.id as id, r.name as role
			FROM users AS u
			JOIN user_roles AS ur
			ON u.id = ur.user_id
			JOIN roles AS r
			ON ur.role_id = r.id
			WHERE u.email = ?;
		`,
		email,
	)
	if err != nil {
		return user.User{}, fmt.Errorf("repository: role lookup fail %w", err)
	}
	defer rows.Close()

	userRoles := []user.Roles{}
	for rows.Next() {
		role := user.Roles{}
		rows.Scan(&role.Id, &role.Name)
		userRoles = append(userRoles, role)
	}
	err = tx.Commit()
	if err != nil {
		return user.User{}, fmt.Errorf("failed to commit transaction %w", err)
	}

	return user.User{
		Id:       userFromDb.Id,
		Fullname: userFromDb.Fullname,
		Email:    userFromDb.Email,
		Password: userFromDb.Password,
		Roles:    userRoles,
	}, nil
}

func (repository *RepositoryImpl) register(ctx context.Context, newUser user.User, auth authentication) (publicUserData, error) {
	tx, err := repository.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return publicUserData{}, fmt.Errorf("repository: transaction begin error: %w", err)
	}
	defer func() {
		// handle panic in extreamely rare case condition e.g driver fails
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO users (id, fullname, email, password, created_at) VALUES (?,?,?,?,?)",
		newUser.Id,
		newUser.Fullname,
		newUser.Email,
		newUser.Password,
		newUser.CreatedAt,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == DUPLICATE_CONSTRAINT_ERROR {
			return publicUserData{}, apperror.New(http.StatusBadRequest, "this email is already registered, please try signin instead", err)
		}
		return publicUserData{}, fmt.Errorf("repository: insert new user fail: %w", err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO authentication (id, refresh_token, last_login, remote_ip, agent, user_id) VALUES(?,?,?,?,?,?)",
		auth.id,
		auth.refreshToken,
		auth.lastLogin,
		auth.remoteIP,
		auth.agent,
		newUser.Id,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == DUPLICATE_CONSTRAINT_ERROR {
			return publicUserData{}, apperror.New(http.StatusBadRequest, "fail to process your request, please insert corrrect information and try again", err)
		}
		return publicUserData{}, fmt.Errorf("repository: insert new user auth credentials failed: %w", err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO user_roles (user_id, role_id) VALUES(?,?)",
		newUser.Id,
		user.ROLE_ID_MEMBER,
	)
	if err != nil {
		return publicUserData{}, fmt.Errorf("repository: attaching new role to new user failed %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return publicUserData{}, fmt.Errorf("repository: failed to commit transaction: %w", err)
	}
	return publicUserData{
		id:       newUser.Id,
		fullname: newUser.Fullname,
		email:    newUser.Email,
		roles: []user.Roles{
			{
				Id:   user.ROLE_ID_MEMBER,
				Name: "member",
			},
		},
	}, nil
}
