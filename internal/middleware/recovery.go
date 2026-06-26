package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Skypieee6/redintel-sentinel/pkg/response"
)

// Recovery returns middleware that recovers from panics, logs the stack via Zap
// and responds with a standardized 500 error instead of crashing the server.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic recovered",
					zap.String("request_id", GetRequestID(c)),
					zap.String("path", c.Request.URL.Path),
					zap.Any("error", err),
					zap.Stack("stack"),
				)
				if !c.Writer.Written() {
					response.Error(c, http.StatusInternalServerError,
						"internal_error", "an unexpected error occurred")
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
