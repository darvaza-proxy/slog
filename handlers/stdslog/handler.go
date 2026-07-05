package stdslog

import (
	"context"
	stdslog "log/slog"

	"darvaza.org/core"
	"darvaza.org/slog"
)

var (
	_ stdslog.Handler = (*Handler)(nil)
)

// Handler implements the standard library slog.Handler interface using
// a slog.Logger as backend.
type Handler struct {
	logger      slog.Logger
	groupPrefix string
}

// Enabled reports whether the backend logger would log at the given
// stdlib level.
func (h *Handler) Enabled(_ context.Context, level stdslog.Level) bool {
	if h == nil || h.logger == nil {
		return false
	}
	return h.logger.WithLevel(MapFromSLogLevel(level)).Enabled()
}

// Handle logs the given record through the backend slog.Logger. Levels
// map through the shared floor semantics, capping at slog.Error, so
// inbound records never trigger Fatal or Panic terminal behaviour. The
// record's PC is discarded: slog has no caller-attribution concept, and
// synthesising a stack trace would misrepresent the call site.
func (h *Handler) Handle(_ context.Context, record stdslog.Record) error {
	if h == nil || h.logger == nil {
		return nil
	}

	logger := h.logger.WithLevel(MapFromSLogLevel(record.Level))
	logger = withKeysAndValues(logger,
		SLogRecordAttrs(record, h.groupPrefix))
	logger.Print(record.Message)
	return nil
}

// WithAttrs returns a new Handler with the given attributes attached to
// the backend logger, applying any open group prefix to their keys.
func (h *Handler) WithAttrs(attrs []stdslog.Attr) stdslog.Handler {
	if h == nil || h.logger == nil || len(attrs) == 0 {
		return h
	}

	kv := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		kv = AppendSLogAttr(kv, h.groupPrefix, attr)
	}
	if len(kv) == 0 {
		return h
	}

	return &Handler{
		logger:      withKeysAndValues(h.logger, kv),
		groupPrefix: h.groupPrefix,
	}
}

// WithGroup returns a new Handler that prefixes the keys of all
// subsequently attached attributes with the given name. slog fields are
// flat, so groups map to dot-separated key prefixes. An empty name is a
// no-op, as required by the slog.Handler contract.
func (h *Handler) WithGroup(name string) stdslog.Handler {
	if h == nil || h.logger == nil || name == "" {
		return h
	}

	return &Handler{
		logger:      h.logger,
		groupPrefix: h.groupPrefix + name + ".",
	}
}

// withKeysAndValues attaches flattened key/value pairs to the logger.
// Keys produced by AppendSLogAttr are always non-empty
// strings; duplicate keys resolve last-wins.
func withKeysAndValues(logger slog.Logger, kv []any) slog.Logger {
	if len(kv) == 0 {
		return logger
	}

	fields := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		if key, ok := kv[i].(string); ok && key != "" {
			fields[key] = kv[i+1]
		}
	}
	return logger.WithFields(fields)
}

// NewHandler creates a standard library slog.Handler that writes to the
// given slog.Logger, or nil if none was passed.
func NewHandler(logger slog.Logger) stdslog.Handler {
	if core.IsNil(logger) {
		return nil
	}
	return &Handler{
		logger: logger,
	}
}

// NewSLogger returns a standard library *slog.Logger backed by the
// given slog.Logger, ready for injection into libraries that take one.
// It returns nil if no backend logger was passed.
func NewSLogger(logger slog.Logger) *stdslog.Logger {
	h := NewHandler(logger)
	if h == nil {
		return nil
	}
	return stdslog.New(h)
}
