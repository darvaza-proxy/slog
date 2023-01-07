package slog

import (
	"bytes"
	"io"
)

var (
	_ io.Writer = (*LogWriter)(nil)
)

type LogWriterFunc func(Logger, string) error

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

	return
}

func defaultLogWriter(l Logger, s string) error {
	l.Printf("%s", s)
	return nil
}

func NewLogWriter(l Logger, fn LogWriterFunc) *LogWriter {
	if l == nil {
		return nil
	}
	if fn == nil {
		fn = defaultLogWriter
	}
	return &LogWriter{l, fn}
}
