package server

import (
	"database/sql"
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

func New(db *sql.DB) (*gin.Engine, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if logger.Log == nil {
		return nil, fmt.Errorf("logger is not initialised")
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.GinZap())

	postRepository := repository.NewSQLPostRepository(db, config.AppConfig.DatabaseDriver)
	postService := service.NewPostService(postRepository)
	postHandler := handler.NewPostHandler(postService)
	postHandler.RegisterRoutes(r)

	adminHandler := handler.NewAdminHandler()
	adminHandler.RegisterRoutes(r)

	docsHandler := handler.NewDocsHandler()
	docsHandler.RegisterRoutes(r)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r, nil
}
