package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-sql-driver/mysql"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
)

type authError struct {
	Msg  string
	Code int
}

func (authError authError) Error() string {
	return authError.Msg
}

type RepositoryImpl struct {
	*slog.Logger
	*sql.DB
}

func NewUserRepository(logger *slog.Logger, db *sql.DB) *RepositoryImpl {
	return &RepositoryImpl{
		Logger: logger,
		DB:     db,
	}
}

const (
	DUPLICATE_CONSTRAINT_ERROR = 1062
)

const (
	ROLE_ID_CREATE_SUBFORUM = 1
	ROLE_ID_UPDATE_SUBFORUM = 2
	ROLE_ID_DELETE_SUBFORUM = 3
	ROLE_ID_MEMBER          = 4
	ROLE_ID_DELETE_POST     = 5
	ROLE_ID_APPROVE_POST    = 6
	ROLE_ID_TAKE_DOWN_POST  = 7
)

type roles struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type publicUserData struct {
	id       string
	fullname string
	email    string
	roles    []roles
}

func (repo *RepositoryImpl) findRefreshToken(ctx context.Context, token string) (publicUserData, error) {
	user := new(publicUserData)
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
	).Scan(&user.email, &user.id, &user.fullname)
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
		user.id,
	)
	if err != nil {
		return publicUserData{}, fmt.Errorf("repository: role lookup fail %w", err)
	}
	defer rows.Close()

	userRoles := []roles{}
	for rows.Next() {
		role := roles{}
		rows.Scan(&role.Id, &role.Name)
		userRoles = append(userRoles, role)
	}
	user.roles = userRoles
	err = tx.Commit()
	if err != nil {
		return publicUserData{}, fmt.Errorf("failed to commit transaction %w", err)
	}
	return *user, nil
}

func (repo *RepositoryImpl) findByEmail(ctx context.Context, email string, auth authentication) (User, error) {
	user := new(User)

	tx, err := repo.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return User{}, fmt.Errorf("repository: transaction begin error: %w", err)
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
	).Scan(&user.Id, &user.Fullname, &user.Password, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// we use authError to make it easier to directly handle this case
			return User{}, authError{Msg: "email or password is invalid", Code: http.StatusBadRequest}
		}
		return User{}, fmt.Errorf("repository: db query scan failed, %w", err)
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO authentication (id, refresh_token, last_login, remote_ip, agent, user_id) VALUES(?,?,?,?,?,?)",
		auth.id,
		auth.refreshToken,
		auth.lastLogin,
		auth.remoteIP,
		auth.agent,
		user.Id,
	)
	if err != nil {
		return User{}, fmt.Errorf("repository: insert new user auth credentials failed %w", err)
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
		return User{}, fmt.Errorf("repository: role lookup fail %w", err)
	}
	defer rows.Close()

	userRoles := []roles{}
	for rows.Next() {
		role := roles{}
		rows.Scan(&role.Id, &role.Name)
		userRoles = append(userRoles, role)
	}
	err = tx.Commit()
	if err != nil {
		return User{}, fmt.Errorf("failed to commit transaction %w", err)
	}

	return User{
		Id:       user.Id,
		Fullname: user.Fullname,
		Email:    user.Email,
		Password: user.Password,
		Roles:    userRoles,
	}, nil
}

type userAndAuth struct {
	id        string
	fullname  string
	email     string
	password  string
	createdAt int64
	authentication
}

func (repository *RepositoryImpl) register(ctx context.Context, user userAndAuth) (publicUserData, error) {
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
		user.id,
		user.fullname,
		user.email,
		user.password,
		user.createdAt,
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
		user.authentication.id,
		user.authentication.refreshToken,
		user.authentication.lastLogin,
		user.authentication.remoteIP,
		user.authentication.agent,
		user.id,
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
		user.id,
		ROLE_ID_MEMBER,
	)
	if err != nil {
		return publicUserData{}, fmt.Errorf("repository: attaching new role to new user failed %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return publicUserData{}, fmt.Errorf("repository: failed to commit transaction: %w", err)
	}
	return publicUserData{
		id:       user.id,
		fullname: user.fullname,
		email:    user.email,
		roles: []roles{
			roles{
				Id:   ROLE_ID_MEMBER,
				Name: "member",
			},
		},
	}, nil
}
