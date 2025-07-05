package filter_test

import (
	"testing"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	slogtest "darvaza.org/slog/internal/testing"
)

func TestFilterCompliance(t *testing.T) {
	compliance := slogtest.ComplianceTest{
		NewLogger: func() slog.Logger {
			base := slogtest.NewLogger()
			return filter.New(base, slog.Debug)
		},
	}

	compliance.Run(t)
}

func TestFilterWithThresholdCompliance(t *testing.T) {
	// Test with different thresholds
	thresholds := []struct {
		name  string
		level slog.LogLevel
	}{
		{"Debug", slog.Debug},
		{"Info", slog.Info},
		{"Warn", slog.Warn},
		{"Error", slog.Error},
	}

	for _, tc := range thresholds {
		t.Run(tc.name, func(t *testing.T) {
			threshold := tc.level
			compliance := slogtest.ComplianceTest{
				NewLogger: func() slog.Logger {
					base := slogtest.NewLogger()
					return filter.New(base, threshold)
				},
			}

			compliance.Run(t)
		})
	}
}
