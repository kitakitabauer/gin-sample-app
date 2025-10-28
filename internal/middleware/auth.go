package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kitakitabauer/gin-sample-app/config"
)

const apiKeyHeader = "X-API-Key"

func RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.AppConfig == nil {
			c.Next()
			return
		}

		expected := config.AppConfig.APIKey
		if expected == "" {
			c.Next()
			return
		}

		if provided := c.GetHeader(apiKeyHeader); provided != expected {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		c.Next()
	}
}
