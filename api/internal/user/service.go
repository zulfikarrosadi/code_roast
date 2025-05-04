package user

import (
	"context"
	"errors"

	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type repository interface {
	findById(ctx context.Context, id string) (User, error)
}

type serviceImpl struct {
	repo repository
}

func NewService(repo repository) serviceImpl {
	return serviceImpl{
		repo: repo,
	}
}

type UserDTO struct {
	Id       string `json:"id"`
	Fullname string `json:"fullname"`
	Email    string `json:"email"`
}

type FindByIdResponse struct {
	User UserDTO `json:"user"`
}

func (s serviceImpl) findById(ctx context.Context, id string) (schema.Response[FindByIdResponse], error) {
	result, err := s.repo.findById(ctx, id)
	if errors.Is(err, &UserNotFound{}) {
		return schema.Response[FindByIdResponse]{
			Status: "fail",
			Code:   404,
			Error: schema.Error{
				Message: err.Error(),
			},
		}, err
	} else if err != nil {
		return schema.Response[FindByIdResponse]{
			Status: "fail",
			Code:   500,
			Error: schema.Error{
				Message: "something went wrong, please try again later",
			},
		}, err
	}

	return schema.Response[FindByIdResponse]{
		Status: "success",
		Code:   200,
		Data: FindByIdResponse{
			User: UserDTO{
				Id:       result.Id,
				Fullname: result.Fullname,
				Email:    result.Email,
			},
		},
	}, nil
}
