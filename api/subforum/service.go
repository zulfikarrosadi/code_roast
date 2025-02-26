package subforum

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
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
}

type subforumResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserId      string `json:"user_id"`
	CreatedAt   int64  `json:"created_at"`
}

func NewService(repo repository, v *validator.Validate) *ServiceImpl {
	return &ServiceImpl{
		repo: repo,
		v:    v,
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
	subForumId, err := uuid.NewV7()
	if err != nil {
		return schema.Response[subforumResponse]{}, fmt.Errorf("service: fail to generate subforum uuid v7 %w", err)
	}

	result, err := service.repo.create(ctx, subforum{
		id:          subForumId.String(),
		name:        data.Name,
		description: data.Description,
		icon:        data.Icon,
		banner:      data.Banner,
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
