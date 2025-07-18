package logrus

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"

	"darvaza.org/core"
	"darvaza.org/slog"
)

// SlogHook is a logrus hook that forwards all log entries to a slog.Logger
type SlogHook struct {
	logger slog.Logger
}

var _ logrus.Hook = (*SlogHook)(nil)

// NewSlogHook creates a logrus Hook that forwards to slog
func NewSlogHook(logger slog.Logger) *SlogHook {
	if logger == nil {
		panic("nil slog.Logger")
	}
	return &SlogHook{
		logger: logger,
	}
}

// Levels returns all log levels (hook is called for all levels)
func (*SlogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is called when logging happens
func (h *SlogHook) Fire(entry *logrus.Entry) (err error) {
	defer func() {
		if e := core.AsRecovered(recover()); e != nil {
			// panic in Fire should not crash the program
			err = e
		}
	}()

	// Map logrus level to slog level
	slogLevel := mapLogrusToSlogLevel(entry.Level)
	logger := h.logger.WithLevel(slogLevel)

	// Add all fields from the entry
	for k, v := range entry.Data {
		logger = logger.WithField(k, v)
	}

	// Add caller information if available
	if entry.HasCaller() {
		logger = logger.WithField("caller", formatCaller(entry.Caller))
	}

	// Log the message
	logger.Print(entry.Message)

	return nil
}

// SetupLogrusToSlog configures a logrus logger to send all output to slog
func SetupLogrusToSlog(logrusLogger *logrus.Logger, slogLogger slog.Logger) {
	// Add our hook
	logrusLogger.AddHook(NewSlogHook(slogLogger))

	// Disable logrus's own output to avoid double logging
	logrusLogger.SetOutput(io.Discard)
}

// NewLogrusLogger creates a new logrus.Logger that outputs to slog
func NewLogrusLogger(slogLogger slog.Logger) *logrus.Logger {
	logger := logrus.New()
	// Set to trace level to ensure all messages are processed
	logger.SetLevel(logrus.TraceLevel)
	SetupLogrusToSlog(logger, slogLogger)
	return logger
}

// mapLogrusToSlogLevel converts logrus levels to slog levels
func mapLogrusToSlogLevel(level logrus.Level) slog.LogLevel {
	switch level {
	case logrus.TraceLevel:
		return slog.Debug // slog doesn't have Trace, map to Debug
	case logrus.DebugLevel:
		return slog.Debug
	case logrus.InfoLevel:
		return slog.Info
	case logrus.WarnLevel:
		return slog.Warn
	case logrus.ErrorLevel:
		return slog.Error
	case logrus.FatalLevel:
		return slog.Fatal
	case logrus.PanicLevel:
		return slog.Panic
	default:
		return slog.Info
	}
}

// formatCaller formats runtime.Frame as a string
func formatCaller(frame *runtime.Frame) string {
	if frame == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
}
