package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
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
			c.Set("requestId", v.RequestID)
			if v.Error != nil {
				logger.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST_ERROR",
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

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "hello world")
	})
	e.Start("localhost:3000")
}
