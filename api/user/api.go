package user

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Service interface {
	Create(context.Context, userCreateRequest) (*SuccessResponse[authResponse], *ErrorResponse)
	Login(context.Context, userLoginRequest) (*SuccessResponse[authResponse], *ErrorResponse)
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

const WEEK_IN_SECOND = 604_800

func (api *ApiHandler) Login(c echo.Context) error {
	user := new(userLoginRequest)
	ctx := context.WithValue(context.TODO(), "REQUEST_ID", c.Response().Header().Get(echo.HeaderXRequestID))

	if err := c.Bind(user); err != nil {
		api.Logger.LogAttrs(ctx,
			slog.LevelError,
			"REQUEST_ERROR",
			slog.Group("details",
				slog.String("message", "Failed to assert claims"),
				slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
			))
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"fail to process your request, send corerct data and try again",
		)
	}
	successResponse, errResponse := api.Service.Login(ctx, *user)
	if errResponse != nil {
		return echo.NewHTTPError(errResponse.Error.Code, errResponse.Error.Message)
	}

	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    successResponse.Data.RefreshToken,
		Secure:   true,
		MaxAge:   WEEK_IN_SECOND,
		Path:     "/api/v1/refresh",
		HttpOnly: true,
	})
	err := c.JSON(http.StatusOK, successResponse)
	if err != nil {
		api.Logger.LogAttrs(ctx,
			slog.LevelError,
			"REQUEST_ERROR",
			slog.Group("details",
				slog.String("message", "fail sending JSON response"),
				slog.String("request_id", ctx.Value("REQUEST_ID").(string)),
			))
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"something went wrong, please try again later",
		)
	}
	return nil
}

func (api *ApiHandler) Register(c echo.Context) error {
	newUser := new(userCreateRequest)
	ctx := context.WithValue(context.TODO(), "REQUEST_ID", c.Response().Header().Get(echo.HeaderXRequestID))

	if err := c.Bind(newUser); err != nil {
		api.Logger.LogAttrs(
			c.Request().Context(),
			slog.LevelError,
			"fail request binding",
			slog.Any("details", err),
		)
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"fail to process your request, send corerct data and try again",
		)
	}
	successRes, errorRes := api.Service.Create(ctx, *newUser)
	if errorRes != nil {
		return echo.NewHTTPError(errorRes.Error.Code, errorRes.Error.Message)
	}
	c.SetCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    successRes.Data.RefreshToken,
		Secure:   true,
		MaxAge:   WEEK_IN_SECOND,
		Path:     "/api/v1/refresh",
		HttpOnly: true,
	})
	err := c.JSON(http.StatusCreated, successRes)
	if err != nil {
		api.Logger.LogAttrs(
			c.Request().Context(),
			slog.LevelError, "fail sending json response",
			slog.Any("details", err),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"something went wrong, please try again later",
		)
	}
	return nil
}
