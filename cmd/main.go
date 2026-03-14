package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CXTACLYSM/postgres-ha-practice/configs"
	"github.com/CXTACLYSM/postgres-ha-practice/internal"
	"github.com/CXTACLYSM/postgres-ha-practice/internal/di"
	"go.uber.org/zap"
)

func main() {
	cfg := configs.Create()

	container := &di.Container{}
	if err := container.Init(cfg); err != nil {
		log.Fatalf("error initializing container: %v", err)
	}
	logger := container.Infrastructure.Logger
	defer logger.Sync()
	defer container.Infrastructure.PgConnector.Close()

	r := internal.InitRouter(container.Handlers, container.Middlewares)
	srv := &http.Server{
		Addr:              cfg.App.HttpSocketStr(),
		Handler:           r,
		ReadHeaderTimeout: cfg.App.Http.ReadHeaderTimeout,
		ReadTimeout:       cfg.App.Http.ReadTimeout,
		WriteTimeout:      cfg.App.Http.WriteTimeout,
		IdleTimeout:       cfg.App.Http.IdleTimeout,
		MaxHeaderBytes:    cfg.App.Http.MaxHeaderBytes,
	}

	go func() {
		logger.Info("starting http server", zap.String("addr", cfg.App.HttpSocketStr()))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down server...")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}
	logger.Info("server stopped")
}
