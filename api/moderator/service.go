package moderator

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
	"github.com/zulfikarrosadi/code_roast/schema"
)

type repository interface {
	addRoles(context.Context, string, []int) (userAndRole, error)
	removeRoles(context.Context, string, []int) (userAndRole, error)
}

type serviceImpl struct {
	repository
	v *validator.Validate
}

type updateRoleRequest struct {
	RoleId []int `json:"role_id" validate:"required"`
}

type updateRoleResponse struct {
	UserId string  `json:"id"`
	Roles  []roles `json:"roles"`
}

type UpdatePermissionResponse struct {
	User updateRoleResponse `json:"user"`
}

func (service *serviceImpl) addRoles(
	ctx context.Context,
	userId string,
	role updateRoleRequest,
) (schema.Response[UpdatePermissionResponse], error) {
	if err := service.v.Struct(role); err != nil {
		validationErrorDetail := apperror.HandlerValidatorError(err.(validator.ValidationErrors))
		return schema.Response[UpdatePermissionResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: apperror.VALIDATION_ERROR,
				Details: validationErrorDetail,
			},
		}, fmt.Errorf("service: input validation error %w", err)
	}

	result, err := service.repository.addRoles(ctx, userId, role.RoleId)
	if err != nil {
		var appError apperror.AppError
		if errors.As(err, &appError) {
			return schema.Response[UpdatePermissionResponse]{
				Status: "fail",
				Code:   appError.Code,
				Error: schema.Error{
					Message: appError.Message,
				},
			}, err
		}
		return schema.Response[UpdatePermissionResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "something went wrong, please try again later",
			},
		}, err
	}
	return schema.Response[UpdatePermissionResponse]{
		Status: "success",
		Code:   http.StatusCreated,
		Data: UpdatePermissionResponse{
			User: updateRoleResponse{
				UserId: userId,
				Roles:  result.roles,
			},
		},
	}, nil
}

func (service *serviceImpl) removeRoles(
	ctx context.Context,
	userId string,
	role updateRoleRequest,
) (schema.Response[UpdatePermissionResponse], error) {
	if err := service.v.Struct(role); err != nil {
		validationErrorDetail := apperror.HandlerValidatorError(err.(validator.ValidationErrors))
		return schema.Response[UpdatePermissionResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: apperror.VALIDATION_ERROR,
				Details: validationErrorDetail,
			},
		}, fmt.Errorf("service: input validation error %w", err)
	}
	result, err := service.repository.removeRoles(ctx, userId, role.RoleId)
	if err != nil {
		var appError apperror.AppError
		if errors.As(err, &appError) {
			return schema.Response[UpdatePermissionResponse]{
				Status: "fail",
				Code:   appError.Code,
				Error: schema.Error{
					Message: appError.Message,
				},
			}, err
		}
		return schema.Response[UpdatePermissionResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "something went wrong, please try again later",
			},
		}, err
	}
	return schema.Response[UpdatePermissionResponse]{
		Status: "success",
		Code:   http.StatusOK,
		Data: UpdatePermissionResponse{
			User: updateRoleResponse{
				UserId: userId,
				Roles:  result.roles,
			},
		},
	}, nil
}
