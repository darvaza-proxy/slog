package internal

import (
	"fmt"

	"github.com/darvaza-proxy/core"
)

var (
	_ core.Recovered = (*PanicError)(nil)
)

// PanicError is an error to be sent via panic, ideally
// to be caught using slog.Recover()
type PanicError struct {
	payload any
	stack   core.Stack
}

// Error returns the payload as a string
func (p *PanicError) Error() string {
	return fmt.Sprintf("panic: %s", p.payload)
}

// Unwrap returns the payload if it's and error
func (p *PanicError) Unwrap() error {
	if err, ok := p.payload.(error); ok {
		return err
	}
	return nil
}

// Recovered returns the payload of the panic
func (p *PanicError) Recovered() any {
	return p.payload
}

// Stack returns the call stack associated to this panic() event
func (p *PanicError) Stack() core.Stack {
	return p.stack
}

// NewPanicError creates a new PanicError with arbitrary payload
func NewPanicError(skip int, payload any) *PanicError {
	return &PanicError{
		payload: payload,
		stack:   core.StackTrace(skip + 1),
	}
}

// NewPanicErrorf creates a new PanicError with a formated string as payload
func NewPanicErrorf(skip int, format string, args ...any) *PanicError {
	return &PanicError{
		payload: fmt.Errorf(format, args...),
		stack:   core.StackTrace(skip + 1),
	}
}

// NewPanicWrapf creates a new PanicError wrapping a given error as part of the payload
func NewPanicWrapf(skip int, err error, format string, args ...any) *PanicError {
	return &PanicError{
		payload: core.Wrapf(err, format, args...),
		stack:   core.StackTrace(skip + 1),
	}
}
