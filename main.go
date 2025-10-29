package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kitakitabauer/gin-sample-app/config"
	"github.com/kitakitabauer/gin-sample-app/internal/database"
	"github.com/kitakitabauer/gin-sample-app/internal/server"
	"github.com/kitakitabauer/gin-sample-app/logger"
	"go.uber.org/zap"
)

func main() {
	config.Load()

	if err := logger.Init(logger.Config{
		Env:     config.AppConfig.Env,
		Level:   config.AppConfig.LogLevel,
		Service: "gin-sample-app",
	}); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, err := database.Open(dbCtx, database.Config{
		Driver: config.AppConfig.DatabaseDriver,
		DSN:    config.AppConfig.DatabaseDSN,
	})
	cancel()
	if err != nil {
		logger.Log.Fatal("failed to connect database", zap.Error(err))
	}
	defer db.Close()

	if err := database.MigrateUp(db, config.AppConfig.DatabaseDriver); err != nil {
		logger.Log.Fatal("failed to apply migrations", zap.Error(err))
	}

	r, err := server.New(db)
	if err != nil {
		logger.Log.Fatal("failed to create server", zap.Error(err))
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.AppConfig.Port),
		Handler: r,
	}

	go func() {
		logger.Log.Info("starting server",
			zap.String("addr", srv.Addr),
			zap.String("env", config.AppConfig.Env),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("server exited gracefully")
}
