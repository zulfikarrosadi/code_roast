package user

import (
	"context"

	"github.com/labstack/echo/v4"
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
	ctx := context.WithValue(context.TODO(), REQUEST_ID_KEY, c.Response().Header().Get(echo.HeaderXRequestID))

	response, err := api.service.findById(ctx, c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}

	err = c.JSON(response.Code, response)
	if err != nil {
		return echo.NewHTTPError(500, "something went wrong, please try again later")
	}
	return nil
}
