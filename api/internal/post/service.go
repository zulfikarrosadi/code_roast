package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	apperror "github.com/zulfikarrosadi/code_roast/internal/app-error"
	imagehelper "github.com/zulfikarrosadi/code_roast/internal/image-helper"
	"github.com/zulfikarrosadi/code_roast/internal/subforum"
	"github.com/zulfikarrosadi/code_roast/internal/user"
	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type repository interface {
	create(context.Context, post) (newPost, error)
	takeDown(context.Context, string, sql.NullInt64) error
	like(context.Context, newLike) (int, error)
	findAll(context.Context) ([]getAllPost, error)
}

type serviceImpl struct {
	repo repository
	v    *validator.Validate
	cld  *cloudinary.Cloudinary
}

func NewService(repo repository, v *validator.Validate, cld *cloudinary.Cloudinary) *serviceImpl {
	return &serviceImpl{
		repo: repo,
		v:    v,
		cld:  cld,
	}
}

type postDTO struct {
	Id        string            `json:"id"`
	Caption   string            `json:"caption"`
	Media     []string          `json:"media"`
	CreatedAt int64             `json:"created_at"`
	UpdatedAt int64             `json:"updated_at"`
	Subforum  subforum.Subforum `json:"subforum"`
	User      user.User         `json:"user"`
}
type createRequest struct {
	userId     string
	Caption    string                  `validate:"required"`
	SubforumId string                  `validate:"required"`
	Media      []*multipart.FileHeader `validate:"required,min=1,max=10"`
}

type createResponse struct {
	Post postDTO `json:"post"`
}

type getAllResponse struct {
	Posts []postDTO `json:"posts"`
}

func (service *serviceImpl) create(ctx context.Context, data createRequest) (schema.Response[createResponse], error) {
	err := service.v.Struct(data)
	if err != nil {
		validationError := apperror.HandlerValidatorError(err.(validator.ValidationErrors))
		return schema.Response[createResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: apperror.VALIDATION_ERROR,
				Details: validationError,
			},
		}, fmt.Errorf("service: create post validation error %w", err)
	}
	postId, err := uuid.NewV7()
	if err != nil {
		return schema.Response[createResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "something went wrong, please try again later",
			},
		}, fmt.Errorf("service: fail to generate post uuid %w", err)
	}

	var media []postMedia
	for _, item := range data.Media {
		mediaId, err := uuid.NewV7()
		if err != nil {
			return schema.Response[createResponse]{
				Status: "fail",
				Code:   http.StatusInternalServerError,
				Error: schema.Error{
					Message: "something went wrong, please try again later",
				},
			}, fmt.Errorf("service: fail to generate post media uuid %w", err)
		}
		postMediaSrc, err := item.Open()
		if err != nil {
			return schema.Response[createResponse]{
				Status: "fail",
				Code:   http.StatusInternalServerError,
				Error: schema.Error{
					Message: "fail to create new post, failed to open media file",
				},
			}, fmt.Errorf("service: failed to open media file %w", err)
		}
		defer postMediaSrc.Close()
		if _, err := imagehelper.IsImage(postMediaSrc); err != nil {
			return schema.Response[createResponse]{
				Status: "fail",
				Code:   http.StatusBadRequest,
				Error: schema.Error{
					Message: "fail to create new post, unsupported media file type. Only upload jpg or png file",
				},
			}, fmt.Errorf("media is not image %w", err)
		}
		mediaUpload, err := service.cld.Upload.Upload(
			ctx,
			postMediaSrc,
			uploader.UploadParams{
				ResourceType: "image",
			},
		)
		if err != nil {
			return schema.Response[createResponse]{
				Status: "fail",
				Code:   http.StatusInternalServerError,
				Error: schema.Error{
					Message: "fail to create new post, failed to upload media file",
				},
			}, fmt.Errorf("service: failed to upload media file %w", err)
		}

		media = append(media, postMedia{
			Id:       mediaId.String(),
			MediaUrl: mediaUpload.SecureURL,
		})
	}

	result, err := service.repo.create(ctx, post{
		id:         postId.String(),
		caption:    data.Caption,
		createdAt:  time.Now().Unix(),
		updatedAt:  sql.NullInt64{},
		postMedia:  media,
		userId:     data.userId,
		subforumId: data.SubforumId,
	})
	if err != nil {
		return schema.Response[createResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "failed to upload new post, enter correct information and try again",
			},
		}, err
	}
	return schema.Response[createResponse]{
		Status: "success",
		Code:   http.StatusCreated,
		Data: createResponse{
			Post: postDTO{
				Id:        result.id,
				Caption:   result.caption,
				Media:     result.mediaUrl,
				CreatedAt: result.createdAt,
				UpdatedAt: result.updatedAt.Int64,
				Subforum: subforum.Subforum{
					Id:   result.subforum.Id,
					Name: result.subforum.Name,
				},
				User: user.User{
					Id:       result.user.Id,
					Fullname: result.user.Fullname,
				},
			},
		},
	}, nil
}

