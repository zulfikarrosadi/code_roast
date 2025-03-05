package subforum

import (
	"context"
	"database/sql"
	"fmt"
)

type RepositoryImpl struct {
	DB *sql.DB
}

type Subforum struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserId      string `json:"user_id"`
	CreatedAt   int64  `json:"created_at"`
	Icon        string `json:"icon"`
	Banner      string `json:"banner"`
}

func NewRepository(db *sql.DB) *RepositoryImpl {
	return &RepositoryImpl{
		DB: db,
	}
}

func (repo *RepositoryImpl) create(ctx context.Context, data Subforum) (Subforum, error) {
	_, err := repo.DB.ExecContext(
		ctx,
		"INSERT INTO subforums (id, name, description, user_id, icon, banner, created_at) VALUES (?,?,?,?,?,?,?)",
		data.Id,
		data.Name,
		data.Description,
		data.UserId,
		data.Icon,
		data.Banner,
		data.CreatedAt,
	)
	if err != nil {
		return Subforum{}, fmt.Errorf("repository: fail to create new subforum %w", err)
	}

	return data, nil
}

func (repo *RepositoryImpl) findByName(ctx context.Context, name string) ([]Subforum, error) {
	subforums := []Subforum{}
	rows, err := repo.DB.QueryContext(
		ctx,
		"SELECT id, name, description, user_id, icon, created_at FROM subforums WHERE name = ?",
		name,
	)
	if err != nil {
		return []Subforum{}, fmt.Errorf("repository: fail to retrieve subforums by name %w", err)
	}
	for rows.Next() {
		sf := Subforum{}
		err = rows.Scan(&sf.Id, &sf.Name, &sf.Description, &sf.UserId, &sf.Icon, &sf.CreatedAt)
		if err != nil {
			return []Subforum{}, fmt.Errorf("repository: fail to scan subforums %w", err)
		}
		subforums = append(subforums, sf)
	}
	return subforums, nil
}

func (repo *RepositoryImpl) deleteById(ctx context.Context, id string, userId string) error {
	tx, err := repo.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("repository: transaction begin error %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(
		ctx,
		"DELETE FROM subforums WHERE id = ? AND user_id = ?",
		id,
		userId,
	)
	if err != nil {
		return fmt.Errorf("repository: fail to delete subforums (forumId: %s, userId: %s) %w", id, userId, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: fail to get rows affected (forumId: %s, userId: %s) %w", id, userId, err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("repository: rows affected is not 1 (forumId: %s, userId: %s)  %w", id, userId, err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("repository: failed to commit transaction deleting subforum (forumId: %s, userId: %s) %w", id, userId, err)
	}

	return nil
}

func (repo *RepositoryImpl) findById(ctx context.Context, forumId string) (Subforum, error) {
	sf := &Subforum{}
	err := repo.DB.QueryRowContext(
		ctx,
		"SELECT id, name, description, icon, banner, created_at FROM subforums WHERE id = ?",
		forumId,
	).Scan(&sf.Id, &sf.Name, &sf.Description, &sf.Icon, &sf.Banner, &sf.CreatedAt)
	if err != nil {
		return Subforum{}, fmt.Errorf("repository: failed to get subforum by id %w", err)
	}
	return *sf, nil
}
