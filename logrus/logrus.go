package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*LogrusLogger)(nil)
)

type LogrusLogger struct {
	entry *logrus.Logger
}

func NewWrapper(entry *logrus.Logger) slog.Logger {
	return &LogrusLogger{
		entry: entry,
	}
}
