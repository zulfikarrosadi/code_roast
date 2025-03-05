package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zulfikarrosadi/code_roast/internal/auth"
	"github.com/zulfikarrosadi/code_roast/internal/moderator"
	"github.com/zulfikarrosadi/code_roast/internal/post"
	"github.com/zulfikarrosadi/code_roast/internal/subforum"
	"github.com/zulfikarrosadi/code_roast/internal/user"
)

type Error struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  Error  `json:"error"`
}

const (
	CLOUDINARY_API_KEY    = "CLOUDINARY_API_KEY"
	CLOUDINARY_API_SECRET = "CLOUDINARY_API_SECRET"
	CLOUDINARY_CLOUD_NAME = "dxz9dwknn"
)

func loadEnv() {
	if os.Getenv("CI") == "true" {
		return
	}
	if os.Getenv("environment") == "production" {
		return
	}
	err := godotenv.Load("../../config/.env")
	if err != nil {
		panic(err)
	}
}

func main() {
	e := echo.New()
	loadEnv()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	db := OpenDBConnection(logger)
	if db == nil {
		panic("db connection fail to open")
	}

	e.Use(middleware.RequestID())

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:    true,
		LogRequestID: true,
		LogMethod:    true,
		LogLatency:   true,
		LogURIPath:   true,
		LogUserAgent: true,
		LogHeaders:   []string{"Authorization"},
		LogRemoteIP:  true,
		LogError:     true,
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			requestDetails := slog.Group("request",
				slog.String("id", v.RequestID),
				slog.String("method", v.Method),
				slog.String("path", v.URIPath),
				slog.String("user_agent", v.UserAgent),
				slog.String("ip", v.RemoteIP),
				slog.Any("authorization", v.Headers),
			)

			if v.Error != nil {
				// Differentiate user-caused errors (4xx) and server errors (5xx)
				var echoErrorRequest *echo.HTTPError
				if errors.As(v.Error, &echoErrorRequest) && v.Status >= 400 && v.Status < 500 {
					logger.LogAttrs(c.Request().Context(), slog.LevelWarn, "REQUEST_ERROR",
						slog.Int("status", v.Status),
						slog.Int("latency_ms", int(v.Latency)),
						requestDetails,
						slog.String("error", echoErrorRequest.Message.(string)),
					)
				} else {
					logger.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST_ERROR",
						slog.Int("status", v.Status),
						slog.Int("latency_ms", int(v.Latency)),
						requestDetails,
						slog.String("error", v.Error.Error()),
					)
				}
			} else {
				// Log successful requests as INFO
				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "REQUEST",
					slog.Int("status", v.Status),
					slog.Int("latency_ms", int(v.Latency)),
					requestDetails,
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Secure())
	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(os.Getenv("JWT_SECRETS")),
		SigningMethod: echojwt.AlgorithmHS256,
		Skipper: func(c echo.Context) bool {
			fmt.Println(c.Path())
			if c.Path() == "/api/v1/signin" || c.Path() == "/api/v1/signup" || c.Path() == "/api/v1/refresh" {
				return true
			}
			return false
		},
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &auth.CustomJWTClaims{}
		},
		ErrorHandler: func(c echo.Context, err error) error {
			if errors.Is(err, jwt.ErrTokenExpired) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Access token expired")
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				return echo.NewHTTPError(http.StatusBadRequest, "Malformed access token")
			}
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing access token")
		},
	}))
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}
		report, ok := err.(*echo.HTTPError)
		var errResponse ErrorResponse

		if ok {
			// Client error (4xx) or known server errors (5xx)
			errResponse = ErrorResponse{
				Status: "fail",
				Code:   report.Code,
				Error: Error{
					// need type asserition because it's interface{}
					Message: report.Message.(string),
				},
			}
		} else {
			errResponse = ErrorResponse{
				Status: "fail",
				Code:   http.StatusInternalServerError,
				Error: Error{
					Message: "something went wrong, please try again later",
				},
			}
		}

		if err := c.JSON(errResponse.Code, errResponse); err != nil {
			c.Logger().Error("FAILED_TO_SEND_ERROR_RESPONSE", slog.Any("error", err))
		}
	}

	cld, err := cloudinary.NewFromParams(
		CLOUDINARY_CLOUD_NAME,
		os.Getenv(CLOUDINARY_API_KEY),
		os.Getenv(CLOUDINARY_API_SECRET),
	)
	if err != nil {
		panic("cloudnary fail to initiate")
	}
	v := validator.New()
	userRepository := auth.NewUserRepository(logger, db)
	userService := auth.NewUserService(userRepository, v)
	userApi := auth.NewApiHandler(logger, userService)

	subforumRepository := subforum.NewRepository(db)
	subforumService := subforum.NewService(subforumRepository, v, cld)
	subforumApi := subforum.NewApi(subforumService, logger)

	postRepository := post.NewRepository(db)
	postService := post.NewService(postRepository, v, cld)
	postApi := post.NewApi(postService, logger)

	moderatorRepository := moderator.NewRepository(db)
	moderatorService := moderator.NewService(moderatorRepository, v)
	moderatorApi := moderator.NewApi(moderatorService, logger)

	r := e.Group("/api/v1")
	r.POST("/signup", userApi.Register)
	r.POST("/signin", userApi.Login)
	r.GET("/refresh", userApi.RefreshToken)
	r.POST("/subforums", subforumApi.Create, roles([]int{user.ROLE_ID_CREATE_SUBFORUM}))
	r.POST("/posts", postApi.Create)
	r.POST("/posts/:id/likes", postApi.Like)
	r.PUT("/moderators/posts/:postId/status", postApi.TakeDown, roles([]int{user.ROLE_ID_TAKE_DOWN_POST}))
	r.POST("/moderators", moderatorApi.AddRoles)

	e.Start("localhost:3000")
}

func roles(requiredRoles []int) func(echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Get("user").(*jwt.Token)

			claims, ok := token.Claims.(*auth.CustomJWTClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "you don't have perimission to do this operation")
			}

			roleSet := make(map[int]bool)
			for _, role := range claims.Roles {
				roleSet[role.Id] = true
			}
			for _, reqRole := range requiredRoles {
				if !roleSet[reqRole] {
					return echo.NewHTTPError(http.StatusForbidden, "you don't have permission to do this operation")
				}
			}
			return next(c)
		}
	}
}

func OpenDBConnection(logger *slog.Logger) *sql.DB {
	db, err := sql.Open("mysql", os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		logger.LogAttrs(context.Background(), slog.LevelError, "DB_CONNECTION_ERR", slog.Any("details", err))
		return nil
	}
	return db
}
