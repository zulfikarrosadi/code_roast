package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	apperror "github.com/zulfikarrosadi/code_roast/internal/app-error"

	"github.com/zulfikarrosadi/code_roast/internal/middleware"
	"github.com/zulfikarrosadi/code_roast/internal/user"
	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type Service interface {
	register(context.Context, registrationRequest) (schema.Response[authResponse], error)
	login(context.Context, loginRequest) (schema.Response[authResponse], error)
	refreshToken(context.Context, string) (schema.Response[authResponse], error)
}

type ApiHandler struct {
	*slog.Logger
	Service
}

func NewApiHandler(logger *slog.Logger, service Service) *ApiHandler {
	return &ApiHandler{
		Logger:  logger,
		Service: service,
	}
}

func GetUserFromContext(c echo.Context) (*CustomJWTClaims, error) {
	userData := c.Get("user").(*jwt.Token)
	if userData == nil {
		return nil, echo.NewHTTPError(401, "User not found in context")
	}
	claims, ok := userData.Claims.(*CustomJWTClaims)
	if !ok {
		return nil, echo.NewHTTPError(401, "Invalid user claims")
	}

	return claims, nil
}

const (
	WEEK_IN_SECOND     = 604_800
	REQUEST_ID_KEY     = "REQUEST_ID"
	REFRESH_TOKEN_NAME = "refresh_token"
	ACCESS_TOKEN_NAME  = "access_token"
)

func (api *ApiHandler) Current(c echo.Context) error {
	result, err := GetUserFromContext(c)
	if err != nil {
		return err
	}
	response := schema.Response[user.FindByIdResponse]{
		Status: "success",
		Code:   http.StatusOK,
		Data: user.FindByIdResponse{
			User: user.UserDTO{
				Id:       result.Id,
				Fullname: result.Fullname,
				Email:    result.Email,
			},
		},
	}

	err = c.JSON(response.Code, response)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
	}
	return nil
}

func (api *ApiHandler) LogOut(c echo.Context) error {
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	_, err := GetUserFromContext(c)
	if err != nil {
		logger.Warn("get user from context fail", slog.String("error", err.Error()))
		return err
	}

	c.SetCookie(&http.Cookie{
		Name:     REFRESH_TOKEN_NAME,
		Value:    "",
		Secure:   true,
		MaxAge:   0,
		Path:     "/api/v1/refresh",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	return c.NoContent(http.StatusNoContent)
}

func (api *ApiHandler) RefreshToken(c echo.Context) error {
	refreshToken, err := c.Request().Cookie(REFRESH_TOKEN_NAME)
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	if err != nil {
		logger.Warn("refresh token not available in cookie", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusUnauthorized, "something went wrong, refresh token extraction from cookie fails")
	}
	response, err := api.refreshToken(ctx, refreshToken.Value)
	if err != nil {
		logger.Debug("refresh token service failed", slog.String("error", err.Error()))
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}
	c.SetCookie(&http.Cookie{
		Name:     REFRESH_TOKEN_NAME,
		Value:    response.Data.RefreshToken,
		Secure:   true,
		MaxAge:   WEEK_IN_SECOND,
		Path:     "/api/v1/refresh",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	if err := c.JSON(response.Code, response); err != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
	}
	return nil
}

func (api *ApiHandler) Login(c echo.Context) error {
	user := new(loginRequest)
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	if err := c.Bind(user); err != nil {
		logger.Debug("failed to bind data user request", slog.String("error", err.Error()))
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"fail to process your request, send corerct data and try again",
		)
	}
	user.authentication = authentication{
		lastLogin: time.Now().Unix(),
		remoteIP:  c.Request().RemoteAddr,
		agent:     c.Request().UserAgent(),
	}
	response, err := api.Service.login(ctx, *user)
	if err != nil {
		if response.Error.Message == apperror.VALIDATION_ERROR {
			logger.Debug("request validation fail", slog.Any("error", err))
			err = c.JSON(response.Code, response)
			if err != nil {
				logger.Error("failed to write JSON response", slog.String("error", err.Error()))
				return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
			}
			return nil
		}

		logger.Debug("login service fail", slog.Any("error", err))
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}

	c.SetCookie(&http.Cookie{
		Name:     REFRESH_TOKEN_NAME,
		Value:    response.Data.RefreshToken,
		Secure:   true,
		MaxAge:   WEEK_IN_SECOND,
		Path:     "/api/v1/refresh",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	err = c.JSON(response.Code, response)
	if err != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"something went wrong, please try again later",
		)
	}
	return nil
}

func (api *ApiHandler) Register(c echo.Context) error {
	user := new(registrationRequest)
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	if err := c.Bind(user); err != nil {
		logger.Debug("failed to bind data user request", slog.String("error", err.Error()))
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"fail to process your request, send corerct data and try again",
		)
	}
	response, err := api.Service.register(ctx, registrationRequest{
		Id:                   user.Id,
		Fullname:             user.Fullname,
		PasswordConfirmation: user.PasswordConfirmation,
		Email:                user.Email,
		Password:             user.Password,
		Agent:                c.Request().UserAgent(),
		RemoteIp:             c.Request().RemoteAddr,
	})
	if err != nil {
		if response.Error.Message == apperror.VALIDATION_ERROR {
			logger.Debug("request validation fail", slog.Any("error", err))
			err = c.JSON(response.Code, response)
			if err != nil {
				logger.Error("failed to write JSON response", slog.String("error", err.Error()))
				return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
			}
			return nil
		}

		logger.Debug("register service fail", slog.String("error", err.Error()))
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}
	c.SetCookie(&http.Cookie{
		Name:     REFRESH_TOKEN_NAME,
		Value:    response.Data.RefreshToken,
		Secure:   true,
		MaxAge:   WEEK_IN_SECOND,
		Path:     "/api/v1/refresh",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
	err = c.JSON(response.Code, response)
	if err != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"something went wrong, please try again later",
		)
	}
	return nil
}
