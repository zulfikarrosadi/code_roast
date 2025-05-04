package user

import (
	"context"
	"database/sql"
	"errors"
)

type repositoryImpl struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) repositoryImpl {
	return repositoryImpl{
		DB: db,
	}
}

type UserNotFound struct {
	Message string
}

func (unf *UserNotFound) Error() string {
	return unf.Message
}

func (repo repositoryImpl) findById(ctx context.Context, id string) (User, error) {
	row := repo.DB.QueryRowContext(ctx, "SELECT id, fullname, email FROM users WHERE id = ?", id)
	user := User{}

	err := row.Scan(&user.Id, &user.Fullname, &user.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, &UserNotFound{Message: "User not found"}
	} else if err != nil {
		return User{}, err
	}

	return user, nil
}
