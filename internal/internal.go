// Package internal provides helpers for concrete slog implementations
package internal

import "darvaza.org/slog"

// IsTerminalLevel tells if a level is Fatal or Panic, the levels whose
// entries end in an exit or a panic.
func IsTerminalLevel(level slog.LogLevel) bool {
	return level >= slog.Panic && level <= slog.Fatal
}
