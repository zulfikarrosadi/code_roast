package post

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/labstack/echo/v4"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
	"github.com/zulfikarrosadi/code_roast/schema"
)

type service interface {
	create(context.Context, postCreateRequest) (schema.Response[postResponse], error)
	takeDown(context.Context, string, sql.NullInt64) (schema.Response[postResponse], error)
}

type ApiImpl struct {
	service
	*slog.Logger
}

func NewApi(service service, logger *slog.Logger) *ApiImpl {
	return &ApiImpl{
		service: service,
		Logger:  logger,
	}
}

type REQUEST_ID string

var (
	REQUEST_ID_KEY REQUEST_ID = "REQUEST_ID"
)

func (api *ApiImpl) Create(c echo.Context) error {
	ctx := context.WithValue(context.TODO(), REQUEST_ID_KEY, c.Response().Header().Get(echo.HeaderXRequestID))

	newPost := postCreateRequest{}
	if err := c.Bind(&newPost); err != nil {
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
		return echo.NewHTTPError(http.StatusBadRequest, "fail to process your request, send correct data and try again")
	}

	response, err := api.service.create(ctx, newPost)
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

func (api *ApiImpl) TakeDown(c echo.Context) error {
	ctx := context.WithValue(context.TODO(), REQUEST_ID_KEY, c.Response().Header().Get(echo.HeaderXRequestID))

	postId := c.Param("postId")
	response, err := api.service.takeDown(ctx, postId, sql.NullInt64{
		Int64: time.Now().Unix(),
	})
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
