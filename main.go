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
	"github.com/kitakitabauer/gin-sample-app/internal/server"
	"github.com/kitakitabauer/gin-sample-app/logger"
	"go.uber.org/zap"
)

func main() {
	config.Load()
	r, err := server.New()
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	defer logger.Log.Sync()

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
