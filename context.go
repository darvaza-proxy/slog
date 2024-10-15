package slog

import (
	"context"

	"darvaza.org/core"
)

// WithLogger attaches a [Logger] to the given context.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return ctxLoggerKey.WithValue(ctx, logger)
}

// GetLogger attempts to extract a [Logger] from the given
// context.
func GetLogger(ctx context.Context) (Logger, bool) {
	return ctxLoggerKey.Get(ctx)
}

var ctxLoggerKey = core.NewContextKey[Logger]("logger")
