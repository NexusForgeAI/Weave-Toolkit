package middleware

import (
	"Weave-Toolkit/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 崩溃恢复中间件
func RecoveryMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				log.Error().
					Interface("panic", err).
					Str("path", c.Request.URL.Path).
					Msg("Recovered from panic")
			}
		}()
		c.Next()
	}
}
