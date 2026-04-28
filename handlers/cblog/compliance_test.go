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
func newSyncCblogFactory(t *testing.T) (func() slog.Logger, func(slog.Logger) slog.Logger, func() []slogtest.Message) {
	newLogger := newCblogRecordingFactory(1000)

	// Messages are handled synchronously, so no waiting needed.
	getMessages := func() []slogtest.Message { return nil }

	factory := func() slog.Logger { return newDrainedCblogLogger(t) }
	return factory, newLogger, getMessages
}

// newCblogRecordingFactory returns a factory that builds a cblog logger
// whose callback forwards each message to the supplied recorder.
func newCblogRecordingFactory(buffer int) func(slog.Logger) slog.Logger {
	return func(recorder slog.Logger) slog.Logger {
		return cblog.NewWithCallback(buffer, func(msg cblog.LogMsg) {
			forwardCblogMsg(recorder, msg)
		})
	}
}

// forwardCblogMsg replays msg into recorder, mirroring stack and fields
// when the source message carries them.
func forwardCblogMsg(recorder slog.Logger, msg cblog.LogMsg) {
	recLogger := recorder.WithLevel(msg.Level)
	if msg.Stack != nil {
		recLogger = recLogger.WithStack(0)
	}
	if msg.Fields != nil {
		recLogger = recLogger.WithFields(msg.Fields)
	}
	recLogger.Print(msg.Message)
}

// newDrainedCblogLogger returns a cblog logger whose channel is drained
// in the background so callers do not block on the buffer. The channel
// is closed in t.Cleanup so the drainer goroutine exits when the test
// finishes.
func newDrainedCblogLogger(t *testing.T) slog.Logger {
	ch := make(chan cblog.LogMsg, 100)
	logger, _ := cblog.New(ch)
	t.Cleanup(func() { close(ch) })
	go func() {
		//revive:disable-next-line:empty-block
		for range ch {
		}
	}()
	return logger
}

// TestCblogCompliance tests cblog compliance with slog.Logger interface.
func TestCblogCompliance(t *testing.T) {
	// Create synchronized cblog factory
	newLogger, newLoggerWithRecorder, _ := newSyncCblogFactory(t)

	compliance := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger:             newLogger,
			NewLoggerWithRecorder: newLoggerWithRecorder,
		},
	}

	compliance.Run(t)
}

func newDiscardingCblogLogger() slog.Logger {
	return cblog.NewWithCallback(100, func(_ cblog.LogMsg) {})
}

func TestCblogWithCallbackCompliance(t *testing.T) {
	compliance := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger:             newDiscardingCblogLogger,
			NewLoggerWithRecorder: newCblogRecordingFactory(1000),
		},
	}
	compliance.Run(t)
}
