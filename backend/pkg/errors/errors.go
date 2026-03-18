package errors

import (
	"errors"
	"fmt"
)

// Sentinel domain errors — map these at the infra boundary.
var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrForbidden  = errors.New("forbidden")
	ErrBadRequest = errors.New("bad request")
)

// DomainError wraps a sentinel with a human-readable message.
type DomainError struct {
	Sentinel error
	Message  string
}

func (e *DomainError) Error() string { return e.Message }
func (e *DomainError) Unwrap() error { return e.Sentinel }

func NotFound(format string, args ...any) error {
	return &DomainError{Sentinel: ErrNotFound, Message: fmt.Sprintf(format, args...)}
}

func Conflict(format string, args ...any) error {
	return &DomainError{Sentinel: ErrConflict, Message: fmt.Sprintf(format, args...)}
}

func Forbidden(format string, args ...any) error {
	return &DomainError{Sentinel: ErrForbidden, Message: fmt.Sprintf(format, args...)}
}

func BadRequest(format string, args ...any) error {
	return &DomainError{Sentinel: ErrBadRequest, Message: fmt.Sprintf(format, args...)}
}

// Is delegates to stdlib so errors.Is works transparently.
var Is = errors.Is
