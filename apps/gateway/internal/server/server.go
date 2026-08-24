// Package server builds and runs the gateway's Echo HTTP server.
package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/handler"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/readmodel"
	"github.com/mralaminahamed/sitemon/apps/gateway/internal/service"
	"github.com/mralaminahamed/sitemon/packages/shared/logger"
)

// New builds a configured Echo instance with middleware and routes.
func New(svc *service.Service, rm *readmodel.ReadModel) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	// Open CORS for the React app in dev; tighten per-origin in Phase 7.
	e.Use(middleware.CORS())

	h := handler.New(svc, rm)

	e.GET("/", h.Root)
	e.GET("/health", h.Health)

	api := e.Group("/api")
	api.GET("/status", h.Status)
	api.GET("/history", h.History)
	api.GET("/stats", h.Stats)
	api.POST("/checks", h.Checks)
	api.GET("/ssl", h.SSL)
	api.POST("/loadtest", h.LoadTest)

	return e
}

// Run starts the server and blocks until SIGINT/SIGTERM, then shuts down
// gracefully.
func Run(svc *service.Service, rm *readmodel.ReadModel, addr string) error {
	e := New(svc, rm)

	go func() {
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error().Err(err).Msg("gateway server stopped")
		}
	}()
	logger.Log.Info().Str("addr", addr).Msg("gateway listening")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Log.Info().Msg("gateway shutting down")
	return e.Shutdown(ctx)
}
