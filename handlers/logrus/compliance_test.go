package logrus_test

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"

	"darvaza.org/slog"
	slogrus "darvaza.org/slog/handlers/logrus"
	slogtest "darvaza.org/slog/internal/testing"
)

func TestLogrusCompliance(t *testing.T) {
	newLogger := func() slog.Logger {
		// Create a basic logrus logger
		logrusLogger := logrus.New()
		logrusLogger.SetLevel(logrus.DebugLevel)
		return slogrus.New(logrusLogger)
	}

	// For bidirectional testing, we need to create a logrus logger
	// that outputs to the recorder
	newLoggerWithRecorder := func(recorder slog.Logger) slog.Logger {
		// Create logrus logger with custom hook that forwards to recorder
		logrusLogger := logrus.New()
		logrusLogger.SetLevel(logrus.DebugLevel)

		// Add hook to forward to recorder
		hook := slogrus.NewSlogHook(recorder)
		logrusLogger.AddHook(hook)

		// Disable normal output since hook handles it
		logrusLogger.SetOutput(io.Discard)

		return slogrus.New(logrusLogger)
	}

	compliance := slogtest.ComplianceTest{
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger:             newLogger,
			NewLoggerWithRecorder: newLoggerWithRecorder,
		},
		// Logrus will exit on Fatal/Panic
		SkipPanicTests: true,
	}

	compliance.Run(t)
}
