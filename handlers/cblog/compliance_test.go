package cblog_test

import (
	"testing"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/cblog"
	slogtest "darvaza.org/slog/internal/testing"
)

// newSyncCblogFactory creates a synchronized cblog logger factory for testing.
// It returns three functions:
// 1. A simple logger factory (for tests that don't need verification)
// 2. A logger factory that creates cblog loggers with message tracking
// 3. A GetMessages function that waits for all messages to be processed
func newSyncCblogFactory() (func() slog.Logger, func(slog.Logger) slog.Logger, func() []slogtest.Message) {
	// Use NewWithCallback for synchronous testing to avoid race conditions
	newLogger := func(recorder slog.Logger) slog.Logger {
		return cblog.NewWithCallback(1000, func(msg cblog.LogMsg) {
			// Forward to recorder synchronously
			recLogger := recorder.WithLevel(msg.Level)
			if msg.Stack != nil {
				recLogger = recLogger.WithStack(0)
			}
			if msg.Fields != nil {
				recLogger = recLogger.WithFields(msg.Fields)
			}
			recLogger.Print(msg.Message)
		})
	}

	getMessages := func() []slogtest.Message {
		// Messages are handled synchronously, so no waiting needed
		return nil
	}

	// Simple logger factory without recording
	simpleLogger := func() slog.Logger {
		// Use a buffered channel to prevent blocking
		ch := make(chan cblog.LogMsg, 100)
		logger, _ := cblog.New(ch)
		// Drain channel in background to prevent blocking
		go func() {
			for msg := range ch {
				_ = msg // Discard messages
			}
		}()
		return logger
	}

	return simpleLogger, newLogger, getMessages
}

// TestCblogCompliance tests cblog compliance with slog.Logger interface.
func TestCblogCompliance(t *testing.T) {
	// Create synchronized cblog factory
	newLogger, newLoggerWithRecorder, _ := newSyncCblogFactory()

	compliance := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger:             newLogger,
			NewLoggerWithRecorder: newLoggerWithRecorder,
		},
	}

	compliance.Run(t)
}

func TestCblogWithCallbackCompliance(t *testing.T) {
	newLogger := func() slog.Logger {
		return cblog.NewWithCallback(100, func(msg cblog.LogMsg) {
			_ = msg // Discard messages
		})
	}

	newLoggerWithRecorder := func(recorder slog.Logger) slog.Logger {
		return cblog.NewWithCallback(1000, func(msg cblog.LogMsg) { // Larger buffer
			// Forward to recorder
			recLogger := recorder.WithLevel(msg.Level)
			if msg.Stack != nil {
				recLogger = recLogger.WithStack(0)
			}
			if msg.Fields != nil {
				recLogger = recLogger.WithFields(msg.Fields)
			}
			recLogger.Print(msg.Message)
		})
	}

	compliance := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger:             newLogger,
			NewLoggerWithRecorder: newLoggerWithRecorder,
		},
	}

	compliance.Run(t)
}
