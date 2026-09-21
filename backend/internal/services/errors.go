package services

import (
	"errors"
	"fmt"
)

// Kind classifies a service error so the API layer can pick a status code
// without inspecting messages.
type Kind int

const (
	KindInvalid Kind = iota + 1
	KindUnauthorized
	KindForbidden
	KindNotFound
	KindConflict
)

// Error is a service-level failure with a caller-safe message.
type Error struct {
	Kind    Kind
	Message string
}

func (e *Error) Error() string { return e.Message }

// KindOf returns the Kind of err, or 0 when err is not a service Error.
func KindOf(err error) Kind {
	var se *Error
	if errors.As(err, &se) {
		return se.Kind
	}
	return 0
}

func invalidf(format string, args ...any) error {
	return &Error{Kind: KindInvalid, Message: fmt.Sprintf(format, args...)}
}

func unauthorized(message string) error {
	return &Error{Kind: KindUnauthorized, Message: message}
}

func forbidden(message string) error {
	return &Error{Kind: KindForbidden, Message: message}
}

func notFound(resource string) error {
	return &Error{Kind: KindNotFound, Message: resource + " not found"}
}

func conflict(message string) error {
	return &Error{Kind: KindConflict, Message: message}
}
