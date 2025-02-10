package user

import (
	"context"
	"database/sql"
	"errors"
)

type RepositoryImpl struct {
	*sql.DB
}

func (repository *RepositoryImpl) Create(ctx context.Context, user User) (User, error) {
	result, err := repository.DB.ExecContext(
		ctx,
		"INSERT INTO users (id, fullname, email, password, created_at) VALUES (?,?,?,?,?)",
		user.Id, user.Fullname, user.Email, user.Password, user.CreatedAt,
	)
	if err != nil {
		return user, errors.New("something went wrong, please try again later")
	}
	if rowsAffected, err := result.RowsAffected(); err != nil || rowsAffected == 0 {
		return User{}, errors.New("something went wrong, please try again later")
	}

	return user, nil
}
