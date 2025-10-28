package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kitakitabauer/gin-sample-app/config"
	"github.com/kitakitabauer/gin-sample-app/handler"
	"github.com/kitakitabauer/gin-sample-app/internal/middleware"
	"github.com/kitakitabauer/gin-sample-app/logger"
	"github.com/kitakitabauer/gin-sample-app/repository"
	"github.com/kitakitabauer/gin-sample-app/service"
)

func New() (*gin.Engine, error) {
	if err := logger.Init(config.AppConfig.Env); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.GinZap())

	postRepository := repository.NewInMemoryPostRepository()
	postService := service.NewPostService(postRepository)
	postHandler := handler.NewPostHandler(postService)
	postHandler.RegisterRoutes(r)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r, nil
}
