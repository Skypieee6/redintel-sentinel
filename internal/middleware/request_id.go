// Package middleware contains Gin HTTP middleware.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader is the header used to propagate the request correlation ID.
const RequestIDHeader = "X-Request-ID"

const requestIDKey = "request_id"

// RequestID ensures every request has a correlation ID. An inbound
// X-Request-ID header is honored; otherwise a new UUID is generated. The value
// is stored in the context and echoed back in the response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(requestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// GetRequestID returns the correlation ID associated with the request.
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(requestIDKey); ok {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}
