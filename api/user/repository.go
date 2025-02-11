package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
)

type RepositoryImpl struct {
	*slog.Logger
	*sql.DB
}

func (repository *RepositoryImpl) Create(ctx context.Context, user User) (User, error) {
	result, err := repository.DB.ExecContext(
		ctx,
		"INSERT INTO users (id, fullname, email, password, created_at) VALUES (?,?,?,?,?)",
		user.Id, user.Fullname, user.Email, user.Password, user.CreatedAt,
	)
	if err != nil {
		repository.Logger.LogAttrs(ctx, slog.LevelError, "fail insert new user", slog.Any("details", err))
		return user, errors.New("something went wrong, please try again later")
	}
	if rowsAffected, err := result.RowsAffected(); err != nil || rowsAffected == 0 {
		repository.Logger.LogAttrs(ctx, slog.LevelError, "fail get rows affected", slog.Any("details", err))
		return User{}, errors.New("something went wrong, please try again later")
	}

	return user, nil
}
