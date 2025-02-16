package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/zulfikarrosadi/code_roast/user"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  Error  `json:"error"`
}

func main() {
	e := echo.New()
	err := godotenv.Load()
	if err != nil {
		panic("env not loaded")
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
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
		LogRemoteIP:  true,
		LogError:     true,
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "REQUEST_ERROR",
					slog.Int("status", v.Status),
					slog.String("path", v.URIPath),
					slog.String("id", v.RequestID),
					slog.String("method", v.Method),
					slog.Int("latency", int(v.Latency)),
					slog.String("agent", v.UserAgent),
					slog.String("ip", v.RemoteIP),
					slog.String("error", v.Error.Error()),
				)
			} else {
				logger.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST",
					slog.Int("status", v.Status),
					slog.String("path", v.URIPath),
					slog.String("id", v.RequestID),
					slog.String("method", v.Method),
					slog.Int("latency", int(v.Latency)),
					slog.String("agent", v.UserAgent),
					slog.String("ip", v.RemoteIP),
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
			if c.Path() == "/api/v1/signin" || c.Path() == "/api/v1/signup" {
				return true
			}
			return false
		},
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &user.CustomJWTClaims{}
		},
		SuccessHandler: func(c echo.Context) {
			token := c.Get("user").(*jwt.Token)
			claims, ok := token.Claims.(*user.CustomJWTClaims)
			if !ok {
				// Log the error and skip further processing
				slog.LogAttrs(c.Request().Context(),
					slog.LevelError,
					"REQUEST_ERROR",
					slog.Group("details",
						slog.String("message", "Failed to assert claims"),
						slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
					))
				return
			}

			slog.LogAttrs(c.Request().Context(),
				slog.LevelInfo,
				"INFO",
				slog.Group("details",
					slog.String("message", "access token successfully verified"),
					slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
					slog.String("user_id", claims.Id),
				))
		},
		ErrorHandler: func(c echo.Context, err error) error {
			if errors.Is(err, jwt.ErrTokenExpired) {
				slog.LogAttrs(c.Request().Context(),
					slog.LevelError,
					"REQUEST_ERROR",
					slog.Group("details",
						slog.String("message", "Access token expired"),
						slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
					))
				return echo.NewHTTPError(http.StatusUnauthorized, "Access token expired")
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				slog.LogAttrs(c.Request().Context(),
					slog.LevelError,
					"REQUEST_ERROR",
					slog.Group("details",
						slog.String("message", "Malformed access token"),
						slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
					))
				return echo.NewHTTPError(http.StatusBadRequest, "Malformed access token")
			}
			authHeader := c.Request().Header.Get("Authorization")
			slog.LogAttrs(c.Request().Context(),
				slog.LevelError,
				"REQUEST_ERROR",
				slog.Group("details",
					slog.String("message", err.Error()),
					slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
					slog.String("authorization_header", authHeader),
				))
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
			code := report.Code
			message := report.Message

			errResponse = ErrorResponse{
				Status: "fail",
				Error: Error{
					Code:    code,
					Message: message.(string), // Type assertion since report.Message is interface{}
				},
			}
		} else {
			errResponse = ErrorResponse{
				Status: "fail",
				Error: Error{
					Code:    http.StatusInternalServerError,
					Message: "something went wrong, please try again later",
				},
			}
		}
		if err := c.JSON(errResponse.Error.Code, errResponse); err != nil {
			c.Logger().Error("Error sending error response:", err)
		}
	}

	userRepository := user.NewUserRepository(logger, db)
	userService := user.NewUserService(logger, userRepository)
	userApi := user.NewApiHandler(logger, userService)

	r := e.Group("/api/v1")
	r.POST("/signup", userApi.Register)
	r.POST("/signin", userApi.Login)
	r.GET("/", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)

		// Extract the claims from the token
		claims, ok := token.Claims.(*user.CustomJWTClaims)
		if !ok {
			fmt.Println("token: ", token)
		}
		fmt.Println(claims.Email)
		err := c.String(http.StatusOK, "ok")
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "error")
		}
		return nil
	})
	e.Start("localhost:3000")
}

func OpenDBConnection(logger *slog.Logger) *sql.DB {
	db, err := sql.Open("mysql", os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		logger.LogAttrs(context.Background(), slog.LevelError, "DB_CONNECTION_ERR", slog.Any("details", err))
		return nil
	}
	return db
}
