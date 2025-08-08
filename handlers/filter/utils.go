package filter

import (
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

func doWithLevel(config *Logger, parent *internal.Loglet, level slog.LogLevel) slog.Logger {
	var out slog.Logger
	switch {
	case level <= slog.UndefinedLevel:
		// fix your code
		config.Panic().WithStack(1).Printf("slog: invalid log level %v", level)
	case level > slog.Fatal && config.Parent == nil:
		// Parentless non-Fatal, NOOP
		out = config
	default:
		out = &LogEntry{
			entry:  parent.WithLevel(level),
			config: config,
		}
	}
	return out
}

func doWithStack(config *Logger, parent *internal.Loglet, skip int) slog.Logger {
	return &LogEntry{
		entry:  parent.WithStack(skip + 1),
		config: config,
	}
}

func doWithField(config *Logger, parent *internal.Loglet, label string, value any) slog.Logger {
	// Hierarchy: FieldFilter → FieldsFilter → no filter (handled by caller)

	// 1. Try FieldFilter first (most specific for single field)
	if out, ok := createFieldFilter(config, parent, label, value); ok {
		return out
	}

	// 2. Try FieldsFilter as fallback (treat as single-field map)
	if out, ok := createFieldFilterFallback(config, parent, label, value); ok {
		return out
	}

	// 3. No filter - let caller handle
	return nil
}

func doWithFields(config *Logger, parent *internal.Loglet, fields map[string]any) slog.Logger {
	// Hierarchy: FieldsFilter → FieldFilter → no filter (handled by caller)

	// 1. Try FieldsFilter first (most specific for field maps)
	if out, ok := createFieldsFilter(config, parent, fields); ok {
		return out
	}

	// 2. Try FieldFilter as fallback (apply to each field individually)
	if out, ok := createFieldsFilterFallback(config, parent, fields); ok {
		return out
	}

	// 3. No filter - let caller handle
	return nil
}

func createFieldFilter(config *Logger, parent *internal.Loglet, label string, value any) (slog.Logger, bool) {
	fn := config.FieldFilter
	if fn == nil {
		// not handled
		return nil, false
	}

	filteredLabel, filteredValue, ok := fn(label, value)
	if ok && filteredLabel != "" {
		// transformed
		return &LogEntry{
			entry:  parent.WithField(filteredLabel, filteredValue),
			config: config,
		}, true
	}

	// filtered out
	return nil, true
}

func createFieldFilterFallback(config *Logger, parent *internal.Loglet, label string, value any) (slog.Logger, bool) {
	fn := config.FieldsFilter
	if fn == nil {
		// not handled
		return nil, false
	}

	singleFieldMap := map[string]any{label: value}
	filteredFields, ok := fn(singleFieldMap)
	if ok && internal.HasFields(filteredFields) {
		// transformed
		return &LogEntry{
			entry:  parent.WithFields(filteredFields),
			config: config,
		}, true
	}

	// filtered out
	return nil, true
}

func createFieldsFilter(config *Logger, parent *internal.Loglet, fields map[string]any) (slog.Logger, bool) {
	fn := config.FieldsFilter
	if fn == nil {
		// not handled
		return nil, false
	}

	filteredFields, ok := fn(fields)
	if ok && internal.HasFields(filteredFields) {
		// transformed
		return &LogEntry{
			entry:  parent.WithFields(filteredFields),
			config: config,
		}, true
	}

	// filtered out
	return nil, true
}

func createFieldsFilterFallback(config *Logger, parent *internal.Loglet, fields map[string]any) (slog.Logger, bool) {
	fn := config.FieldFilter
	if fn == nil {
		// not handled
		return nil, false
	}

	filtered := make(map[string]any, len(fields))
	for k, v := range fields {
		if newK, newV, ok := fn(k, v); ok && newK != "" {
			filtered[newK] = newV
		}
	}

	if internal.HasFields(filtered) {
		// transformed
		return &LogEntry{
			entry:  parent.WithFields(filtered),
			config: config,
		}, true
	}

	// filtered out
	return nil, true
}

func applyMessageFilter(config *Logger, msg string) (string, bool) {
	if fn := config.MessageFilter; fn != nil {
		return fn(msg)
	}
	return msg, true
}
