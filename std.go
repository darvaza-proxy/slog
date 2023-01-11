package slog

import (
	"log"
)

// NewStdLogger creates a standard *log.Logger using a slog.Logger
// behind the scenes
func NewStdLogger(l Logger, prefix string, flags int) *log.Logger {
	w := NewLogWriter(l, newStdLogSink(prefix, flags))

	return log.New(w, "", flags)
}

type stdLogSink struct {
	prefix string
	flags  int
}

func (w *stdLogSink) LogWrite(l Logger, s string) error {
	// TODO: parse s based of the rules of w.flags
	if len(w.prefix) > 0 {
		l.Printf("%s: %s", w.prefix, s)
	} else {
		l.Print(s)
	}
	return nil
}

func newStdLogSink(prefix string, flags int) LogWriterFunc {
	w := stdLogSink{
		prefix: prefix,
		flags:  flags,
	}

	return w.LogWrite
}
