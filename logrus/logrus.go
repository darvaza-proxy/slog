package logrus

import (
	"github.com/darvaza-proxy/slog"
)

var (
	_ slog.Logger = (*StructuredLogger)(nil)
)

type StructuredLogger struct{}
