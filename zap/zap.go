package zap

import (
	_ "go.uber.org/zap"

	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*ZapLogger)(nil)
)

// ZapLoggerEntry is either a zap.Logger or a zap.SugaredLogger
type ZapLoggerEntry interface{}

type ZapLogger struct {
	entry ZapLoggerEntry
}

func NewWrapper(entry ZapLoggerEntry) slog.Logger {
	return &ZapLogger{
		entry: entry,
	}
}
