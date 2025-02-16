package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-sql-driver/mysql"
	"github.com/zulfikarrosadi/code_roast/lib"
)

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

func (repository *RepositoryImpl) Create(ctx context.Context, user User) (User, error) {
	result, err := repository.DB.ExecContext(
		ctx,
		"INSERT INTO users (id, fullname, email, password, created_at) VALUES (?,?,?,?,?)",
		user.Id, user.Fullname, user.Email, user.Password, user.CreatedAt,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == DUPLICATE_CONSTRAINT_ERROR {
			repository.Logger.LogAttrs(ctx, slog.LevelError, "duplicate email", slog.Any("details", err))
			return user, lib.AuthError{
				Msg:  "this email is already registered, try login instead",
				Code: http.StatusBadRequest,
			}
		}
		repository.Logger.LogAttrs(ctx, slog.LevelError, "fail insert new user", slog.Any("details", err))
		return user, errors.New("something went wrong, please try again later")
	}
	if rowsAffected, err := result.RowsAffected(); err != nil || rowsAffected == 0 {
		repository.Logger.LogAttrs(ctx, slog.LevelError, "fail get rows affected", slog.Any("details", err))
		return User{}, errors.New("something went wrong, please try again later")
	}

	return user, nil
}
