package filter_test

import (
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
)

// NewTestFilter creates a filter logger for testing.
// Unlike filter.New, it preserves UndefinedLevel threshold for testing disabled loggers.
func NewTestFilter(parent slog.Logger, threshold slog.LogLevel) slog.Logger {
	if threshold == slog.UndefinedLevel {
		// Direct struct creation to avoid filter.New() defaulting to Error
		return &filter.Logger{
			Parent:    parent,
			Threshold: slog.UndefinedLevel,
		}
	}
	return filter.New(parent, threshold)
}
