package moderator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-sql-driver/mysql"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
)

const (
	DUPLICATE_CONSTRAINT_ERROR = 1062
)

type repositoryImpl struct {
	db *sql.DB
}

type userAndRole struct {
	userId string
	roles  []roles
}
type roles struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

func (repo *repositoryImpl) addRoles(
	ctx context.Context,
	userId string, roleId []int,
) (userAndRole, error) {
	insertQueryParams := []string{}
	insertQueryValue := []interface{}{}

	for _, role := range roleId {
		insertQueryParams = append(insertQueryParams, "(?,?)")
		insertQueryValue = append(insertQueryValue, userId, role)
	}
	tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return userAndRole{}, fmt.Errorf("repository: failed to begin transaction %w", err)
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

	result, err := tx.ExecContext(
		ctx,
		fmt.Sprintf("INSERT INTO user_roles (user_id, role_id) VALUES %v", strings.Join(insertQueryParams, ",")),
		insertQueryValue...,
	)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == DUPLICATE_CONSTRAINT_ERROR {
			return userAndRole{}, apperror.New(http.StatusBadRequest, "This user already have that role", err)
		}
		return userAndRole{}, fmt.Errorf("repository: fail to add new role to user %s %w", userId, err)
	}
	rowsAffectd, err := result.RowsAffected()
	if err != nil {
		return userAndRole{}, fmt.Errorf("repository: fail to get rows affected %w", err)
	}
	if rowsAffectd == 0 {
		return userAndRole{}, fmt.Errorf("repository: add new role failed, 0 rows affected %w", err)
	}
	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT r.id, r.name
		FROM user_roles ur
		JOIN roles r
		ON ur.role_id = r.id
		WHERE ur.user_id = ?
		`,
		userId,
	)
	userRoles := []roles{}
	defer rows.Close()
	for rows.Next() {
		userRole := roles{}
		err = rows.Scan(&userRole.Id, &userRole.Name)
		if err != nil {
			return userAndRole{}, fmt.Errorf("repository: fail to scan user role %w", err)
		}
		userRoles = append(userRoles, userRole)
	}

	if err != nil {
		return userAndRole{}, fmt.Errorf("repository: get user role failed %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return userAndRole{}, fmt.Errorf("repository: failed to commit transaction %w", err)
	}

	return userAndRole{
		userId: userId,
		roles:  userRoles,
	}, nil
}

// batch query to delete user role consecutively with transaction
func (repo *repositoryImpl) removeRoles(
	ctx context.Context,
	userId string,
	roleId []int,
) (userAndRole, error) {
	type deleteQuery struct {
		query string
		value []interface{}
	}
	deleteQueries := []deleteQuery{}

	for _, role := range roleId {
		dq := deleteQuery{
			query: "DELETE FROM user_roles WHERE user_id = ? AND role_id = ?",
			value: []interface{}{userId, role},
		}
		deleteQueries = append(deleteQueries, dq)
	}

	tx, err := repo.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return userAndRole{}, fmt.Errorf("repository: failed to begin transaction %w", err)
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

	for _, query := range deleteQueries {
		result, err := tx.ExecContext(
			ctx,
			query.query,
			query.value...,
		)
		if err != nil {
			return userAndRole{}, fmt.Errorf("repository: failed to remove roles %w", err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return userAndRole{}, fmt.Errorf("repository: failed to get rows affected %w", err)
		}
		if rowsAffected == 0 {
			return userAndRole{}, apperror.New(http.StatusBadRequest, "remove roles failed, enter correct user and role data and try again", err)
		}
	}

	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT ur.id, ur.name
		FROM user_roles ur
		JOIN roles r
		ON ur.role_id = r.id
		WHERE ur.user_id = ?
		`,
		userId,
	)
	userRoles := []roles{}
	defer rows.Close()
	for rows.Next() {
		userRole := roles{}
		err = rows.Scan(&userRole.Id, &userRole.Name)
		if err != nil {
			return userAndRole{}, fmt.Errorf("repository: fail to scan user role %w", err)
		}
		userRoles = append(userRoles, userRole)
	}

	if err != nil {
		return userAndRole{}, fmt.Errorf("repository: get user role failed %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return userAndRole{}, fmt.Errorf("failed to commit transaction %w", err)
	}

	return userAndRole{}, nil
}
