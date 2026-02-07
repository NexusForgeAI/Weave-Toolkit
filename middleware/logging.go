package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"Weave-Toolkit/internal/logger"
)

// LoggingMiddleware 日志中间件
func LoggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// 处理请求
		c.Next()

		// 记录日志
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		log.Info().
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Str("client_ip", clientIP).
			Str("user_agent", userAgent).
			Dur("latency", latency).
			Msg("HTTP request")
	}
}