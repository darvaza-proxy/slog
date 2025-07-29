package filter

import (
	"fmt"
	"log"
	"os"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/internal"
)

var (
	_ slog.Logger = (*LogEntry)(nil)
)

// LogEntry implements a level filtered logger
type LogEntry struct {
	internal.Loglet

	logger *Logger
	entry  slog.Logger
}

// Enabled tells this logger would record logs
func (l *LogEntry) Enabled() bool {
	if l == nil || l.logger == nil {
		return false
	}
	level := l.Level()
	if level <= slog.UndefinedLevel || level > l.logger.Threshold {
		return false
	}
	return l.entry == nil || l.entry.Enabled()
}

// WithEnabled returns itself and if it's enabled
func (l *LogEntry) WithEnabled() (slog.Logger, bool) {
	return l, l.Enabled()
}

// Print would, if conditions are met, add a log entry with the arguments
// in the manner of fmt.Print
func (l *LogEntry) Print(args ...any) {
	if l.Enabled() {
		l.msg(fmt.Sprint(args...))
	}
}

// Println would, if conditions are met, add a log entry with the arguments
// in the manner of fmt.Println
func (l *LogEntry) Println(args ...any) {
	if l.Enabled() {
		l.msg(fmt.Sprintln(args...))
	}
}

// Printf would, if conditions are met, add a log entry with the arguments
// in the manner of fmt.Printf
func (l *LogEntry) Printf(format string, args ...any) {
	if l.Enabled() {
		l.msg(fmt.Sprintf(format, args...))
	}
}

// msg applies MessageFilter and FieldFilter before sending the message to
// the parent Logger. This is where all field filtering happens with complete context.
func (l *LogEntry) msg(msg string) {
	if fn := l.logger.MessageFilter; fn != nil {
		var ok bool

		msg, ok = fn(msg)
		if !ok {
			return
		}
	}

	if l.entry == nil {
		// parentless is either Fatal or Panic
		_ = log.Output(3, msg)

		if l.Level() != slog.Fatal {
			panic(msg)
		}

		// revive:disable:deep-exit
		os.Exit(1)
	}

	// Apply field filtering with complete context at print time
	entry := l.applyFieldFiltersAtPrintTime(l.entry)
	entry.Print(msg)
}

// applyFieldFiltersAtPrintTime collects all fields from the logger chain
// and applies field filters with complete context before forwarding to parent
func (l *LogEntry) applyFieldFiltersAtPrintTime(entry slog.Logger) slog.Logger {
	allFields := l.collectAllFields()
	if len(allFields) == 0 {
		return entry
	}

	return l.processFields(entry, allFields)
}

// collectAllFields gathers all fields from the complete logger chain
func (l *LogEntry) collectAllFields() map[string]any {
	count := l.FieldsCount()
	if count == 0 {
		return nil
	}

	fields := make(map[string]any, count)
	iter := l.Loglet.Fields()
	for iter.Next() {
		k, v := iter.Field()
		if k != "" {
			fields[k] = v
		}
	}
	return fields
}

// processFields applies field filters/overrides and forwards to parent
func (l *LogEntry) processFields(entry slog.Logger, allFields map[string]any) slog.Logger {
	// FieldsOverride intercepts all fields at once
	if fn := l.logger.FieldsOverride; fn != nil {
		fn(entry, allFields)
		return entry
	}

	// FieldOverride processes each field individually
	if fn := l.logger.FieldOverride; fn != nil {
		for _, key := range core.SortedKeys(allFields) {
			fn(entry, key, allFields[key])
		}
		return entry
	}

	// FieldFilter modifies fields before forwarding
	if fn := l.logger.FieldFilter; fn != nil {
		allFields = applyFieldFilter(allFields, fn)
	}

	return entry.WithFields(allFields)
}

// applyFieldFilter applies the FieldFilter function to all fields
func applyFieldFilter(
	fields map[string]any,
	filter func(string, any) (string, any, bool),
) map[string]any {
	filtered := make(map[string]any, len(fields))
	for key, value := range fields {
		if newKey, newValue, ok := filter(key, value); ok && newKey != "" {
			filtered[newKey] = newValue
		}
	}
	return filtered
}

// Debug creates a new filtered logger on level slog.Debug
func (l *LogEntry) Debug() slog.Logger {
	return l.logger.WithLevel(slog.Debug)
}

// Info creates a new filtered logger on level slog.Info
func (l *LogEntry) Info() slog.Logger {
	return l.logger.WithLevel(slog.Info)
}

// Warn creates a new filtered logger on level slog.Warn
func (l *LogEntry) Warn() slog.Logger {
	return l.logger.WithLevel(slog.Warn)
}

// Error creates a new filtered logger on level slog.Error
func (l *LogEntry) Error() slog.Logger {
	return l.logger.WithLevel(slog.Error)
}

// Fatal creates a new filtered logger on level slog.Fatal
func (l *LogEntry) Fatal() slog.Logger {
	return l.logger.WithLevel(slog.Fatal)
}

// Panic creates a new filtered logger on level slog.Panic
func (l *LogEntry) Panic() slog.Logger {
	return l.logger.WithLevel(slog.Panic)
}

// WithLevel creates a new filtered logger on the given level
func (l *LogEntry) WithLevel(level slog.LogLevel) slog.Logger {
	return l.logger.WithLevel(level)
}

// WithStack would, if conditions are met, attach a call stack to the log entry
func (l *LogEntry) WithStack(skip int) slog.Logger {
	out := &LogEntry{
		Loglet: l.Loglet.WithStack(skip + 1),
		logger: l.logger,
		entry:  l.entry,
	}

	if l.Enabled() && l.entry != nil {
		out.entry = l.entry.WithStack(skip + 1)
	}
	return out
}

// WithField would, if conditions are met, attach a field to the log entry. This
// field could be altered if a FieldFilter is used
func (l *LogEntry) WithField(label string, value any) slog.Logger {
	if label != "" {
		out := &LogEntry{
			Loglet: l.Loglet.WithField(label, value),
			logger: l.logger,
			entry:  l.entry,
		}

		if l.Enabled() && l.entry != nil {
			out.entry = l.entry
		}
		return out
	}
	return l
}

// WithFields would, if conditions are met, attach fields to the log entry.
// These fields could be altered if a FieldFilter is used
func (l *LogEntry) WithFields(fields map[string]any) slog.Logger {
	if internal.HasFields(fields) {
		out := &LogEntry{
			Loglet: l.Loglet.WithFields(fields),
			logger: l.logger,
			entry:  l.entry,
		}

		if l.Enabled() && l.entry != nil {
			out.entry = l.entry
		}
		return out
	}
	return l
}
