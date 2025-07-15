# `slog` Testing Utilities

This package provides shared test utilities for testing slog handler
implementations. It helps reduce code duplication and ensures consistent testing
patterns across all handlers.

## Overview

The testing package includes:

- **Test Logger** - A logger implementation that records messages for
  verification
- **Assertion Helpers** - Functions to verify message properties and fields
- **Standalone Test Functions** - Reusable tests for common interface methods
- **Compliance Test Suite** - Comprehensive tests for slog.Logger
  implementations including basic concurrency safety
- **Stress Test Suite** - High-volume concurrent testing for performance and
  stability
- **Bidirectional Testing** - Support for adapters that convert between
  logging libraries
- **Options Architecture** - Flexible configuration for different handler
  capabilities

## Usage

### Basic Testing with Test Logger

```go
import (
    "testing"
    "darvaza.org/slog"
    slogtest "darvaza.org/slog/internal/testing"
)

func TestMyHandler(t *testing.T) {
    // Create a test logger that records messages
    recorder := slogtest.NewLogger()

    // Use it with your handler
    handler := myhandler.New(recorder)

    // Perform logging operations
    handler.Info().
        WithField("user", "john").
        Print("User action")

    // Verify the results
    msgs := recorder.GetMessages()
    slogtest.AssertMessageCount(t, msgs, 1)
    slogtest.AssertMessage(t, msgs[0], slog.Info, "User action")
    slogtest.AssertField(t, msgs[0], "user", "john")
}
```

### Testing Interface Methods

The package provides standalone test functions for common slog.Logger methods:

```go
func TestHandlerMethods(t *testing.T) {
    // Test WithField method
    t.Run("WithField", func(t *testing.T) {
        logger := myhandler.New()
        slogtest.TestWithField(t, logger)
    })

    // Test WithFields method
    t.Run("WithFields", func(t *testing.T) {
        logger := myhandler.New()
        slogtest.TestWithFields(t, logger)
    })

    // Test WithStack method
    t.Run("WithStack", func(t *testing.T) {
        logger := myhandler.New()
        slogtest.TestWithStack(t, logger)
    })
}
```

### Running the Compliance Test Suite

For comprehensive testing of a handler implementation:

```go
func TestHandlerCompliance(t *testing.T) {
    compliance := slogtest.ComplianceTest{
        FactoryOptions: slogtest.FactoryOptions{
            NewLogger: func() slog.Logger {
                return myhandler.New()
            },
        },
        // Skip tests that might not apply
        SkipPanicTests:   true, // if handler exits on panic
        SkipEnabledTests: false, // if handler supports Enabled()
    }

    compliance.Run(t)
}
```

For bidirectional adapters that can create loggers backed by recorders:

```go
func TestBidirectionalHandlerCompliance(t *testing.T) {
    compliance := slogtest.ComplianceTest{
        FactoryOptions: slogtest.FactoryOptions{
            NewLogger: func() slog.Logger {
                return myhandler.New()
            },
            // NewLoggerWithRecorder enables field isolation tests
            // Without this, testFieldChainIsolation subtests will be skipped
            NewLoggerWithRecorder: func(recorder slog.Logger) slog.Logger {
                // Create handler that writes to recorder
                return myhandler.NewWithBackend(recorder)
            },
        },
        AdapterOptions: slogtest.AdapterOptions{
            LevelExceptions: map[slog.LogLevel]slog.LogLevel{
                slog.Warn: slog.Info, // if handler maps Warn to Info
            },
        },
    }

    compliance.Run(t)
}
```

### Running Stress Tests

For performance and stability testing:

```go
func TestHandlerStress(t *testing.T) {
    suite := slogtest.StressTestSuite{
        NewLogger: func() slog.Logger {
            return myhandler.New()
        },
        // Optional: for handlers that can write to a recorder
        NewLoggerWithRecorder: func(recorder slog.Logger) slog.Logger {
            return myhandler.NewWithBackend(recorder)
        },
    }

    suite.Run(t)
}
```

