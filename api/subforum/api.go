package subforum

import (
	"context"
	"log/slog"
	"mime/multipart"
	"net/http"
	"runtime/debug"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	apperror "github.com/zulfikarrosadi/code_roast/app-error"
	imagehelper "github.com/zulfikarrosadi/code_roast/image-helper"
	"github.com/zulfikarrosadi/code_roast/schema"
	"github.com/zulfikarrosadi/code_roast/user"
)

type service interface {
	create(context.Context, subforumCreateRequest) (schema.Response[subforumResponse], error)
}

type ApiImpl struct {
	service service
	cld     *cloudinary.Cloudinary
	*slog.Logger
}

type subforumCreateRequest struct {
	UserId      string
	Name        string                `form:"name" validate:"required"`
	Description string                `form:"description" validate:"required"`
	Icon        *multipart.FileHeader `form:"icon" validate:"required"`
	Banner      *multipart.FileHeader `form:"banner" validate:"required"`
}

func NewApi(service service, cld *cloudinary.Cloudinary, logger *slog.Logger) *ApiImpl {
	return &ApiImpl{
		service: service,
		cld:     cld,
		Logger:  logger,
	}
}

const (
	REQUEST_ID_KEY = "REQUEST_ID"
)

func (api *ApiImpl) Create(c echo.Context) error {
	ctx := context.WithValue(context.TODO(), REQUEST_ID_KEY, c.Response().Header().Get(echo.HeaderXRequestID))
	token := c.Get("user").(*jwt.Token)
	user, ok := token.Claims.(*user.CustomJWTClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Please use correct user credential and try again later")
	}

	newSubforum := subforumCreateRequest{}
	err := c.Bind(&newSubforum)
	if err != nil {
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
		return echo.NewHTTPError(http.StatusBadRequest, "fail to create new subforum, bind error")
	}
	newSubforum.UserId = user.Id
	response, err := api.service.create(c.Request().Context(), newSubforum)
	if err != nil {
		if response.Error.Message == apperror.VALIDATION_ERROR {
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
			if c.JSON(http.StatusBadRequest, response) != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "something went wrong, please try again later")
			}
			return nil
		}
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
	err = c.JSON(response.Code, response)
	if err != nil {
		api.Logger.Debug("REQUEST_DEBUG",
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