func (service *serviceImpl) takeDown(ctx context.Context, postId string, updatedAt sql.NullInt64) (schema.Response[createResponse], error) {
	err := service.repo.takeDown(ctx, postId, updatedAt)
	if err != nil {
		var appError *apperror.AppError
		if errors.As(err, &appError) {
			return schema.Response[createResponse]{
				Status: "fail",
				Code:   appError.Code,
				Error: schema.Error{
					Message: appError.Message,
				},
			}, err
		}
		return schema.Response[createResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "something went wrong, please try again later",
			},
		}, err
	}
	return schema.Response[createResponse]{
		Status: "success",
		Code:   http.StatusOK,
		Data: createResponse{
			Post: postDTO{
				Id:        postId,
				UpdatedAt: updatedAt.Int64,
			},
		},
	}, nil
}

type likeCreateResponse struct {
	PostId    string `json:"id"`
	LikeCount int    `json:"like_count"`
}

type likeCreateRequest struct {
	UserId string
	PostId string `param:"id" validate:"required"`
}

type likeResponse struct {
	Post likeCreateResponse `json:"post"`
}

func (service *serviceImpl) like(
	ctx context.Context,
	data likeCreateRequest,
) (schema.Response[likeResponse], error) {
	err := service.v.Struct(data)
	if err != nil {
		validationError := apperror.HandlerValidatorError(err.(validator.ValidationErrors))
		return schema.Response[likeResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: apperror.VALIDATION_ERROR,
				Details: validationError,
			},
		}, fmt.Errorf("service: error input")
	}
	likeCount, err := service.repo.like(ctx, newLike{
		userId:    data.UserId,
		postId:    data.PostId,
		createdAt: time.Now().Unix(),
	})
	if err != nil {
		return schema.Response[likeResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "failed to like this post, please try again later",
			},
		}, err
	}
	return schema.Response[likeResponse]{
		Status: "success",
		Code:   http.StatusCreated,
		Data: likeResponse{
			Post: likeCreateResponse{
				PostId:    data.PostId,
				LikeCount: likeCount,
			},
		},
	}, nil
}

func (service *serviceImpl) getAll(ctx context.Context) (schema.Response[getAllResponse], error) {
	result, err := service.repo.findAll(ctx)
	if err != nil {
		return schema.Response[getAllResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "something went wrong, please try again later",
			},
		}, err
	}
	posts := []postDTO{}
	for _, v := range result {
		p := postDTO{
			Id:        v.id,
			Caption:   v.caption,
			Media:     v.mediaUrl,
			CreatedAt: v.createdAt,
			Subforum:  v.subforum,
			User:      v.user,
		}
		if v.updatedAt.Valid {
			p.UpdatedAt = v.updatedAt.Int64
		}
		posts = append(posts, p)
	}
	return schema.Response[getAllResponse]{
		Status: "success",
		Code:   http.StatusOK,
		Data: getAllResponse{
			Posts: posts,
		},
	}, nil
}
