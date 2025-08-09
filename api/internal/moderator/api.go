package moderator

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	apperror "github.com/zulfikarrosadi/code_roast/internal/app-error"
	"github.com/zulfikarrosadi/code_roast/internal/middleware"

	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type service interface {
	addRoles(context.Context, updateRoleRequest) (schema.Response[UpdatePermissionResponse], error)
	removeRoles(context.Context, updateRoleRequest) (schema.Response[UpdatePermissionResponse], error)
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
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	data := updateRoleRequest{}
	err := c.Bind(&data)
	if err != nil {
		logger.Debug("failed to bind data user request", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusBadRequest, "fail to process your request, send correct data and try again")
	}

	response, err := api.service.addRoles(ctx, data)
	if err != nil {
		if response.Error.Message == apperror.VALIDATION_ERROR {
			logger.Debug("request validation fail", slog.Any("error", err))
			err = c.JSON(response.Code, response)
			if err != nil {
				logger.Error("failed to write JSON response", slog.String("error", err.Error()))
				return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
			}

			logger.Debug("add role service fail", slog.Any("error", err))
			return nil
		}
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}
	if err := c.JSON(response.Code, response); err != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"something went wrong, please try again later",
		)
	}
	return nil
}
