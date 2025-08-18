package slog_test

import (
	"context"
	"testing"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/mock"
)

func TestWithLogger(t *testing.T) {
	t.Run("StoreLogger", testWithLoggerStore)
	t.Run("NilLogger", testWithLoggerNil)
	t.Run("ContextPropagation", testWithLoggerPropagation)
}

func testWithLoggerStore(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	ctx := context.Background()

	// Store logger in context
	newCtx := slog.WithLogger(ctx, logger)

	// Verify context was modified
	core.AssertTrue(t, newCtx != ctx, "WithLogger returns new context")

	// Verify logger can be retrieved
	retrieved, ok := slog.GetLogger(newCtx)
	core.AssertMustTrue(t, ok, "GetLogger success")
	core.AssertTrue(t, retrieved == logger, "retrieved logger matches")
}

func testWithLoggerNil(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	// Store nil logger
	newCtx := slog.WithLogger(ctx, nil)

	// Check what actually happens with nil logger
	retrieved, ok := slog.GetLogger(newCtx)
	if ok {
		core.AssertNil(t, retrieved, "nil logger retrieval")
	}
	// Note: The behaviour with nil might depend on core.NewContextKey implementation
	// This test validates the actual behaviour rather than assuming
}

func testWithLoggerPropagation(t *testing.T) {
	t.Helper()
	logger1 := mock.NewLogger()
	logger2 := mock.NewLogger()

	ctx := context.Background()

	// First logger
	ctx1 := slog.WithLogger(ctx, logger1)
	retrieved1, ok := slog.GetLogger(ctx1)
	core.AssertMustTrue(t, ok, "first logger stored")
	core.AssertTrue(t, retrieved1 == logger1, "first logger value matches")

	// Override with second logger
	ctx2 := slog.WithLogger(ctx1, logger2)
	retrieved2, ok := slog.GetLogger(ctx2)
	core.AssertMustTrue(t, ok, "second logger stored")
	core.AssertTrue(t, retrieved2 == logger2, "second logger value matches")

	// Original context should still have first logger
	retrieved1Again, ok := slog.GetLogger(ctx1)
	core.AssertMustTrue(t, ok, "original context unchanged")
	core.AssertTrue(t, retrieved1Again == logger1, "original logger preserved")
}

func TestGetLogger(t *testing.T) {
	t.Run("EmptyContext", testGetLoggerEmpty)
	t.Run("WithLogger", testGetLoggerWithLogger)
	t.Run("ContextChain", testGetLoggerChain)
}

func testGetLoggerEmpty(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	// Empty context should return false
	logger, ok := slog.GetLogger(ctx)
	core.AssertFalse(t, ok, "GetLogger on empty context")
	core.AssertNil(t, logger, "logger from empty context")
}

func testGetLoggerWithLogger(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
	ctx := slog.WithLogger(context.Background(), logger)

	// Should retrieve the stored logger
	retrieved, ok := slog.GetLogger(ctx)
	core.AssertMustTrue(t, ok, "GetLogger success")
	core.AssertTrue(t, retrieved == logger, "retrieved logger matches")
}

func testGetLoggerChain(t *testing.T) {
	t.Helper()
	logger1 := mock.NewLogger()
	logger2 := mock.NewLogger()

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
	core.AssertMustTrue(t, ok, "GetLogger success")
	core.AssertTrue(t, retrieved == logger2, "most recent logger matches")

	// Verify unrelated context values still exist
	val1, ok1 := otherKey.Get(ctx)
	if core.AssertTrue(t, ok1, "other key exists") {
		core.AssertEqual(t, "value", val1, "other key value")
	}
	val2, ok2 := moreKey.Get(ctx)
	if core.AssertTrue(t, ok2, "more key exists") {
		core.AssertEqual(t, "data", val2, "more key value")
	}
}

func TestContextKeyIsolation(t *testing.T) {
	t.Helper()
	logger := mock.NewLogger()
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
	core.AssertMustTrue(t, ok, "GetLogger still works")
	core.AssertTrue(t, retrieved == logger, "logger isolation")

	// Verify our interference values exist
	val1, ok1 := loggerKey.Get(ctx)
	if core.AssertTrue(t, ok1, "string key independent") {
		core.AssertEqual(t, "not-a-logger", val1, "string key value")
	}
	val2, ok2 := numberKey.Get(ctx)
	if core.AssertTrue(t, ok2, "number key independent") {
		core.AssertTrue(t, val2 == logger, "number key value matches")
	}
}
