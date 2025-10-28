package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kitakitabauer/gin-sample-app/logger"
	"go.uber.org/zap"
)

func GinZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		logger.Log.Info("request completed",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
		)
	}
}
