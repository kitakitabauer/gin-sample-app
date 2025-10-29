package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kitakitabauer/gin-sample-app/internal/middleware"
	"github.com/kitakitabauer/gin-sample-app/logger"
	"go.uber.org/zap"
)

type AdminHandler struct{}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) RegisterRoutes(router *gin.Engine) {
	admin := router.Group("/admin")
	admin.Use(middleware.RequireAPIKey())

	admin.GET("/log-level", h.getLogLevel)
	admin.PUT("/log-level", h.updateLogLevel)
}

func (h *AdminHandler) getLogLevel(c *gin.Context) {
	level, err := logger.CurrentLevel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"level": level.String()})
}

type updateLogLevelRequest struct {
	Level string `json:"level" binding:"required"`
}

func (h *AdminHandler) updateLogLevel(c *gin.Context) {
	var req updateLogLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldLevel, err := logger.CurrentLevel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := logger.SetLevel(req.Level); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newLevel, err := logger.CurrentLevel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.Log.Info("log level updated",
		zap.String("old_level", oldLevel.String()),
		zap.String("new_level", newLevel.String()),
		zap.String("requested_level", req.Level),
		zap.String("client_ip", c.ClientIP()),
		zap.Bool("api_key_present", c.GetHeader("X-API-Key") != ""),
		zap.String("user_agent", c.Request.UserAgent()),
		zap.Time("timestamp", time.Now().UTC()),
	)

	c.JSON(http.StatusOK, gin.H{"level": newLevel.String()})
}
