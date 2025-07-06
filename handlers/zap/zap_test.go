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
		FactoryOptions: slogtest.FactoryOptions{
			NewLogger: func() slog.Logger {
				logger, _ := slogzap.New(nil)
				return logger
			},
			NewLoggerWithRecorder: newZapLoggerWithRecorder,
		},
	}
	compliance.Run(t)
}

func TestStress(t *testing.T) {
	suite := slogtest.StressTestSuite{
		NewLogger: func() slog.Logger {
			logger, _ := slogzap.New(nil)
			return logger
		},
		NewLoggerWithRecorder: newZapLoggerWithRecorder,
	}
	suite.Run(t)
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
	t.Run("NewWithCallback", testNewWithCallback)
	t.Run("NewWithCallback_NilCallback", testNewWithCallbackNilCallback)
	t.Run("NewNoop", testNewNoop)
	t.Run("Unwrap", testUnwrap)
	t.Run("ErrorHandling", testErrorHandling)
	t.Run("Enabled_NilLogger", testEnabledNilLogger)
	t.Run("WithLevel_SameLevel", testWithLevelSameLevel)
}

func testNewWithCallback(t *testing.T) {
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
}

func testNewWithCallbackNilCallback(t *testing.T) {
	logger, _ := slogzap.New(nil)
	zapLogger := logger.(*slogzap.Logger)

	// Test with nil callback
	newLogger := zapLogger.NewWithCallback(nil)

	// Should return the same logger instance when callback is nil
	if newLogger != zapLogger {
		t.Error("NewWithCallback with nil callback should return the same logger instance")
	}
}

func testNewNoop(t *testing.T) {
	logger := slogzap.NewNoop()
	if logger == nil {
		t.Fatal("NewNoop returned nil")
	}

	// Should not panic
	logger.Debug().Print("test")
	logger.Info().Printf("test %d", 123)
	logger.Warn().Println("test")
}

func testUnwrap(t *testing.T) {
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
}

func testErrorHandling(t *testing.T) {
	// Create an invalid config
	cfg := &zap.Config{
		Encoding: "invalid-encoding",
	}

	_, err := slogzap.New(cfg)
	if err == nil {
		t.Error("expected error for invalid config")
	}
}

func testEnabledNilLogger(t *testing.T) {
	// Test nil logger
	var nilLogger *slogzap.Logger
	if nilLogger.Enabled() {
		t.Error("nil logger should not be enabled")
	}

	// Test logger with nil internal logger
	logger := &slogzap.Logger{}
	if logger.Enabled() {
		t.Error("logger with nil internal logger should not be enabled")
	}
}

func testWithLevelSameLevel(t *testing.T) {
	logger, _ := slogzap.New(nil)
	zapLogger := logger.(*slogzap.Logger)

	// First set the logger to a valid level
	infoLogger := zapLogger.WithLevel(slog.Info).(*slogzap.Logger)

	// Get current level
	currentLevel := infoLogger.Level()
	if currentLevel != slog.Info {
		t.Fatalf("expected Info level, got %v", currentLevel)
	}

	// WithLevel with same level should return same instance
	sameLogger := infoLogger.WithLevel(currentLevel)
	if sameLogger != infoLogger {
		t.Error("WithLevel with same level should return the same logger instance")
	}
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

func TestMapToZapLevel_InvalidLevel(t *testing.T) {
	// This tests the default case in mapToZapLevel (lines 249-251)
	// by creating a logger with an invalid level through reflection
	// or by testing the Enabled() method with such a logger

	logger, _ := slogzap.New(nil)
	zapLogger := logger.(*slogzap.Logger)

	// Test that invalid level causes panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid log level")
		}
	}()

	// Create a logger with an invalid level (100 is not a valid slog level)
	invalidLogger := &slogzap.Logger{}
	// Copy the internal logger but set an invalid level
	*invalidLogger = *zapLogger
	invalidLogger.Loglet = zapLogger.Loglet.WithLevel(slog.LogLevel(100))

	// This should panic
	invalidLogger.Enabled()
}
