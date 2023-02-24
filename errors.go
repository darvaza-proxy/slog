package slog

import (
	"fmt"
)

// Wrapf annotates an error with a formated string. if %w is used the argument
// will be unwrapped
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	return Wrap(err, fmt.Errorf(format, args...).Error())
}

// Wrap annotates an error with a string
func Wrap(err error, note string) error {
	if err == nil {
		return nil
	}

	if len(note) == 0 {
		return err
	}

	return &WrappedError{
		cause: err,
		note:  note,
	}
}

// WrappedError is an annotated error that can be Unwrapped
type WrappedError struct {
	cause error
	note  string
}

func (w *WrappedError) Error() string {
	s := w.cause.Error()
	if len(s) == 0 {
		return w.note
	}

	return fmt.Sprintf("%s: %s", w.note, s)
}

func (w *WrappedError) Unwrap() error {
	return w.cause
}
