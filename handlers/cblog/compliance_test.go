package cblog_test

import (
	"testing"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/cblog"
	slogtest "darvaza.org/slog/internal/testing"
)

func TestCblogCompliance(t *testing.T) {
	compliance := slogtest.ComplianceTest{
		NewLogger: func() slog.Logger {
			// Use a buffered channel to prevent blocking
			ch := make(chan cblog.LogMsg, 100)
			logger, _ := cblog.New(ch)
			// Drain channel in background to prevent blocking
			go func() {
				var count int
				for range ch {
					count++
				}
			}()
			return logger
		},
	}

	compliance.Run(t)
}

func TestCblogWithCallbackCompliance(t *testing.T) {
	compliance := slogtest.ComplianceTest{
		NewLogger: func() slog.Logger {
			return cblog.NewWithCallback(100, func(_ cblog.LogMsg) {
				// Discard messages
			})
		},
	}

	compliance.Run(t)
}
