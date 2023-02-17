package slog

import (
	"bytes"
	"io"
)

var (
	_ io.Writer = (*LogWriter)(nil)
)

// LogWriterFunc is the prototype of the functions
// used to process and log Write() calls
type LogWriterFunc func(Logger, string) error

// LogWriter is a io.Writer that calls a given function
// to log each Write() call
type LogWriter struct {
	l  Logger
	fn LogWriterFunc
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	// handler
	fn := w.fn
	if fn == nil {
		fn = defaultLogWriter
	}

	// remove trailing \n
	n = len(p)
	p = bytes.TrimRight(p, "\n")

	// call handler
	if err = fn(w.l, string(p)); err != nil {
		n = 0
	}

	return n, err
}

func defaultLogWriter(l Logger, s string) error {
	l.Print(s)
	return nil
}

// NewLogWriter creates a new LogWriter with the given slog.Logger
// and handler function
func NewLogWriter(l Logger, fn LogWriterFunc) *LogWriter {
	if l == nil {
		return nil
	}
	if fn == nil {
		fn = defaultLogWriter
	}
	return &LogWriter{l, fn}
}
