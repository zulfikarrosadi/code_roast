package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-sql-driver/mysql"
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

func (repo *RepositoryImpl) FindUserByEmail(ctx context.Context, email string) (User, error) {
	user := new(User)
	err := repo.DB.QueryRowContext(ctx, "SELECT id, fullname, password, email FROM users WHERE email = ?", email).Scan(&user.Id, &user.Fullname, &user.Password, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			repo.Logger.LogAttrs(ctx,
				slog.LevelError,
				"REQUEST_ERROR",
				slog.Group("details",
					slog.String("message", "email not found"),
					slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
				))
			return *user, lib.AuthError{Msg: "email or password is invalid", Code: http.StatusBadRequest}
		}
		repo.Logger.LogAttrs(ctx,
			slog.LevelError,
			"REQUEST_ERROR",
			slog.Group("details",
				slog.String("message", err.Error()),
				slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
			))
		return *user, errors.New("something went wrong, please try again later")
	}
	return *user, nil
}

func (repository *RepositoryImpl) register(ctx context.Context, user userCreateRequest) (User, error) {
	tx, err := repository.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		repository.Logger.LogAttrs(ctx,
			slog.LevelError,
			"REQUEST_ERROR",
			slog.Group("details",
				slog.String("message", err.Error()),
				slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
			))
		return User{}, errors.New("something went wrong, please try again later")
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
		user.Id,
		user.Fullname,
		user.Email,
		user.Password,
		user.CreatedAt,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == DUPLICATE_CONSTRAINT_ERROR {
			repository.Logger.LogAttrs(ctx, slog.LevelError, "duplicate email", slog.Any("details", err))
			return User{}, authError{
				Msg:  "this email is already registered, try login instead",
				Code: http.StatusBadRequest,
			}
		}
		repository.Logger.LogAttrs(ctx, slog.LevelError, "fail insert new user", slog.Any("details", err))
		return User{}, errors.New("something went wrong, please try again later")
	}
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO authentication (id, refresh_token, last_login, remote_ip, agent, user_id) VALUES(?,?,?,?,?,?)",
		user.Authentication.Id,
		user.Authentication.RefreshToken,
		user.Authentication.LastLogin,
		user.Authentication.RemoteIP,
		user.Authentication.Agent, user.Id,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == DUPLICATE_CONSTRAINT_ERROR {
			repository.Logger.LogAttrs(ctx, slog.LevelError, "duplicate refresh token found: possibly stolen", slog.Any("details", err))
			return User{}, authError{
				Msg:  "fail to process your request, please insert corrrect information and try again",
				Code: http.StatusBadRequest,
			}
		}
		repository.Logger.LogAttrs(ctx, slog.LevelError, err.Error(), slog.Any("details", err))
		return User{}, errors.New("something went wrong, please try again later")
	}
	err = tx.Commit()
	if err != nil {
		repository.Logger.LogAttrs(ctx, slog.LevelError, "transaction commit error", slog.Any("details", err))
		return User{}, errors.New("something went wrong, please try again later")
	}
	return User{
		Id:       user.Id,
		Fullname: user.Fullname,
		Email:    user.Email,
	}, nil
}
