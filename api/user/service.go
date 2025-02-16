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
	"github.com/zulfikarrosadi/code_roast/lib"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	Create(context.Context, User) (User, error)
	FindUserByEmail(context.Context, string) (User, error)
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

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details"`
}

type SuccessResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  Error  `json:"error"`
}

type CustomJWTClaims struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
	jwt.RegisteredClaims
}

func (service *ServiceImpl) Login(
	ctx context.Context,
	user userLoginRequest,
) (*SuccessResponse[authResponse], *ErrorResponse) {
	refreshToken, err := uuid.NewV7()
	if err != nil {
		service.Logger.LogAttrs(ctx,
			slog.LevelError,
			"REQUEST_ERROR",
			slog.Group("details",
				slog.String("message", "fail to generate refresh token"),
				slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
			))
		return nil, &ErrorResponse{
			Status: "fail",
			Error: Error{
				Code:    http.StatusInternalServerError,
				Message: "fail to process your request, please try again later",
			},
		}
	}
	result, err := service.Repository.FindUserByEmail(ctx, user.Email)
	if err != nil {
		service.Logger.LogAttrs(ctx,
			slog.LevelError,
			"REQUEST_ERROR",
			slog.Group("details",
				slog.String("message", err.Error()),
				slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
			))
		var authError lib.AuthError
		if errors.As(err, &authError) {
			return nil, &ErrorResponse{
				Status: "fail",
				Error: Error{
					Code:    authError.Code,
					Message: authError.Msg,
				},
			}
		}
		return nil, &ErrorResponse{
			Status: "fail",
			Error: Error{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			},
		}
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))
	if err != nil {
		service.Logger.LogAttrs(ctx,
			slog.LevelError,
			"REQUEST_ERROR",
			slog.Group("details",
				slog.String("message", err.Error()),
				slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
			))
		return nil, &ErrorResponse{
			Status: "fail",
			Error: Error{
				Code:    http.StatusBadRequest,
				Message: "email or password is invalid",
			},
		}
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
		service.Logger.LogAttrs(ctx,
			slog.LevelError,
			"REQUEST_ERROR",
			slog.Group("details",
				slog.String("message", "fail create access token"),
				slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
			))
		return nil, &ErrorResponse{
			Status: "fail",
			Error: Error{
				Code:    http.StatusInternalServerError,
				Message: "fail to process your request, please try again later",
			},
		}
	}
	return &SuccessResponse[authResponse]{
		Status: "success",
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

func (service *ServiceImpl) Create(
	ctx context.Context,
	newUser userCreateRequest,
) (*SuccessResponse[authResponse], *ErrorResponse) {
	newUserId, err := uuid.NewV7()
	if err != nil {
		service.Logger.LogAttrs(
			ctx,
			slog.LevelError,
			"fail to generate uuid v7",
			slog.Any("details", err),
		)
		return nil, &ErrorResponse{
			Status: "fail",
			Error: Error{
				Code:    http.StatusInternalServerError,
				Message: "fail to create your account, please try again later",
			},
		}
	}

	refreshToken, err := uuid.NewV7()
	if err != nil {
		service.Logger.LogAttrs(
			ctx,
			slog.LevelError,
			"fail to generate refresh token",
			slog.Any("details", err),
		)
		return nil, &ErrorResponse{
			Status: "fail",
			Error: Error{
				Code:    http.StatusInternalServerError,
				Message: "fail to create your account, please try again later",
			},
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
	if err != nil {
		service.Logger.LogAttrs(ctx,
			slog.LevelError,
			"REQUEST_ERROR",
			slog.Group("details",
				slog.String("message", err.Error()),
				slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
			))
		return nil, &ErrorResponse{
			Status: "fail",
			Error: Error{
				Code:    http.StatusInternalServerError,
				Message: "fail to create your account, please try again later",
			},
		}
	}

	user, err := service.Repository.Create(ctx, User{
		Id:        newUserId.String(),
		Fullname:  newUser.Fullname,
		Email:     newUser.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().Unix(),
	})
	if err != nil {
		fmt.Println(ctx)
		service.Logger.LogAttrs(
			ctx,
			slog.LevelError,
			"repo error in service",
			slog.Any("details", err),
		)
		var authError lib.AuthError
		if errors.As(err, &authError) {
			return nil, &ErrorResponse{
				Status: "fail",
				Error: Error{
					Code:    authError.Code,
					Message: authError.Msg,
				},
			}
		}
		return nil, &ErrorResponse{
			Status: "fail",
			Error: Error{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			},
		}
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomJWTClaims{
		Id:       user.Id,
		Email:    user.Email,
		Fullname: user.Fullname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	},
	).SignedString([]byte(os.Getenv("JWT_SECRETS")))
	if err != nil {
		service.Logger.LogAttrs(
			ctx,
			slog.LevelError,
			"fail create access token",
			slog.Any("details", err),
		)
		return nil, &ErrorResponse{
			Status: "fail",
			Error: Error{
				Code:    http.StatusInternalServerError,
				Message: "fail to process your request, please try again later",
			},
		}
	}
	return &SuccessResponse[authResponse]{
		Status: "success",
		Data: authResponse{
			User: userCreateResponse{
				ID:       user.Id,
				Email:    user.Email,
				Fullname: user.Fullname,
			},
			AccessToken:  accessToken,
			RefreshToken: refreshToken.String(),
		},
	}, nil
}
