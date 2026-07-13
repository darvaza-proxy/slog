package stdslog

import (
	stdslog "log/slog"

	"darvaza.org/slog"
)

// MapFromSLogLevel maps standard library log/slog levels to slog levels.
// Any level at or above LevelError maps to Error; other values bucket
// down to the nearest named level below them, and anything below
// LevelInfo is Debug.
func MapFromSLogLevel(level stdslog.Level) slog.LogLevel {
	switch {
	case level >= stdslog.LevelError:
		return slog.Error
	case level >= stdslog.LevelWarn:
		return slog.Warn
	case level >= stdslog.LevelInfo:
		return slog.Info
	default:
		return slog.Debug
	}
}

// MapToSLogLevel maps slog levels to standard library log/slog levels.
// Fatal and Panic have no standard equivalent and map above LevelError —
// Fatal to LevelError+4 and Panic to LevelError+8 — preserving slog's
// severity ordering. It rejects UndefinedLevel and values outside the
// range slog defines.
func MapToSLogLevel(level slog.LogLevel) (stdslog.Level, bool) {
	switch level {
	case slog.Panic:
		return stdslog.LevelError + 8, true
	case slog.Fatal:
		return stdslog.LevelError + 4, true
	case slog.Error:
		return stdslog.LevelError, true
	case slog.Warn:
		return stdslog.LevelWarn, true
	case slog.Info:
		return stdslog.LevelInfo, true
	case slog.Debug:
		return stdslog.LevelDebug, true
	default:
		return stdslog.LevelInfo, false
	}
}

// SLogRecordAttrs collects a record's attributes as flattened key/value
// pairs, applying the given group prefix to every key.
func SLogRecordAttrs(record stdslog.Record, prefix string) []any {
	kv := make([]any, 0, record.NumAttrs()*2)
	record.Attrs(func(attr stdslog.Attr) bool {
		kv = AppendSLogAttr(kv, prefix, attr)
		return true
	})
	return kv
}

// AppendSLogAttr flattens attr into kv as a prefixed key/value pair.
// Group-valued attributes expand recursively, extending the prefix with
// the group key, dot-separated; slog fields are flat, so groups map to
// key prefixes. Attributes with an empty key are dropped, except inline
// groups, which expand unprefixed. Values are resolved first, honouring
// stdlib slog.LogValuer.
func AppendSLogAttr(kv []any, prefix string, attr stdslog.Attr) []any {
	value := attr.Value.Resolve()
	if value.Kind() == stdslog.KindGroup {
		return appendSLogGroup(kv, prefix, attr.Key, value.Group())
	}

	if attr.Key == "" {
		return kv
	}
	return append(kv, prefix+attr.Key, value.Any())
}

// appendSLogGroup flattens a group's attributes into kv, extending the
// prefix with the group key. A group with an empty key is inlined
// without extending the prefix.
func appendSLogGroup(kv []any, prefix, key string, attrs []stdslog.Attr) []any {
	if key != "" {
		prefix += key + "."
	}

	for _, attr := range attrs {
		kv = AppendSLogAttr(kv, prefix, attr)
	}
	return kv
}
