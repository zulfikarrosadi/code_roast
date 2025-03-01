package subforum

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
	imagehelper "github.com/zulfikarrosadi/code_roast/image-helper"
	"github.com/zulfikarrosadi/code_roast/schema"
)

type repository interface {
	create(context.Context, subforum) (subforum, error)
	findByName(context.Context, string) ([]subforum, error)
	deleteById(context.Context, string, string) error
}

type ServiceImpl struct {
	repo repository
	v    *validator.Validate
	cld  *cloudinary.Cloudinary
}

type subforumResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserId      string `json:"user_id"`
	CreatedAt   int64  `json:"created_at"`
}

func NewService(repo repository, v *validator.Validate, cloudinaryInstance *cloudinary.Cloudinary) *ServiceImpl {
	return &ServiceImpl{
		repo: repo,
		v:    v,
		cld:  cloudinaryInstance,
	}
}

func (service *ServiceImpl) create(ctx context.Context, data subforumCreateRequest) (schema.Response[subforumResponse], error) {
	err := service.v.Struct(data)
	if err != nil {
		validatorError := apperror.HandlerValidatorError(err.(validator.ValidationErrors))
		return schema.Response[subforumResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: apperror.VALIDATION_ERROR,
				Details: validatorError,
			},
		}, fmt.Errorf("service: create subforum validation error %w", err)
	}

	iconSrc, err := data.Icon.Open()
	if err != nil {
		return schema.Response[subforumResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create new subforum, failed to open icon file",
			},
		}, fmt.Errorf("service: failed to open icon file %w", err)
	}
	defer iconSrc.Close()
	if _, err := imagehelper.IsImage(iconSrc); err != nil {
		return schema.Response[subforumResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: "fail to create new subforum, unsupported icon file type. Only upload jpg or png file",
			},
		}, fmt.Errorf("service: icon not image %w", err)
	}

	subForumIconUpload, err := service.cld.Upload.Upload(
		ctx,
		iconSrc,
		uploader.UploadParams{
			ResourceType: "image",
		},
	)
	if err != nil {
		return schema.Response[subforumResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create new subforum, failed to upload icon file",
			},
		}, fmt.Errorf("service: failed to upload icon file %w", err)
	}
	bannerSrc, err := data.Banner.Open()
	if err != nil {
		return schema.Response[subforumResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create new subforum, failed to open banner file",
			},
		}, fmt.Errorf("service: failed to open banner file %w", err)
	}
	defer bannerSrc.Close()
	if _, err := imagehelper.IsImage(bannerSrc); err != nil {
		return schema.Response[subforumResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: "fail to create new subforum, unsupported banner file type. Only upload jpg or png file",
			},
		}, fmt.Errorf("service: banner not image %w", err)
	}

	subForumBannerUpload, err := service.cld.Upload.Upload(
		ctx,
		bannerSrc,
		uploader.UploadParams{
			ResourceType: "image",
		},
	)
	if err != nil {
		return schema.Response[subforumResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create new subforum, failed to upload banner file",
			},
		}, fmt.Errorf("service: failed to upload banner file %w", err)
	}
	iconSecureUrl := subForumIconUpload.SecureURL
	bannerSecureUrl := subForumBannerUpload.SecureURL

	subForumId, err := uuid.NewV7()
	if err != nil {
		return schema.Response[subforumResponse]{}, fmt.Errorf("service: fail to generate subforum uuid v7 %w", err)
	}

	result, err := service.repo.create(ctx, subforum{
		id:          subForumId.String(),
		name:        data.Name,
		description: data.Description,
		icon:        iconSecureUrl,
		banner:      bannerSecureUrl,
		userId:      data.UserId,
		createdAt:   time.Now().Unix(),
	})
	if err != nil {
		return schema.Response[subforumResponse]{}, err
	}

	return schema.Response[subforumResponse]{
		Status: "success",
		Code:   http.StatusCreated,
		Data: subforumResponse{
			UserId:      result.userId,
			Id:          result.id,
			Name:        result.name,
			Description: result.description,
			CreatedAt:   result.createdAt,
		},
	}, nil
}

// func (service *ServiceImpl) takeDown(ctx context.Context){}
