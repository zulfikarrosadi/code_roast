package subforum

import (
	"context"
	"errors"
	"fmt"
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
	create(context.Context, Subforum) (Subforum, error)
	findByName(context.Context, string) ([]Subforum, error)
	deleteById(context.Context, string, string) error
	findAll(context.Context) ([]Subforum, error)
}

type ServiceImpl struct {
	repo repository
	v    *validator.Validate
	cld  *cloudinary.Cloudinary
}

type SubforumMedia struct {
	Icon   string `json:"icon"`
	Banner string `json:"banner"`
}

type subforumDetailDTO struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	CreatedAt     int64  `json:"created_at"`
	SubforumMedia `json:"media"`
}

type createResponse struct {
	Subforum subforumDetailDTO `json:"subforum"`
}

type getAllResponse struct {
	Subforums []subforumDetailDTO `json:"subforums"`
}

func NewService(repo repository, v *validator.Validate, cloudinaryInstance *cloudinary.Cloudinary) *ServiceImpl {
	return &ServiceImpl{
		repo: repo,
		v:    v,
		cld:  cloudinaryInstance,
	}
}

func (service *ServiceImpl) create(ctx context.Context, data subforumCreateRequest) (schema.Response[createResponse], error) {
	err := service.v.Struct(data)
	if err != nil {
		validatorError := apperror.HandlerValidatorError(err.(validator.ValidationErrors))
		return schema.Response[createResponse]{
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
		return schema.Response[createResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create new subforum, failed to open icon file",
			},
		}, fmt.Errorf("service: failed to open icon file %w", err)
	}
	defer iconSrc.Close()
	if _, err := imagehelper.IsImage(iconSrc); err != nil {
		return schema.Response[createResponse]{
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
		return schema.Response[createResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create new subforum, failed to upload icon file",
			},
		}, fmt.Errorf("service: failed to upload icon file %w", err)
	}
	bannerSrc, err := data.Banner.Open()
	if err != nil {
		return schema.Response[createResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create new subforum, failed to open banner file",
			},
		}, fmt.Errorf("service: failed to open banner file %w", err)
	}
	defer bannerSrc.Close()
	if _, err := imagehelper.IsImage(bannerSrc); err != nil {
		return schema.Response[createResponse]{
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
		return schema.Response[createResponse]{
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
		return schema.Response[createResponse]{}, fmt.Errorf("service: fail to generate subforum uuid v7 %w", err)
	}

	result, err := service.repo.create(ctx, Subforum{
		Id:          subForumId.String(),
		Name:        data.Name,
		Description: data.Description,
		Icon:        iconSecureUrl,
		Banner:      bannerSecureUrl,
		UserId:      data.UserId,
		CreatedAt:   time.Now().Unix(),
	})
	if err != nil {
		return schema.Response[createResponse]{}, err
	}

	return schema.Response[createResponse]{
		Status: "success",
		Code:   http.StatusCreated,
		Data: createResponse{
			Subforum: subforumDetailDTO{
				Id:          result.Id,
				Name:        result.Name,
				Description: result.Description,
				CreatedAt:   result.CreatedAt,
				SubforumMedia: SubforumMedia{
					Icon:   iconSecureUrl,
					Banner: bannerSecureUrl,
				},
			},
		},
	}, nil
}

func (service *ServiceImpl) getAll(ctx context.Context) (schema.Response[getAllResponse], error) {
	result, err := service.repo.findAll(ctx)
	if err != nil {
		return schema.Response[getAllResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "Something went wrong, please try again later",
			},
		}, fmt.Errorf("service error: %w", err)
	}
	if len(result) == 0 {
		return schema.Response[getAllResponse]{
			Status: "fail",
			Code:   http.StatusNotFound,
			Error: schema.Error{
				Message: "No subforums found",
			},
		}, errors.New("no subforum found")
	}

	subforumsDetail := []subforumDetailDTO{}
	for _, v := range result {
		subforumDetail := subforumDetailDTO{
			Id:          v.Id,
			Name:        v.Name,
			Description: v.Description,
			CreatedAt:   v.CreatedAt,
			SubforumMedia: SubforumMedia{
				Icon:   v.Icon,
				Banner: v.Banner,
			},
		}
		subforumsDetail = append(subforumsDetail, subforumDetail)
	}
	return schema.Response[getAllResponse]{
		Status: "success",
		Code:   200,
		Data: getAllResponse{
			Subforums: subforumsDetail,
		},
	}, nil
}

// func (service *ServiceImpl) takeDown(ctx context.Context){}
