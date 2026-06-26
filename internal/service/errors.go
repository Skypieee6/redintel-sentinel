// Package service holds the business logic, sitting between HTTP handlers and
// the repository layer.
package service

import (
	"errors"
	"fmt"
)

// Sentinel error kinds. Handlers map these to HTTP status codes via errors.Is.
var (
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("already exists")
	ErrValidation         = errors.New("validation failed")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrForbidden          = errors.New("forbidden")
	ErrInactive           = errors.New("account is inactive")
	ErrRateLimited        = errors.New("too many attempts")
)

// Error wraps a sentinel kind with a human-readable message.
type Error struct {
	Kind error
	Msg  string
}

func (e *Error) Error() string { return e.Msg }
func (e *Error) Unwrap() error { return e.Kind }

// wrap creates an *Error for a kind with a formatted message.
func wrap(kind error, format string, args ...any) *Error {
	return &Error{Kind: kind, Msg: fmt.Sprintf(format, args...)}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
