package subforum

import (
	"context"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	apperror "github.com/zulfikarrosadi/code_roast/internal/app-error"
	"github.com/zulfikarrosadi/code_roast/internal/auth"
	"github.com/zulfikarrosadi/code_roast/internal/middleware"
	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type service interface {
	create(context.Context, subforumCreateRequest) (schema.Response[createResponse], error)
	getAll(context.Context) (schema.Response[getAllResponse], error)
}

type ApiImpl struct {
	service service
	*slog.Logger
}

type subforumCreateRequest struct {
	UserId      string
	Name        string                `validate:"required"`
	Description string                `validate:"required"`
	Icon        *multipart.FileHeader `validate:"required"`
	Banner      *multipart.FileHeader `validate:"required"`
}

func NewApi(service service, logger *slog.Logger) *ApiImpl {
	return &ApiImpl{
		service: service,
		Logger:  logger,
	}
}

const (
	REQUEST_ID_KEY = "REQUEST_ID"
)

func (api *ApiImpl) Create(c echo.Context) error {
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	token := c.Get("user").(*jwt.Token)
	user, ok := token.Claims.(*auth.CustomJWTClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please use correct user credential and try again later")
	}

	newSubforum := subforumCreateRequest{}
	newSubforum.UserId = user.Id
	newSubforum.Name = c.FormValue("name")
	newSubforum.Description = c.FormValue("description")
	subForumIcon, err := c.FormFile("icon")
	if err != nil {
		logger.Debug("failed to create sub forum, missing icon file", slog.String("error", err.Error()))
		if err == http.ErrMissingFile {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to create new sub, forum. missing icon file")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "something went wrong, enter correct information and please try again")
	}
	subForumBanner, err := c.FormFile("banner")
	if err != nil {
		logger.Debug("failed to create sub forum, missing banner file", slog.String("error", err.Error()))
		if err == http.ErrMissingFile {
			return echo.NewHTTPError(http.StatusBadRequest, "failed to create new sub, forum. missing banner file")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "something went wrong, enter correct information and please try again")
	}
	newSubforum.UserId = user.Id
	newSubforum.Icon = subForumIcon
	newSubforum.Banner = subForumBanner

	response, err := api.service.create(c.Request().Context(), newSubforum)
	if err != nil {
		if response.Error.Message == apperror.VALIDATION_ERROR {
			logger.Debug("request validation fail", slog.Any("error", err))
			if c.JSON(http.StatusBadRequest, response) != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
			}
			return nil
		}
		logger.Debug("create subforum service fail", slog.Any("error", err))
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}
	err = c.JSON(response.Code, response)
	if err != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
	}

	return nil
}

func (api *ApiImpl) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	result, err := api.service.getAll(ctx)
	if err != nil {
		logger.Debug("get all subforum service fail", slog.Any("error", err))
		if err := c.JSON(result.Code, result); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
		}
	}
	if err := c.JSON(result.Code, result); err != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
	}
	return nil
}
