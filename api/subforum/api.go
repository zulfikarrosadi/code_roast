package subforum

import (
	"context"
	"fmt"
	"log/slog"
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
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Icon        string `json:"icon" validate:"required"`
	Banner      string `json:"banner" validate:"required"`
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
	newSubforum.UserId = user.Id
	newSubforum.Name = c.FormValue("name")
	newSubforum.Description = c.FormValue("description")
	subForumIcon, err := c.FormFile("icon")
	if err != nil {
		fmt.Println(err)
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
		return echo.NewHTTPError(http.StatusBadRequest, "something went wrong, enter correct information and please try again")
	}
	subForumBanner, err := c.FormFile("banner")
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
		return echo.NewHTTPError(http.StatusBadRequest, "something went wrong, enter correct information and please try again")
	}
	iconSrc, err := subForumIcon.Open()
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
	defer iconSrc.Close()
	if _, err := imagehelper.IsImage(iconSrc); err != nil {
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
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported icon file type, only upload jpg or png file")
	}

	subForumIconUpload, err := api.cld.Upload.Upload(
		c.Request().Context(),
		iconSrc,
		uploader.UploadParams{
			ResourceType: "image",
		},
	)
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
	bannerSrc, err := subForumBanner.Open()
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
	defer bannerSrc.Close()
	if _, err := imagehelper.IsImage(bannerSrc); err != nil {
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
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported banner file type, only upload jpg or png file")
	}

	subForumBannerUpload, err := api.cld.Upload.Upload(
		c.Request().Context(),
		bannerSrc,
		uploader.UploadParams{
			ResourceType: "image",
		},
	)
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
	newSubforum.Icon = subForumIconUpload.SecureURL
	newSubforum.Banner = subForumBannerUpload.SecureURL

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
