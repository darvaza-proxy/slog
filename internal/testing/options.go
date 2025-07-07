package testing

import "darvaza.org/slog"

// AdapterOptions provides common options for testing adapters.
type AdapterOptions struct {
	// LevelExceptions maps expected level transformations.
	// For example, if an adapter maps Warn to Info, set:
	// LevelExceptions: map[slog.LogLevel]slog.LogLevel{slog.Warn: slog.Info}
	LevelExceptions map[slog.LogLevel]slog.LogLevel
}

// ExpectedLevel returns the expected level after transformation.
// If there's no exception mapping, returns the original level.
func (opts *AdapterOptions) ExpectedLevel(level slog.LogLevel) slog.LogLevel {
	if opts == nil || len(opts.LevelExceptions) == 0 {
		return level
	}
	if mapped, exists := opts.LevelExceptions[level]; exists {
		return mapped
	}
	return level
}

// FactoryOptions provides factory functions for creating loggers.
type FactoryOptions struct {
	// NewLogger creates a new logger instance for testing.
	// If the logger writes to a test recorder, it should be a fresh instance.
	NewLogger func() slog.Logger

	// NewLoggerWithRecorder creates a logger using the provided recorder as backend.
	// This is used for testing adapters that need message verification.
	// If provided, it takes precedence over NewLogger for tests that need
	// to verify output (like concurrency tests).
	NewLoggerWithRecorder func(slog.Logger) slog.Logger
}

// BidirectionalTestOptions configures bidirectional adapter tests
type BidirectionalTestOptions struct {
	AdapterOptions
}

// ExpectedLevel returns the expected level after transformation.
// This method handles nil receivers properly, unlike calling through embedded field.
func (opts *BidirectionalTestOptions) ExpectedLevel(level slog.LogLevel) slog.LogLevel {
	if opts == nil {
		return level
	}
	return opts.AdapterOptions.ExpectedLevel(level)
}

// ConcurrencyTestOptions provides options for concurrent testing.
type ConcurrencyTestOptions struct {
	AdapterOptions
	FactoryOptions

	// GetMessages returns the messages logged during the test.
	// This allows production loggers to provide their messages for verification.
	// Only used if NewLoggerWithRecorder is nil.
	GetMessages func() []Message
}
