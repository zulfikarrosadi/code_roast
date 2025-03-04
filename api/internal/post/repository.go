package post

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	apperror "github.com/zulfikarrosadi/code_roast/internal/app-error"
	"github.com/zulfikarrosadi/code_roast/internal/subforum"
	"github.com/zulfikarrosadi/code_roast/internal/user"
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
	id         string
	caption    string
	createdAt  int64
	updatedAt  sql.NullInt64
	postMedia  []postMedia
	userId     string
	subforumId string
}

type newPost struct {
	id        string
	caption   string
	mediaUrl  []string
	createdAt int64
	updatedAt sql.NullInt64
	user      user.User
	subforum  subforum.Subforum
}

type createPostResult struct {
	post newPost
}

const (
	POST_STATUS_PUBLISHED = "published"
	POST_STATUS_PENDING   = "pending"
	POST_STATUS_TAKE_DOWN = "take_down"
)

func (repo *RepositoryImpl) create(
	ctx context.Context,
	data post,
) (createPostResult, error) {
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
		return createPostResult{}, fmt.Errorf("repository: failed to begin transaction %w", err)
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
		"INSERT INTO posts (id, caption, created_at, user_id, subforum_id) VALUES (?,?,?,?,?)",
		data.id, data.caption, data.createdAt, data.userId, data.subforumId,
	)
	if err != nil {
		return createPostResult{}, fmt.Errorf("repository: fail to create new posts %w", err)
	}
	if len(data.postMedia) > 0 {
		_, err = tx.ExecContext(
			ctx,
			insertPostMediaQuery,
			postMediaArgs...,
		)
		if err != nil {
			return createPostResult{}, fmt.Errorf("repository: fail to add post media %w", err)
		}
	}

	rows, err := tx.QueryContext(
		ctx,
		`
		SELECT p.id, p.caption, p.updated_at, pm.media_url, u.id AS user_id, u.fullname, sf.id AS subforum_id, sf.name AS subforum_name
		FROM posts p
		JOIN users u
		ON p.user_id = u.id
		JOIN subforums sf
		ON p.subforum_id = sf.id
		LEFT JOIN post_media pm
		ON pm.post_id = p.id
		WHERE p.id = ?`,
		data.id,
	)
	if err != nil {
		return createPostResult{}, err
	}
	defer rows.Close()

	np := &newPost{}
	var mediaURLs []string

	for rows.Next() {
		var mediaURL sql.NullString
		if err := rows.Scan(
			&np.id,
			&np.caption,
			&np.updatedAt,
			&mediaURL,
			&np.user.Id,
			&np.user.Fullname,
			&np.subforum.Id,
			&np.subforum.Name,
		); err != nil {
			return createPostResult{}, err
		}
		if mediaURL.Valid {
			mediaURLs = append(mediaURLs, mediaURL.String)
		}
	}
	np.mediaUrl = mediaURLs

	err = tx.Commit()
	if err != nil {
		return createPostResult{}, fmt.Errorf("repository: fail to create new post. transaction fail to commit %w", err)
	}

	return createPostResult{
		post: newPost{
			id:        data.id,
			caption:   data.caption,
			mediaUrl:  np.mediaUrl,
			createdAt: data.createdAt,
			user: user.User{
				Id:       np.user.Id,
				Fullname: np.user.Fullname,
			},
			subforum: subforum.Subforum{
				Id:   np.subforum.Id,
				Name: np.subforum.Name,
			},
		},
	}, nil
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

type newLike struct {
	userId    string
	postId    string
	createdAt int64
}

func (repo *RepositoryImpl) like(
	ctx context.Context,
	data newLike,
) (int, error) {
	tx, err := repo.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("repository: failed to begin transaction %w", err)
	}
	defer func() {
		// handle panic for extream case like driver fails
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	rows, err := tx.ExecContext(
		ctx,
		"INSERT INTO likes (post_id, user_id, created_at)  VALUES (?,?,?)",
		data.postId,
		data.userId,
		data.createdAt,
	)
	if err != nil {
		return 0, fmt.Errorf("repository: failed to add new like to post %w", err)
	}
	rowsAffected, err := rows.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return 0, fmt.Errorf("repository: failed to add new like (rows affected 0) %w", err)
	}
	type newLike struct {
		count int
	}
	likeCount := &newLike{}
	err = tx.QueryRowContext(
		ctx,
		"SELECT COUNT(post_id) FROM likes WHERE post_id = ?",
		data.postId,
	).Scan(&likeCount.count)
	if err != nil {
		return 0, fmt.Errorf("repository: failed to get likes count, %w", err)
	}
	fmt.Println(*likeCount)
	return likeCount.count, nil
}
