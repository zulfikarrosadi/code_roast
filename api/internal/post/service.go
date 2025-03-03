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
	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type repository interface {
	create(context.Context, post) (post, error)
	takeDown(context.Context, string, sql.NullInt64) error
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
	userId    string
	Caption   string                  `validate:"required"`
	PostMedia []*multipart.FileHeader `validate:"required,min=1,max=10"`
}

type postCreateResponse struct {
	Id        string        `json:"id"`
	Caption   string        `json:"caption"`
	PostMedia []postMedia   `json:"post_media"`
	CreatedAt int64         `json:"created_at"`
	UpdatedAt sql.NullInt64 `json:"updated_at"`
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
	for _, item := range data.PostMedia {
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

	newPost, err := service.repo.create(ctx, post{
		id:        postId.String(),
		caption:   data.Caption,
		createdAt: time.Now().Unix(),
		updatedAt: sql.NullInt64{},
		userId:    data.userId,
		postMedia: media,
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
				Id:        newPost.id,
				Caption:   newPost.caption,
				PostMedia: newPost.postMedia,
				CreatedAt: newPost.createdAt,
				UpdatedAt: newPost.updatedAt,
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
				UpdatedAt: updatedAt,
			},
		},
	}, nil
}
