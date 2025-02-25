package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
	"github.com/zulfikarrosadi/code_roast/schema"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	register(context.Context, userAndAuth) (publicUserData, error)
	findByEmail(context.Context, string) (User, error)
}

type ServiceImpl struct {
	*slog.Logger
	Repository
}

func NewUserService(logger *slog.Logger, repo Repository) *ServiceImpl {
	return &ServiceImpl{
		Logger:     logger,
		Repository: repo,
	}
}

type CustomJWTClaims struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	jwt.RegisteredClaims
}

func (service *ServiceImpl) register(
	ctx context.Context,
	newUser userCreateRequest,
) (schema.Response[authResponse], error) {
	newUserId, err := uuid.NewV7()
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create your account, please try again later",
			},
		}, fmt.Errorf("service: fail generate new user id, %w", err)
	}
	refreshToken, err := uuid.NewV7()
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create your account, please try again later",
			},
		}, fmt.Errorf("service: fail generate new refresh token, %w", err)
	}
	authenticationId, err := uuid.NewV7()
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create your account, please try again later",
			},
		}, fmt.Errorf("service: fail generate new authentication id, %w", err)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create your account, please try again later",
			},
		}, fmt.Errorf("service: fail generate hash from user password, %w", err)
	}

	user, err := service.Repository.register(ctx, userAndAuth{
		id:        newUserId.String(),
		fullname:  newUser.Fullname,
		email:     newUser.Email,
		password:  string(hashedPassword),
		createdAt: time.Now().Unix(),
		authentication: authentication{
			id:           authenticationId.String(),
			refreshToken: refreshToken.String(),
			lastLogin:    time.Now().Unix(),
			userId:       newUserId.String(),
			agent:        newUser.Agent,
			remoteIP:     newUser.RemoteIp,
		},
	})
	if err != nil {
		fmt.Println("error service 1: ", err)
		var appError *apperror.AppError
		if errors.As(err, &appError) {
			return schema.Response[authResponse]{
				Status: "fail",
				Code:   appError.Code,
				Error: schema.Error{
					Message: appError.Message,
				},
			}, err
		}
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to process your request, please try again later",
			},
		}, err
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomJWTClaims{
		Id:       user.id,
		Email:    user.email,
		Fullname: user.fullname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	},
	).SignedString([]byte(os.Getenv("JWT_SECRETS")))
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to create your account, please try again later",
			},
		}, fmt.Errorf("service: fail to generate new access token, %w", err)
	}
	return schema.Response[authResponse]{
		Status: "success",
		Code:   http.StatusCreated,
		Data: authResponse{
			User: userCreateResponse{
				ID:       user.id,
				Email:    user.email,
				Fullname: user.fullname,
			},
			AccessToken:  accessToken,
			RefreshToken: refreshToken.String(),
		},
	}, nil
}

func (service *ServiceImpl) login(
	ctx context.Context,
	user userLoginRequest,
) (schema.Response[authResponse], error) {
	refreshToken, err := uuid.NewV7()
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail process your request, please try again later",
			},
		}, fmt.Errorf("service: fail generate new refresh token, %w", err)
	}
	result, err := service.Repository.findByEmail(ctx, user.Email)
	if err != nil {
		fmt.Println("service: ", err)
		var authError authError
		if errors.As(err, &authError) {
			return schema.Response[authResponse]{
				Status: "fail",
				Code:   authError.Code,
				Error: schema.Error{
					Message: authError.Msg,
				},
			}, err
		}
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail process your request, please try again later",
			},
		}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: "email or password is incorrect",
			},
		}, fmt.Errorf("service: comparing password failed, %w", err)
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomJWTClaims{
		Id:       result.Id,
		Email:    result.Email,
		Fullname: result.Fullname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	},
	).SignedString([]byte(os.Getenv("JWT_SECRETS")))
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail to process your request, please try again later",
			},
		}, fmt.Errorf("service: fail to generate new access token, %w", err)
	}
	return schema.Response[authResponse]{
		Status: "success",
		Code:   http.StatusOK,
		Data: authResponse{
			User: userCreateResponse{
				ID:       result.Id,
				Email:    result.Email,
				Fullname: result.Fullname,
			},
			AccessToken:  accessToken,
			RefreshToken: refreshToken.String(),
		},
	}, nil
}
