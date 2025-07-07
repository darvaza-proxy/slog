package logr

import (
	"strings"

	"github.com/go-logr/logr"

	"darvaza.org/core"
	"darvaza.org/slog"
)

var (
	_ logr.LogSink                = (*Sink)(nil)
	_ logr.CallDepthLogSink       = (*Sink)(nil)
	_ logr.CallStackHelperLogSink = (*Sink)(nil)
)

// Sink implements logr.LogSink using a slog.Logger
type Sink struct {
	logger        slog.Logger
	name          []string
	keysAndValues []any
	callDepth     int
}

// Init initializes the LogSink
func (s *Sink) Init(info logr.RuntimeInfo) {
	s.callDepth = info.CallDepth
}

// Enabled tests whether this LogSink is enabled at the given V-level
func (s *Sink) Enabled(level int) bool {
	if s.logger == nil {
		return false
	}

	// Map logr V-level to slog level and check
	slogLevel := mapFromLogrLevel(level)
	return s.logger.WithLevel(slogLevel).Enabled()
}

// Info logs a non-error message with the given key/value pairs
func (s *Sink) Info(level int, msg string, keysAndValues ...any) {
	if s.logger == nil {
		return
	}

	logger := s.prepareLogger(mapFromLogrLevel(level), keysAndValues)
	logger.Print(msg)
}

// Error logs an error message with the given key/value pairs
func (s *Sink) Error(err error, msg string, keysAndValues ...any) {
	if s.logger == nil {
		return
	}

	logger := s.prepareLogger(slog.Error, keysAndValues)
	if err != nil {
		logger = logger.WithField("error", err.Error())
	}
	logger.Print(msg)
}

// WithValues returns a new LogSink with additional key/value pairs
func (s *Sink) WithValues(keysAndValues ...any) logr.LogSink {
	newKV := make([]any, len(s.keysAndValues)+len(keysAndValues))
	copy(newKV, s.keysAndValues)
	copy(newKV[len(s.keysAndValues):], keysAndValues)

	return &Sink{
		logger:        s.logger,
		name:          s.name,
		keysAndValues: newKV,
		callDepth:     s.callDepth,
	}
}

// WithName returns a new LogSink with the specified name appended
func (s *Sink) WithName(name string) logr.LogSink {
	newName := make([]string, len(s.name)+1)
	copy(newName, s.name)
	newName[len(s.name)] = name

	return &Sink{
		logger:        s.logger,
		name:          newName,
		keysAndValues: s.keysAndValues,
		callDepth:     s.callDepth,
	}
}

// WithCallDepth returns a new LogSink with the specified call depth
func (s *Sink) WithCallDepth(depth int) logr.LogSink {
	return &Sink{
		logger:        s.logger,
		name:          s.name,
		keysAndValues: s.keysAndValues,
		callDepth:     s.callDepth + depth,
	}
}

// GetCallStackHelper returns a function to prune call stacks
func (*Sink) GetCallStackHelper() func() {
	return func() {
		// This is a marker function for call stack pruning
	}
}

// prepareLogger creates a logger with all accumulated context
func (s *Sink) prepareLogger(level slog.LogLevel, keysAndValues []any) slog.Logger {
	logger := s.logger.WithLevel(level)

	// Add name as a field if present
	if len(s.name) > 0 {
		logger = logger.WithField("logger", strings.Join(s.name, "."))
	}

	// Add persistent key-value pairs
	logger = addKeysAndValues(logger, s.keysAndValues)

	// Add call-specific key-value pairs
	logger = addKeysAndValues(logger, keysAndValues)

	// Adjust call depth if needed
	if s.callDepth > 0 {
		logger = logger.WithStack(s.callDepth)
	}

	return logger
}

// addKeysAndValues adds key-value pairs to the logger
func addKeysAndValues(logger slog.Logger, keysAndValues []any) slog.Logger {
	fields := extractFields(keysAndValues)
	if len(fields) == 0 {
		return logger
	}

	// Add fields in sorted order for consistent output
	sortedKeys := core.SortedKeys(fields)
	for _, k := range sortedKeys {
		logger = logger.WithField(k, fields[k])
	}

	return logger
}

// extractFields builds a map from key-value pairs, filtering invalid keys
func extractFields(keysAndValues []any) map[string]any {
	if len(keysAndValues) == 0 {
		return nil
	}

	fields := make(map[string]any)
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		if key, ok := keysAndValues[i].(string); ok && key != "" {
			fields[key] = keysAndValues[i+1]
		}
	}
	return fields
}

// NewSink creates a new logr.LogSink that writes to a slog.Logger
func NewSink(logger slog.Logger) logr.LogSink {
	return &Sink{
		logger: logger,
	}
}

// NewLogr creates a new logr.Logger that writes to a slog.Logger
func NewLogr(logger slog.Logger) logr.Logger {
	return logr.New(NewSink(logger))
}

// mapFromLogrLevel maps logr V-levels to slog levels
// logr V(0) = Info, V(>=1) = Debug
// Note: logr doesn't have a warn level, warnings/errors come via Error()
func mapFromLogrLevel(level int) slog.LogLevel {
	if level <= 0 {
		return slog.Info
	}
	return slog.Debug
}
