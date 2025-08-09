package post

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	apperror "github.com/zulfikarrosadi/code_roast/internal/app-error"
	"github.com/zulfikarrosadi/code_roast/internal/auth"
	"github.com/zulfikarrosadi/code_roast/internal/middleware"
	"github.com/zulfikarrosadi/code_roast/pkg/schema"
)

type service interface {
	create(context.Context, createRequest) (schema.Response[createResponse], error)
	takeDown(context.Context, string, sql.NullInt64) (schema.Response[createResponse], error)
	like(context.Context, likeCreateRequest) (schema.Response[likeResponse], error)
	getAll(ctx context.Context) (schema.Response[getAllResponse], error)
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
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	token := c.Get("user").(*jwt.Token)
	user, ok := token.Claims.(*auth.CustomJWTClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please use correct user credential and try again later")
	}

	newPost := createRequest{}
	media, err := c.MultipartForm()
	if err != nil {
		logger.Debug("failed to bind data user request", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "fail to process your request, failed to open post media files")
	}

	newPost.Media = media.File["post_media"]
	newPost.userId = user.Id
	newPost.SubforumId = c.FormValue("subforum_id")
	newPost.Caption = c.FormValue("caption")
	response, err := api.service.create(ctx, newPost)
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
		logger.Debug("create post service fail", slog.Any("error", err))
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}
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

func (api *ApiImpl) Like(c echo.Context) error {
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	user, err := auth.GetUserFromContext(c)
	if err != nil {
		return echo.ErrUnauthorized
	}
	data := likeCreateRequest{}
	if err = c.Bind(&data); err != nil {
		logger.Debug("failed to bind data user request", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusBadRequest, "failed to like this post. Send correct information and please try again later")
	}
	data.UserId = user.Id

	response, err := api.service.like(ctx, data)
	if err != nil {
		logger.Debug("like post service failed", slog.String("error", err.Error()))
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}
	err = c.JSON(response.Code, response)
	if err != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
	}
	return nil
}

func (api *ApiImpl) TakeDown(c echo.Context) error {
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)

	postId := c.Param("postId")
	response, err := api.service.takeDown(ctx, postId, sql.NullInt64{
		Int64: time.Now().Unix(),
	})
	if err != nil {
		logger.Debug("take down service fail", slog.String("error", err.Error()))
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			"something went wrong, please try again later",
		)
	}
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

func (api *ApiImpl) GetAll(c echo.Context) error {
	ctx := c.Request().Context()
	logger := middleware.GetLogger(ctx)
	response, err := api.service.getAll(ctx)
	if err != nil {
		logger.Debug("get all post service fail", slog.String("error", err.Error()))
		return echo.NewHTTPError(response.Code, response.Error.Message)
	}

	err = c.JSON(response.Code, response)
	if err != nil {
		logger.Error("failed to write JSON response", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong please try again later")
	}
	return nil
}
