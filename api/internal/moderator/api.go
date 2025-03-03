package moderator

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	apperror "github.com/zulfikarrosadi/code_roast/internal/app-error"

	"github.com/zulfikarrosadi/code_roast/internal/auth"
	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type service interface {
	addRoles(context.Context, string, updateRoleRequest) (schema.Response[UpdatePermissionResponse], error)
	removeRoles(context.Context, string, updateRoleRequest) (schema.Response[UpdatePermissionResponse], error)
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

func (api *ApiImpl) AddRoles(c echo.Context) error {
	ctx := context.WithValue(context.TODO(), REQUEST_ID_KEY, c.Response().Header().Get(echo.HeaderXRequestID))
	roles := updateRoleRequest{}
	err := c.Bind(&roles.RoleId)
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
		return echo.NewHTTPError(http.StatusBadRequest, "fail to process your request, send correct data and try again")
	}

	claims, err := auth.GetUserFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "you don't have perimission to do this operation")
	}

	response, err := api.service.addRoles(ctx, claims.Id, roles)
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
	if err := c.JSON(response.Code, response); err != nil {
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
