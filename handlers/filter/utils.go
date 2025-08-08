package filter

import (
	"errors"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

var (
	// errSkip indicates the operation should be skipped (not an error condition).
	errSkip = errors.New("skip")
)

// validateLogLevel checks if a log level is valid, panics if not.
func validateLogLevel(skip int, level slog.LogLevel) error {
	if level <= slog.UndefinedLevel || level > slog.Debug {
		return core.NewPanicErrorf(skip+1, "slog: invalid log level %v", int(level))
	}
	return nil
}

func doWithStack(config *Logger, parent *internal.Loglet, skip int) slog.Logger {
	return &LogEntry{
		loglet: parent.WithStack(skip + 1),
		config: config,
	}
}

func doWithLevel(config *Logger, parent *internal.Loglet, level slog.LogLevel) slog.Logger {
	return &LogEntry{
		loglet: parent.WithLevel(level),
		config: config,
	}
}

func doWithField(config *Logger, parent *internal.Loglet, label string, value any) slog.Logger {
	// Hierarchy: FieldFilter → FieldsFilter → no filter.

	// 1. Try FieldFilter first (most specific for single field).
	if out, ok := createFieldFilter(config, parent, label, value); ok {
		return out
	}

	// 2. Try FieldsFilter as fallback (treat as single-field map).
	if out, ok := createFieldFilterFallback(config, parent, label, value); ok {
		return out
	}

	// 3. No filter - add field as-is.
	return &LogEntry{
		loglet: parent.WithField(label, value),
		config: config,
	}
}

func doWithFields(config *Logger, parent *internal.Loglet, fields map[string]any) slog.Logger {
	// Hierarchy: FieldsFilter → FieldFilter → no filter.

	// 1. Try FieldsFilter first (most specific for field maps).
	if out, ok := createFieldsFilter(config, parent, fields); ok {
		return out
	}

	// 2. Try FieldFilter as fallback (apply to each field individually).
	if out, ok := createFieldsFilterFallback(config, parent, fields); ok {
		return out
	}

	// 3. No filter - add fields as-is.
	return &LogEntry{
		loglet: parent.WithFields(fields),
		config: config,
	}
}

func createFieldFilter(config *Logger, parent *internal.Loglet, label string, value any) (slog.Logger, bool) {
	fn := config.FieldFilter
	if fn == nil {
		// Not handled.
		return nil, false
	}

	filteredLabel, filteredValue, ok := fn(label, value)
	if ok && filteredLabel != "" {
		// Transformed.
		return &LogEntry{
			loglet: parent.WithField(filteredLabel, filteredValue),
			config: config,
		}, true
	}

	// Filtered out.
	return nil, true
}

func createFieldFilterFallback(config *Logger, parent *internal.Loglet, label string, value any) (slog.Logger, bool) {
	fn := config.FieldsFilter
	if fn == nil {
		// Not handled.
		return nil, false
	}

	singleFieldMap := map[string]any{label: value}
	filteredFields, ok := fn(singleFieldMap)
	if ok && internal.HasFields(filteredFields) {
		// Transformed.
		return &LogEntry{
			loglet: parent.WithFields(filteredFields),
			config: config,
		}, true
	}

	// Filtered out.
	return nil, true
}

func createFieldsFilter(config *Logger, parent *internal.Loglet, fields map[string]any) (slog.Logger, bool) {
	fn := config.FieldsFilter
	if fn == nil {
		// Not handled.
		return nil, false
	}

	filteredFields, ok := fn(fields)
	if ok && internal.HasFields(filteredFields) {
		// Transformed.
		return &LogEntry{
			loglet: parent.WithFields(filteredFields),
			config: config,
		}, true
	}

	// Filtered out.
	return nil, true
}

func createFieldsFilterFallback(config *Logger, parent *internal.Loglet, fields map[string]any) (slog.Logger, bool) {
	fn := config.FieldFilter
	if fn == nil {
		// Not handled.
		return nil, false
	}

	filtered := make(map[string]any, len(fields))
	for k, v := range fields {
		if newK, newV, ok := fn(k, v); ok && newK != "" {
			filtered[newK] = newV
		}
	}

	if internal.HasFields(filtered) {
		// Transformed.
		return &LogEntry{
			loglet: parent.WithFields(filtered),
			config: config,
		}, true
	}

	// Filtered out.
	return nil, true
}

func applyMessageFilter(config *Logger, msg string) (string, bool) {
	if fn := config.MessageFilter; fn != nil {
		return fn(msg)
	}
	return msg, true
}
