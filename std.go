package slog

import (
	"log"
)

func NewStdLogger(l Logger, prefix string, flag int) *log.Logger {
	w := NewLogWriter(l, newStdLogSink(prefix, flag))

	return log.New(w, "", flag)
}

type stdLogSink struct {
	prefix string
	flag   int
}

func (w *stdLogSink) LogWrite(l Logger, s string) error {
	if len(w.prefix) > 0 {
		l.Printf("%s: %s", w.prefix, s)
	} else {
		l.Printf("%s", s)
	}
	return nil
}

func newStdLogSink(prefix string, flag int) LogWriterFunc {
	w := stdLogSink{
		prefix: prefix,
		flag:   flag,
	}

	return w.LogWrite
}
