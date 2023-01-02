package zerolog

import (
	"github.com/rs/zerolog"

	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*ZeroLogger)(nil)
)

type ZeroLogger struct {
	entry *zerolog.Logger
}

func NewWrapper(entry *zerolog.Logger) slog.Logger {
	return &ZeroLogger{
		entry: logger,
	}
}
