// Package response provides standardized JSON response helpers for the API.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the standard success response wrapper.
type Envelope struct {
	Data interface{} `json:"data"`
}

// ErrorBody is the standard error payload.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorEnvelope is the standard error response wrapper.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// OK writes a 200 response wrapping data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Data: data})
}

// JSON writes an arbitrary status code wrapping data.
func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Envelope{Data: data})
}

// Error writes a standardized error response.
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorEnvelope{Error: ErrorBody{Code: code, Message: message}})
}
