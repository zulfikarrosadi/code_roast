package user

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
	"github.com/zulfikarrosadi/code_roast/schema"
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

const (
	WEEK_IN_SECOND     = 604_800
	REQUEST_ID_KEY     = "REQUEST_ID"
	REFRESH_TOKEN_NAME = "refresh_token"
)

func (api *ApiHandler) RefreshToken(c echo.Context) error {
	refreshToken, err := c.Request().Cookie(REFRESH_TOKEN_NAME)
	ctx := context.WithValue(context.TODO(), REQUEST_ID_KEY, c.Response().Header().Get(echo.HeaderXRequestID))

	if err != nil {
		api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
			slog.Int("status", http.StatusInternalServerError),
			slog.Group("request",
				slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("ip", c.Request().RemoteAddr),
				slog.Any("authorization", c.Request().Header.Get("Authorization")),
			),
			slog.String("error", err.Error()),
			slog.String("trace", string(debug.Stack())),
		)
		return echo.NewHTTPError(http.StatusUnauthorized, "something went wrong, refresh token extraction from cookie fails")
	}
	response, err := api.refreshToken(ctx, refreshToken.Value)
	if err != nil {
		api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
			slog.Int("status", http.StatusInternalServerError),
			slog.Group("request",
				slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("ip", c.Request().RemoteAddr),
				slog.Any("authorization", c.Request().Header.Get("Authorization")),
			),
			slog.String("error", err.Error()),
			slog.String("trace", string(debug.Stack())),
		)
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}
	c.SetCookie(&http.Cookie{
		Name:     REFRESH_TOKEN_NAME,
		Value:    response.Data.RefreshToken,
		Secure:   true,
		MaxAge:   WEEK_IN_SECOND,
		Path:     "/api/v1/refresh",
		HttpOnly: true,
	})
	if err := c.JSON(response.Code, response); err != nil {
		api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
			slog.Int("status", http.StatusInternalServerError),
			slog.Group("request",
				slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("ip", c.Request().RemoteAddr),
				slog.Any("authorization", c.Request().Header.Get("Authorization")),
			),
			slog.String("error", err.Error()),
			slog.String("trace", string(debug.Stack())),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
	}
	return nil
}

func (api *ApiHandler) Login(c echo.Context) error {
	user := new(loginRequest)
	ctx := context.WithValue(context.TODO(), REQUEST_ID_KEY, c.Response().Header().Get(echo.HeaderXRequestID))

	if err := c.Bind(user); err != nil {
		api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
			slog.Int("status", http.StatusInternalServerError),
			slog.Group("request",
				slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("ip", c.Request().RemoteAddr),
				slog.Any("authorization", c.Request().Header.Get("Authorization")),
			),
			slog.String("error", err.Error()),
			slog.String("trace", string(debug.Stack())),
		)
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
			api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
				slog.Int("status", response.Code),
				slog.Group("request",
					slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
					slog.String("method", c.Request().Method),
					slog.String("path", c.Request().URL.Path),
					slog.String("user_agent", c.Request().UserAgent()),
					slog.String("ip", c.Request().RemoteAddr),
					slog.Any("authorization", c.Request().Header.Get("Authorization")),
				),
				slog.String("error", err.Error()),
				slog.String("trace", string(debug.Stack())),
			)
			err = c.JSON(response.Code, response)
			if err != nil {
				api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
					slog.Int("status", response.Code),
					slog.Group("request",
						slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
						slog.String("method", c.Request().Method),
						slog.String("path", c.Request().URL.Path),
						slog.String("user_agent", c.Request().UserAgent()),
						slog.String("ip", c.Request().RemoteAddr),
						slog.Any("authorization", c.Request().Header.Get("Authorization")),
					),
					slog.String("error", err.Error()),
					slog.String("trace", string(debug.Stack())),
				)
				return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
			}
			return nil
		}
		api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
			slog.Int("status", response.Code),
			slog.Group("request",
				slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("ip", c.Request().RemoteAddr),
				slog.Any("authorization", c.Request().Header.Get("Authorization")),
			),
			slog.String("error", err.Error()),
			slog.String("trace", string(debug.Stack())),
		)
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}

	c.SetCookie(&http.Cookie{
		Name:     REFRESH_TOKEN_NAME,
		Value:    response.Data.RefreshToken,
		Secure:   true,
		MaxAge:   WEEK_IN_SECOND,
		Path:     "/api/v1/refresh",
		HttpOnly: true,
	})
	err = c.JSON(response.Code, response)
	if err != nil {
		api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
			slog.Int("status", response.Code),
			slog.Group("request",
				slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("ip", c.Request().RemoteAddr),
				slog.Any("authorization", c.Request().Header.Get("Authorization")),
			),
			slog.String("error", err.Error()),
			slog.String("trace", string(debug.Stack())),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"something went wrong, please try again later",
		)
	}
	return nil
}

func (api *ApiHandler) Register(c echo.Context) error {
	user := new(registrationRequest)
	ctx := context.WithValue(context.TODO(), "REQUEST_ID", c.Response().Header().Get(echo.HeaderXRequestID))

	if err := c.Bind(user); err != nil {
		api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
			slog.Int("status", http.StatusBadRequest),
			slog.Group("request",
				slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("ip", c.Request().RemoteAddr),
				slog.Any("authorization", c.Request().Header.Get("Authorization")),
			),
			slog.String("error", err.Error()),
			slog.String("trace", string(debug.Stack())),
		)
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
			err = c.JSON(response.Code, response)
			if err != nil {
				api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
					slog.Int("status", http.StatusInternalServerError),
					slog.Group("request",
						slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
						slog.String("method", c.Request().Method),
						slog.String("path", c.Request().URL.Path),
						slog.String("user_agent", c.Request().UserAgent()),
						slog.String("ip", c.Request().RemoteAddr),
						slog.Any("authorization", c.Request().Header.Get("Authorization")),
					),
					slog.String("error", err.Error()),
					slog.String("trace", string(debug.Stack())),
				)
				return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
			}
			return nil
		}
		api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
			slog.Int("status", response.Code),
			slog.Group("request",
				slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("ip", c.Request().RemoteAddr),
				slog.Any("authorization", c.Request().Header.Get("Authorization")),
			),
			slog.String("error", err.Error()),
			slog.String("trace", string(debug.Stack())),
		)
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}
	c.SetCookie(&http.Cookie{
		Name:     REFRESH_TOKEN_NAME,
		Value:    response.Data.RefreshToken,
		Secure:   true,
		MaxAge:   WEEK_IN_SECOND,
		Path:     "/api/v1/refresh",
		HttpOnly: true,
	})
	err = c.JSON(response.Code, response)
	if err != nil {
		api.Logger.LogAttrs(ctx, slog.LevelDebug, "REQUEST_DEBUG",
			slog.Int("status", http.StatusInternalServerError),
			slog.Group("request",
				slog.String("id", ctx.Value(REQUEST_ID_KEY).(string)),
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("ip", c.Request().RemoteAddr),
				slog.Any("authorization", c.Request().Header.Get("Authorization")),
			),
			slog.String("error", err.Error()),
			slog.String("trace", string(debug.Stack())),
		)
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"something went wrong, please try again later",
		)
	}
	return nil
}
