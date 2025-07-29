package slog_test

import (
	"context"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	slogtest "darvaza.org/slog/internal/testing"
)

func TestWithLogger(t *testing.T) {
	t.Run("StoreLogger", testWithLoggerStore)
	t.Run("NilLogger", testWithLoggerNil)
	t.Run("ContextPropagation", testWithLoggerPropagation)
}

func testWithLoggerStore(t *testing.T) {
	logger := slogtest.NewLogger()
	ctx := context.Background()

	// Store logger in context
	newCtx := slog.WithLogger(ctx, logger)

	// Verify context was modified
	if newCtx == ctx {
		t.Error("WithLogger should return new context")
	}

	// Verify logger can be retrieved
	retrieved, ok := slog.GetLogger(newCtx)
	if !ok {
		t.Fatal("GetLogger should return true for context with logger")
	}
	if retrieved != logger {
		t.Error("Retrieved logger should match stored logger")
	}
}

func testWithLoggerNil(t *testing.T) {
	ctx := context.Background()

	// Store nil logger
	newCtx := slog.WithLogger(ctx, nil)

	// Check what actually happens with nil logger
	retrieved, ok := slog.GetLogger(newCtx)
	if ok && retrieved != nil {
		t.Error("Nil logger should result in nil retrieval")
	}
	// Note: The behaviour with nil might depend on core.NewContextKey implementation
	// This test validates the actual behaviour rather than assuming
}

func testWithLoggerPropagation(t *testing.T) {
	logger1 := slogtest.NewLogger()
	logger2 := slogtest.NewLogger()

	ctx := context.Background()

	// First logger
	ctx1 := slog.WithLogger(ctx, logger1)
	retrieved1, ok := slog.GetLogger(ctx1)
	if !ok || retrieved1 != logger1 {
		t.Fatal("First logger not stored correctly")
	}

	// Override with second logger
	ctx2 := slog.WithLogger(ctx1, logger2)
	retrieved2, ok := slog.GetLogger(ctx2)
	if !ok || retrieved2 != logger2 {
		t.Fatal("Second logger not stored correctly")
	}

	// Original context should still have first logger
	retrieved1Again, ok := slog.GetLogger(ctx1)
	if !ok || retrieved1Again != logger1 {
		t.Error("Original context should be unchanged")
	}
}

func TestGetLogger(t *testing.T) {
	t.Run("EmptyContext", testGetLoggerEmpty)
	t.Run("WithLogger", testGetLoggerWithLogger)
	t.Run("ContextChain", testGetLoggerChain)
}

func testGetLoggerEmpty(t *testing.T) {
	ctx := context.Background()

	// Empty context should return false
	logger, ok := slog.GetLogger(ctx)
	if ok {
		t.Error("GetLogger should return false for empty context")
	}
	if logger != nil {
		t.Error("Logger should be nil for empty context")
	}
}

func testGetLoggerWithLogger(t *testing.T) {
	logger := slogtest.NewLogger()
	ctx := slog.WithLogger(context.Background(), logger)

	// Should retrieve the stored logger
	retrieved, ok := slog.GetLogger(ctx)
	if !ok {
		t.Fatal("GetLogger should return true")
	}
	if retrieved != logger {
		t.Error("Retrieved logger should match stored logger")
	}
}

func testGetLoggerChain(t *testing.T) {
	logger1 := slogtest.NewLogger()
	logger2 := slogtest.NewLogger()

	// Create context chain
	ctx := context.Background()
	otherKey := core.NewContextKey[string]("other")
	ctx = otherKey.WithValue(ctx, "value") // Add unrelated value
	ctx = slog.WithLogger(ctx, logger1)
	moreKey := core.NewContextKey[string]("more")
	ctx = moreKey.WithValue(ctx, "data") // Add another unrelated value
	ctx = slog.WithLogger(ctx, logger2)  // Override logger

	// Should get the most recent logger
	retrieved, ok := slog.GetLogger(ctx)
	if !ok {
		t.Fatal("GetLogger should return true")
	}
	if retrieved != logger2 {
		t.Error("Should retrieve most recent logger")
	}

	// Verify unrelated context values still exist
	if val, ok := otherKey.Get(ctx); !ok || val != "value" {
		t.Error("Unrelated context values should be preserved")
	}
	if val, ok := moreKey.Get(ctx); !ok || val != "data" {
		t.Error("Unrelated context values should be preserved")
	}
}

func TestContextKeyIsolation(t *testing.T) {
	logger := slogtest.NewLogger()
	ctx := context.Background()

	// Store logger
	ctx = slog.WithLogger(ctx, logger)

	// Try to access with different key types - should not interfere
	loggerKey := core.NewContextKey[string]("logger")
	numberKey := core.NewContextKey[slog.Logger]("numberKey")

	ctx = loggerKey.WithValue(ctx, "not-a-logger")
	ctx = numberKey.WithValue(ctx, logger) // different key

	// Should still retrieve correct logger
	retrieved, ok := slog.GetLogger(ctx)
	if !ok {
		t.Fatal("GetLogger should still work")
	}
	if retrieved != logger {
		t.Error("Logger retrieval should be isolated from other context keys")
	}

	// Verify our interference values exist
	if val, ok := loggerKey.Get(ctx); !ok || val != "not-a-logger" {
		t.Error("String key should be independent")
	}
	if val, ok := numberKey.Get(ctx); !ok || val != logger {
		t.Error("Numeric key should be independent")
	}
}
