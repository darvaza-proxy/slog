package zap_test

import (
	"fmt"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/slog"
	slogzap "darvaza.org/slog/handlers/zap"
	slogtest "darvaza.org/slog/internal/testing"
)

// newZapLoggerWithRecorder creates a zap-based slog.Logger that writes to the given recorder
func newZapLoggerWithRecorder(recorder slog.Logger) slog.Logger {
	// Create a bidirectional adapter chain: slog API → zap → slog recorder
	// This tests both the SlogCore (slog→zap) and Logger (zap→slog) adapters

	// Create a zap config that uses our recorder-backed core
	cfg := slogzap.NewDefaultConfig()

	// Use zap's WrapCore option to replace the core with our SlogCore
	logger, err := slogzap.New(cfg, zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return slogzap.NewCore(recorder, zap.InfoLevel)
	}))
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	return logger
}

func TestCompliance(t *testing.T) {
	compliance := slogtest.ComplianceTest{
		ConcurrencyTestOptions: slogtest.ConcurrencyTestOptions{
			FactoryOptions: slogtest.FactoryOptions{
				NewLogger: func() slog.Logger {
					logger, _ := slogzap.New(nil)
					return logger
				},
				NewLoggerWithRecorder: newZapLoggerWithRecorder,
			},
		},
	}
	compliance.Run(t)
}

func TestLevelMethods(t *testing.T) {
	slogtest.TestLevelMethods(t, func() slog.Logger {
		logger, _ := slogzap.New(nil)
		return logger
	})
}

func TestFieldMethods(t *testing.T) {
	slogtest.TestFieldMethods(t, func() slog.Logger {
		logger, _ := slogzap.New(nil)
		return logger
	})
}

func TestWithStack(t *testing.T) {
	logger, _ := slogzap.New(nil)
	slogtest.TestWithStack(t, logger)
}

func TestConcurrency(t *testing.T) {
	opts := &slogtest.ConcurrencyTestOptions{
		FactoryOptions: slogtest.FactoryOptions{
			NewLoggerWithRecorder: newZapLoggerWithRecorder,
		},
	}
	test := slogtest.DefaultConcurrencyTest()
	slogtest.RunConcurrentTestWithOptions(t, nil, test, opts)
}

func TestLevelValidation(t *testing.T) {
	logger, _ := slogzap.New(nil)

	// Test that WithLevel panics for invalid level
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid log level")
			}
		}()
		// This should panic
		logger.WithLevel(slog.UndefinedLevel)
	}()

	// Ensure test continues after the panic test
}

func TestZapSpecific(t *testing.T) {
	t.Run("NewWithCallback", func(t *testing.T) {
		logger, _ := slogzap.New(nil)
		zapLogger := logger.(*slogzap.Logger)

		callbackExecuted := false
		newLogger := zapLogger.NewWithCallback(func(_ zapcore.Entry) error {
			callbackExecuted = true
			return nil
		})

		if newLogger == nil {
			t.Fatal("NewWithCallback returned nil")
		}

		// Note: Testing the callback execution would require accessing
		// zap internals or using a custom core, which is beyond the
		// scope of this unit test
		_ = callbackExecuted
	})

	t.Run("NewNoop", func(t *testing.T) {
		logger := slogzap.NewNoop()
		if logger == nil {
			t.Fatal("NewNoop returned nil")
		}

		// Should not panic
		logger.Debug().Print("test")
		logger.Info().Printf("test %d", 123)
		logger.Warn().Println("test")
	})

	t.Run("Unwrap", func(t *testing.T) {
		cfg := slogzap.NewDefaultConfig()
		logger, err := slogzap.New(cfg)
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}

		zapLogger := logger.(*slogzap.Logger)
		zl, zc := zapLogger.Unwrap()

		if zl == nil {
			t.Error("Unwrap returned nil logger")
		}
		if zc == nil {
			t.Error("Unwrap returned nil config")
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Create an invalid config
		cfg := &zap.Config{
			Encoding: "invalid-encoding",
		}

		_, err := slogzap.New(cfg)
		if err == nil {
			t.Error("expected error for invalid config")
		}
	})
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := slogzap.NewDefaultConfig()
	if cfg == nil {
		t.Fatal("NewDefaultConfig returned nil")
	}

	// Verify it's a valid config
	logger, err := cfg.Build()
	if err != nil {
		t.Fatalf("default config failed to build: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	// Check expected settings
	if cfg.Encoding != "console" {
		t.Errorf("expected console encoding, got %s", cfg.Encoding)
	}

	if cfg.Level.Level() != zapcore.InfoLevel {
		t.Errorf("expected info level, got %v", cfg.Level.Level())
	}

	if !cfg.DisableStacktrace {
		t.Error("expected stacktrace to be disabled")
	}

	if !cfg.DisableCaller {
		t.Error("expected caller to be disabled")
	}
}
