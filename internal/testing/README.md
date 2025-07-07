# slog Testing Utilities

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
  implementations
- **Concurrency Utilities** - Tools for testing thread safety

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
        NewLogger: func() slog.Logger {
            return myhandler.New()
        },
        // Skip tests that might not apply
        SkipPanicTests:   true, // if handler exits on panic
        SkipEnabledTests: false, // if handler supports Enabled()
    }

    compliance.Run(t)
}
```

### Testing Concurrency

```go
func TestHandlerConcurrency(t *testing.T) {
    logger := myhandler.New()

    // Run concurrent logging test
    test := slogtest.DefaultConcurrencyTest() // 10 goroutines, 100 ops each
    slogtest.RunConcurrentTest(t, logger, test)

    // Test concurrent field operations
    slogtest.TestConcurrentFields(t, func() slog.Logger {
        return myhandler.New()
    })
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

## API Reference

### Test Logger

- `NewLogger() *Logger` - Creates a new test logger
- `GetMessages() []Message` - Returns all recorded messages
- `Clear()` - Clears all recorded messages

### Assertions

- `AssertMessage(t, msg, level, text)` - Verifies message level and text
- `AssertField(t, msg, key, value)` - Verifies a field exists with expected
  value
- `AssertNoField(t, msg, key)` - Verifies a field does not exist
- `AssertMessageCount(t, messages, count)` - Verifies the number of messages

### Test Functions

- `TestLevelMethods(t, newLogger)` - Tests all level methods (Debug, Info, etc.)
- `TestFieldMethods(t, newLogger)` - Tests WithField and WithFields
- `TestWithField(t, logger)` - Tests WithField behavior
- `TestWithFields(t, logger)` - Tests WithFields behavior
- `TestWithStack(t, logger)` - Tests WithStack behavior

### Compliance Testing

- `ComplianceTest.Run(t)` - Runs full compliance test suite

### Concurrency Testing

- `RunConcurrentTest(t, logger, test)` - Tests concurrent logging
- `TestConcurrentFields(t, newLogger)` - Tests concurrent field operations
- `DefaultConcurrencyTest()` - Returns default concurrency test config

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