For custom stress scenarios:

```go
func TestCustomStress(t *testing.T) {
    logger := myhandler.New()

    // High-volume stress test
    slogtest.RunStressTest(t, logger, slogtest.HighVolumeStressTest())

    // Memory pressure test
    slogtest.RunStressTest(t, logger, slogtest.MemoryPressureStressTest())

    // Duration-based test
    stress := slogtest.DurationBasedStressTest(time.Second)
    slogtest.RunStressTest(t, logger, stress)
}
```

### Testing Bidirectional Adapters

```go
func TestMyBidirectionalAdapter(t *testing.T) {
    // Factory function that creates adapter using given backend
    factory := func(backend slog.Logger) slog.Logger {
        // Create adapter that uses backend
        externalLogger := mylib.NewLogger()
        mylib.SetOutput(externalLogger, backend)

        // Return slog adapter wrapping external logger
        return myadapter.New(externalLogger)
    }

    // Test adapter that converts between logging libraries
    slogtest.TestBidirectional(t, "MyAdapter", factory)
}
```

For adapters with known limitations (e.g., missing log levels):

```go
func TestAdapterWithLimitations(t *testing.T) {
    factory := func(backend slog.Logger) slog.Logger {
        // Adapter that doesn't support Warn level
        return limitedadapter.New(backend)
    }

    // Define expected level mappings
    opts := &slogtest.BidirectionalTestOptions{
        LevelExceptions: map[slog.LogLevel]slog.LogLevel{
            slog.Warn: slog.Info, // Warn gets mapped to Info
        },
    }

    // Test with options to handle known limitations
    slogtest.TestBidirectionalWithOptions(t, "LimitedAdapter", factory, opts)
}
```

For adapters that apply level-based filtering:

```go
func TestAdapterWithLevelFiltering(t *testing.T) {
    factory := func(backend slog.Logger) slog.Logger {
        // Adapter that filters out Debug messages based on minimum level
        return filteredadapter.New(backend, slog.Info) // Only Info and above
    }

    opts := &slogtest.BidirectionalTestOptions{
        LevelExceptions: map[slog.LogLevel]slog.LogLevel{
            slog.Debug: slog.UndefinedLevel, // Debug messages are filtered out
        },
    }

    slogtest.TestBidirectionalWithOptions(t, "FilteredAdapter", factory, opts)
}
```

## API Reference

### Test Logger

- `NewLogger() *Logger` - Creates a new test logger
- `GetMessages() []Message` - Returns all recorded messages
- `Clear()` - Clears all recorded messages
- `Message.String() string` - Returns formatted string representation

### Assertions

- `AssertMessage(t, msg, level, text)` - Verifies message level and text
- `AssertField(t, msg, key, value)` - Verifies a field exists with expected
  value
- `AssertNoField(t, msg, key)` - Verifies a field does not exist
- `AssertMessageCount(t, messages, count)` - Verifies the number of messages

### Test Functions

- `TestLevelMethods(t, newLogger)` - Tests all level methods (Debug, Info, etc.)
- `TestFieldMethods(t, newLogger)` - Tests WithField and WithFields
- `TestWithField(t, logger)` - Tests WithField behaviour.
- `TestWithFields(t, logger)` - Tests WithFields behaviour.
- `TestWithStack(t, logger)` - Tests WithStack behaviour.

### Compliance Testing

The `ComplianceTest` struct provides comprehensive testing:

```go
type ComplianceTest struct {
    AdapterOptions                    // Level transformation exceptions
    FactoryOptions                    // Logger creation functions
    SkipEnabledTests bool            // Skip Enabled() tests
    SkipPanicTests   bool            // Skip Fatal/Panic tests
}
```

