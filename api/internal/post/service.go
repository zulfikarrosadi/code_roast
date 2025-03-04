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
	create(context.Context, post) (createPostResult, error)
	takeDown(context.Context, string, sql.NullInt64) error
	like(context.Context, newLike) (int, error)
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

type postCreateRequest struct {
	userId     string
	Caption    string                  `validate:"required"`
	SubforumId string                  `validate:"required"`
	Media      []*multipart.FileHeader `validate:"required,min=1,max=10"`
}

type postCreateResponse struct {
	Id        string            `json:"id"`
	Caption   string            `json:"caption"`
	Media     []string          `json:"media"`
	CreatedAt int64             `json:"created_at"`
	UpdatedAt int64             `json:"updated_at"`
	Subforum  subforum.Subforum `json:"subforum"`
	User      user.User         `json:"user"`
}

type postResponse struct {
	Post postCreateResponse `json:"post"`
}

func (service *serviceImpl) create(ctx context.Context, data postCreateRequest) (schema.Response[postResponse], error) {
	err := service.v.Struct(data)
	if err != nil {
		validationError := apperror.HandlerValidatorError(err.(validator.ValidationErrors))
		return schema.Response[postResponse]{
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
		return schema.Response[postResponse]{
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
			return schema.Response[postResponse]{
				Status: "fail",
				Code:   http.StatusInternalServerError,
				Error: schema.Error{
					Message: "something went wrong, please try again later",
				},
			}, fmt.Errorf("service: fail to generate post media uuid %w", err)
		}
		postMediaSrc, err := item.Open()
		if err != nil {
			return schema.Response[postResponse]{
				Status: "fail",
				Code:   http.StatusInternalServerError,
				Error: schema.Error{
					Message: "fail to create new post, failed to open media file",
				},
			}, fmt.Errorf("service: failed to open media file %w", err)
		}
		defer postMediaSrc.Close()
		if _, err := imagehelper.IsImage(postMediaSrc); err != nil {
			return schema.Response[postResponse]{
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
			return schema.Response[postResponse]{
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
		return schema.Response[postResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "failed to upload new post, enter correct information and try again",
			},
		}, err
	}
	return schema.Response[postResponse]{
		Status: "success",
		Code:   http.StatusCreated,
		Data: postResponse{
			Post: postCreateResponse{
				Id:        result.post.id,
				Caption:   result.post.caption,
				Media:     result.post.mediaUrl,
				CreatedAt: result.post.createdAt,
				UpdatedAt: result.post.updatedAt.Int64,
				Subforum: subforum.Subforum{
					Id:   result.post.subforum.Id,
					Name: result.post.subforum.Name,
				},
				User: user.User{
					Id:       result.post.user.Id,
					Fullname: result.post.user.Fullname,
				},
			},
		},
	}, nil
}

func (service *serviceImpl) takeDown(ctx context.Context, postId string, updatedAt sql.NullInt64) (schema.Response[postResponse], error) {
	err := service.repo.takeDown(ctx, postId, updatedAt)
	if err != nil {
		var appError apperror.AppError
		if errors.As(err, &appError) {
			return schema.Response[postResponse]{
				Status: "fail",
				Code:   appError.Code,
				Error: schema.Error{
					Message: appError.Message,
				},
			}, err
		}
		return schema.Response[postResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "something went wrong, please try again later",
			},
		}, err
	}
	return schema.Response[postResponse]{
		Status: "success",
		Code:   http.StatusOK,
		Data: postResponse{
			Post: postCreateResponse{
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
