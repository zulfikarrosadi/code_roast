package user

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/zulfikarrosadi/code_roast/internal/middleware"
	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type service interface {
	findById(ctx context.Context, id string) (schema.Response[FindByIdResponse], error)
}

type apiImpl struct {
	service
}

func NewApiHandler(service service) apiImpl {
	return apiImpl{
		service: service,
	}
}

const REQUEST_ID_KEY = "REQUEST_ID_KEY"

func (api apiImpl) FindById(c echo.Context) error {
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	response, err := api.service.findById(ctx, c.Param("id"))
	if err != nil {
		logger.Debug("find user by id service fail", slog.String("error", err.Error()))
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}

	err = c.JSON(response.Code, response)
	if err != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
		return echo.NewHTTPError(500, "something went wrong, please try again later")
	}
	return nil
}
