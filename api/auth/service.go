package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
	"github.com/zulfikarrosadi/code_roast/schema"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	register(context.Context, user, authentication) (publicUserData, error)
	loginByEmail(context.Context, string, authentication) (user, error)
	findRefreshToken(context.Context, string) (publicUserData, error)
}

type ServiceImpl struct {
	Repository
	v *validator.Validate
}

func NewUserService(repo Repository, v *validator.Validate) *ServiceImpl {
	return &ServiceImpl{
		Repository: repo,
		v:          v,
	}
}

type CustomJWTClaims struct {
	Id       string  `json:"id"`
	Email    string  `json:"email"`
	Fullname string  `json:"fullname"`
	Roles    []roles `json:"roles"`
	jwt.RegisteredClaims
}

type registrationRequest struct {
	Id                   string `json:"id"`
	Fullname             string `json:"fullname" validate:"required"`
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"required"`
	PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password"`
	Agent                string
	RemoteIp             string
}

type loginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	authentication
}

type registrationResponse struct {
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	Fullname string  `json:"fullname"`
	Roles    []roles `json:"roles"`
}

// we need this to standarize auth resposne
type authResponse struct {
	User         registrationResponse `json:"user"`
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
}

var JWT_SECRET = os.Getenv("JWT_SECRET")

func (service *ServiceImpl) refreshToken(ctx context.Context, token string) (schema.Response[authResponse], error) {
	user, err := service.findRefreshToken(ctx, token)
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusUnauthorized,
			Error: schema.Error{
				Message: "refresh token lookup not found",
			},
		}, err
	}
	newAccessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomJWTClaims{
		Id:       user.id,
		Email:    user.email,
		Fullname: user.fullname,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	},
	).SignedString([]byte(JWT_SECRET))
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "something went wrong, generate new access token fail, please try again later",
			},
		}, fmt.Errorf("service: generate new access token fail %w", err)
	}
	return schema.Response[authResponse]{
		Status: "success",
		Code:   http.StatusOK,
		Data: authResponse{
			User: registrationResponse{
				ID:       user.id,
				Email:    user.email,
				Fullname: user.fullname,
				Roles:    user.roles,
			},
			AccessToken:  newAccessToken,
			RefreshToken: token,
		},
	}, nil
}

func (service *ServiceImpl) register(
	ctx context.Context,
	newUser registrationRequest,
) (schema.Response[authResponse], error) {
	fmt.Println(newUser)
	err := service.v.Struct(newUser)
	if err != nil {
		validatorError := apperror.HandlerValidatorError(err.(validator.ValidationErrors))
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: apperror.VALIDATION_ERROR,
				Details: validatorError,
			},
		}, fmt.Errorf("service: create new user validation error %w", err)
	}

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

	user, err := service.Repository.register(
		ctx,
		user{
			id:        newUserId.String(),
			fullname:  newUser.Fullname,
			email:     newUser.Email,
			password:  string(hashedPassword),
			createdAt: time.Now().Unix(),
		},
		authentication{
			id:           authenticationId.String(),
			refreshToken: refreshToken.String(),
			lastLogin:    time.Now().Unix(),
			userId:       newUserId.String(),
			agent:        newUser.Agent,
			remoteIP:     newUser.RemoteIp,
		},
	)
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
		Roles:    user.roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	},
	).SignedString([]byte(JWT_SECRET))
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
			User: registrationResponse{
				ID:       user.id,
				Email:    user.email,
				Fullname: user.fullname,
				Roles:    user.roles,
			},
			AccessToken:  accessToken,
			RefreshToken: refreshToken.String(),
		},
	}, nil
}

func (service *ServiceImpl) login(
	ctx context.Context,
	user loginRequest,
) (schema.Response[authResponse], error) {
	err := service.v.Struct(user)
	if err != nil {
		validationErrorDetail := apperror.HandlerValidatorError(err.(validator.ValidationErrors))
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusBadRequest,
			Error: schema.Error{
				Message: apperror.VALIDATION_ERROR,
				Details: validationErrorDetail,
			},
		}, fmt.Errorf("service: input validation error %w", err)
	}

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
	authId, err := uuid.NewV7()
	if err != nil {
		return schema.Response[authResponse]{
			Status: "fail",
			Code:   http.StatusInternalServerError,
			Error: schema.Error{
				Message: "fail process your request, please try again later",
			},
		}, fmt.Errorf("service: fail generate new auth id, %w", err)
	}
	result, err := service.Repository.loginByEmail(
		ctx,
		user.Email,
		authentication{
			id:           authId.String(),
			refreshToken: refreshToken.String(),
			lastLogin:    user.authentication.lastLogin,
			remoteIP:     user.authentication.remoteIP,
			agent:        user.authentication.agent,
		},
	)
	if err != nil {
		var appError apperror.AppError
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
				Message: "fail process your request, please try again later",
			},
		}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(result.password), []byte(user.Password))
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
		Id:       result.id,
		Email:    result.email,
		Fullname: result.fullname,
		Roles:    result.roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	},
	).SignedString([]byte(JWT_SECRET))
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
			User: registrationResponse{
				ID:       result.id,
				Email:    result.email,
				Fullname: result.fullname,
				Roles:    result.roles,
			},
			AccessToken:  accessToken,
			RefreshToken: refreshToken.String(),
		},
	}, nil
}