- `ComplianceTest.Run(t)` - Runs full compliance test suite including:
  - Interface implementation verification
  - Level method testing
  - Field method testing and immutability
  - Print method testing
  - Enabled method testing (if applicable)
  - WithStack functionality
  - Logger branching and immutability
  - Basic concurrency safety

### Stress Testing

The `StressTestSuite` provides comprehensive performance testing:

```go
type StressTestSuite struct {
    NewLogger func() slog.Logger
    NewLoggerWithRecorder func(slog.Logger) slog.Logger

    // Skip specific stress tests
    SkipHighVolume      bool
    SkipMemoryPressure  bool
    SkipDurationBased   bool
    SkipConcurrentField bool
}
```

- `DefaultStressTest()` - Basic stress test (10 goroutines, 100 ops)
- `HighVolumeStressTest()` - High-volume test (50 goroutines, 1000 ops)
- `MemoryPressureStressTest()` - Tests with large messages and many fields
- `DurationBasedStressTest(duration)` - Runs for specified duration
- `RunStressTest(t, logger, test)` - Run a specific stress test
- `RunStressTestWithOptions(t, logger, test, opts)` - Advanced stress testing

### Bidirectional Testing

- `TestBidirectional(t, name, fn)` - Tests bidirectional adapter implementations
  - `fn` should return a logger that uses the given logger as backend
  - Tests message preservation, field handling, and level mapping
  - Verifies round-trip conversion maintains data integrity

- `TestBidirectionalWithOptions(t, name, fn, opts)` - Tests with level mapping
  exceptions
  - Handles adapters with known limitations (e.g., missing log levels)
  - `opts.LevelExceptions` maps expected level transformations
  - Useful for adapters like logr that don't support all slog levels

### Helper Utilities

- `TransformMessages(messages, opts) []Message` - Apply level transformations
  - Messages mapped to `slog.UndefinedLevel` are omitted from the result
- `CompareMessages(first, second) (onlyFirst, onlySecond, both []Message)` -
  Set-based comparison
- `RunWithLogger(t, name, logger, fn)` - Run subtest with logger
- `RunWithLoggerFactory(t, name, newLogger, fn)` - Run subtest with fresh
  logger

## Options Architecture

The testing package uses an embedded options pattern to provide flexible
configuration:

### AdapterOptions

Base options for testing adapters with level transformation support:

```go
type AdapterOptions struct {
    LevelExceptions map[slog.LogLevel]slog.LogLevel
}
```

Level transformations support mapping to `slog.UndefinedLevel` to indicate that
messages at that level should be filtered out entirely. This is useful for
adapters that apply level-based filtering where certain levels are discarded
rather than remapped to a different level.

### FactoryOptions

Factory functions for creating loggers:

```go
type FactoryOptions struct {
    NewLogger             func() slog.Logger
    NewLoggerWithRecorder func(slog.Logger) slog.Logger
}
```

- `NewLogger` - Creates a fresh logger instance
- `NewLoggerWithRecorder` - Creates a logger backed by the provided recorder
  (for bidirectional adapters)

### Test-Specific Options

Each test type embeds the appropriate base options:

- `BidirectionalTestOptions` - Embeds `AdapterOptions`
- `ComplianceTest` - Embeds both `AdapterOptions` and `FactoryOptions`
- `StressTestOptions` - Configuration for stress testing
- `ConcurrencyTestOptions` - Configuration for concurrency testing utilities

This design allows code reuse while maintaining type safety and clear intent.

## Design Principles

1. **Standalone Functions** - Test functions accept logger instances directly
   rather than factory functions where appropriate, reducing complexity

2. **Reduced Complexity** - Complex test scenarios are broken into smaller,
   focused helper functions

3. **Thread Safety** - All utilities are designed to be safe for concurrent use

4. **Flexibility** - Tests can be run individually or as part of comprehensive
   suites

## Examples

See `example_test.go` for complete examples of using these utilities in your
handler tests.
