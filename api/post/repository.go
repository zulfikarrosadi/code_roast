package post

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	apperror "github.com/zulfikarrosadi/code_roast/app-error"
)

type RepositoryImpl struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *RepositoryImpl {
	return &RepositoryImpl{
		DB: db,
	}
}

type postMedia struct {
	Id        string `json:"id"`
	MediaUrl  string `json:"media_url"`
	PostId    string `json:"post_id"`
	CreatedAt int64  `json:"created_at"`
}

type post struct {
	id        string
	caption   string
	createdAt int64
	updatedAt sql.NullInt64
	postMedia []postMedia
	userId    string
}

const (
	POST_STATUS_PUBLISHED = "published"
	POST_STATUS_PENDING   = "pending"
	POST_STATUS_TAKE_DOWN = "take_down"
)

func (repo *RepositoryImpl) create(ctx context.Context, data post) (post, error) {
	postMediaValue := []string{}
	postMediaArgs := []interface{}{}
	var insertPostMediaQuery string

	if len(data.postMedia) > 0 {
		for _, media := range data.postMedia {
			postMediaValue = append(postMediaValue, "(?,?,?,?)")
			postMediaArgs = append(postMediaArgs, media.Id, media.MediaUrl, data.id, media.CreatedAt)
		}
		insertPostMediaQuery = fmt.Sprintf("INSERT INTO post_media (id, media_url, post_id, created_at) VALUES %s", strings.Join(postMediaValue, ","))
	}

	tx, err := repo.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return post{}, fmt.Errorf("repository: failed to begin transaction %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO posts (id, caption, created_at, user_id) VALUES (?,?,?,?)",
		data.id, data.caption, data.createdAt, data.userId,
	)
	if err != nil {
		return post{}, fmt.Errorf("repository: fail to create new posts %w", err)
	}
	if len(data.postMedia) > 0 {
		_, err = tx.ExecContext(
			ctx,
			insertPostMediaQuery,
			postMediaArgs...,
		)
		if err != nil {
			return post{}, fmt.Errorf("repository: fail to add post media %w", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return post{}, fmt.Errorf("repository: fail to create new post. transaction fail to commit %w", err)
	}
	return data, nil
}

func (repo *RepositoryImpl) takeDown(ctx context.Context, postId string, updatedAt sql.NullInt64) error {
	result, err := repo.DB.ExecContext(
		ctx,
		"UPDATE posts SET status = ?, updated_at = ? WHERE id = ?",
		POST_STATUS_TAKE_DOWN, updatedAt.Int64, postId,
	)
	if err != nil {
		return fmt.Errorf("repository: failed to take down post %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository: failed to get rows affected")
	}
	if rowsAffected == 0 {
		return apperror.New(http.StatusBadRequest, "failed to take down post, post id not found", err)
	}
	return nil
}
