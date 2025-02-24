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

type publicUserData struct {
	id       string
	fullname string
	email    string
}

func (repo *RepositoryImpl) findByEmail(ctx context.Context, email string) (User, error) {
	user := new(User)
	err := repo.DB.QueryRowContext(
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
	return User{
		Id:       user.Id,
		Fullname: user.Fullname,
		Email:    user.Email,
		Password: user.Password,
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
	err = tx.Commit()
	if err != nil {
		return publicUserData{}, fmt.Errorf("repository: failed to commit transaction: %w", err)
	}
	return publicUserData{
		id:       user.id,
		fullname: user.fullname,
		email:    user.email,
	}, nil
}
