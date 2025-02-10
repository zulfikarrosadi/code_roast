package user

import (
	"context"
	"database/sql"

	"github.com/zulfikarrosadi/code_roast/lib"
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
		return user, lib.ServerError{Msg: "something went wrong, please try again later"}
	}
	if rowsAffected, err := result.RowsAffected(); err != nil || rowsAffected == 0 {
		return User{}, lib.ServerError{Msg: "something went wrong, please try again later"}
	}

	return user, nil
}
