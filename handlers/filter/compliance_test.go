package filter_test

import (
	"testing"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/filter"
	slogtest "darvaza.org/slog/internal/testing"
)

func TestFilterCompliance(t *testing.T) {
	newLogger := func() slog.Logger {
		base := slogtest.NewLogger()
		return filter.New(base, slog.Debug)
	}

	newLoggerWithRecorder := func(recorder slog.Logger) slog.Logger {
		// Filter with recorder as the base
		return filter.New(recorder, slog.Debug)
	}

	compliance := slogtest.ComplianceTest{
		ConcurrencyTestOptions: slogtest.ConcurrencyTestOptions{
			FactoryOptions: slogtest.FactoryOptions{
				NewLogger:             newLogger,
				NewLoggerWithRecorder: newLoggerWithRecorder,
			},
		},
	}

	compliance.Run(t)
}

func TestFilterWithThresholdCompliance(t *testing.T) {
	// Test with different thresholds
	thresholds := []struct {
		name            string
		level           slog.LogLevel
		levelExceptions map[slog.LogLevel]slog.LogLevel
	}{
		{
			name:            "Debug",
			level:           slog.Debug,
			levelExceptions: nil, // All levels pass through
		},
		{
			name:  "Info",
			level: slog.Info,
			levelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Debug: slog.UndefinedLevel, // Debug is filtered out
			},
		},
		{
			name:  "Warn",
			level: slog.Warn,
			levelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Debug: slog.UndefinedLevel, // Debug is filtered out
				slog.Info:  slog.UndefinedLevel, // Info is filtered out
			},
		},
		{
			name:  "Error",
			level: slog.Error,
			levelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Debug: slog.UndefinedLevel, // Debug is filtered out
				slog.Info:  slog.UndefinedLevel, // Info is filtered out
				slog.Warn:  slog.UndefinedLevel, // Warn is filtered out
			},
		},
	}

	for _, tc := range thresholds {
		t.Run(tc.name, func(t *testing.T) {
			threshold := tc.level

			newLogger := func() slog.Logger {
				base := slogtest.NewLogger()
				return filter.New(base, threshold)
			}

			newLoggerWithRecorder := func(recorder slog.Logger) slog.Logger {
				return filter.New(recorder, threshold)
			}

			compliance := slogtest.ComplianceTest{
				ConcurrencyTestOptions: slogtest.ConcurrencyTestOptions{
					AdapterOptions: slogtest.AdapterOptions{
						LevelExceptions: tc.levelExceptions,
					},
					FactoryOptions: slogtest.FactoryOptions{
						NewLogger:             newLogger,
						NewLoggerWithRecorder: newLoggerWithRecorder,
					},
				},
			}

			compliance.Run(t)
		})
	}
}

func TestFilterBidirectional(t *testing.T) {
	// Test that filter behaves as a bidirectional adapter
	factory := func(backend slog.Logger) slog.Logger {
		return filter.New(backend, slog.Info)
	}

	opts := &slogtest.BidirectionalTestOptions{
		AdapterOptions: slogtest.AdapterOptions{
			LevelExceptions: map[slog.LogLevel]slog.LogLevel{
				slog.Debug: slog.UndefinedLevel, // Debug messages filtered out
			},
		},
	}

	slogtest.TestBidirectionalWithOptions(t, "FilterInfo", factory, opts)
}
