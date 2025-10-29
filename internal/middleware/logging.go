package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kitakitabauer/gin-sample-app/logger"
	"go.uber.org/zap"
)

func GinZap() gin.HandlerFunc {
	return func(c *gin.Context) {
		if logger.Log == nil {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		query := c.Request.URL.RawQuery

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("bytes_sent", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		logger.Log.Info("request completed", fields...)
	}
}
