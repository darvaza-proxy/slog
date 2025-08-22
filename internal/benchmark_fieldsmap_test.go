package internal

import (
	"testing"

	"darvaza.org/slog"
)

// BenchmarkFieldsIteration benchmarks the traditional iterator pattern
func BenchmarkFieldsIteration(b *testing.B) {
	// Create a loglet with fields
	var base Loglet
	l1 := base.WithField("service", "api")
	l2 := l1.WithField("version", "1.0")
	l3 := l2.WithField("user_id", 12345)
	l4 := l3.WithField("request_id", "req-abc-123")
	l5 := l4.WithField("operation", "create_user")

	b.ResetTimer()
	for range b.N {
		// Traditional iterator pattern
		fields := make(map[string]any, l5.FieldsCount())
		iter := l5.Fields()
		for iter.Next() {
			k, v := iter.Field()
			fields[k] = v
		}
		_ = fields
	}
}

// BenchmarkFieldsMap benchmarks the new FieldsMap pattern
func BenchmarkFieldsMap(b *testing.B) {
	// Create a loglet with fields
	var base Loglet
	l1 := base.WithField("service", "api")
	l2 := l1.WithField("version", "1.0")
	l3 := l2.WithField("user_id", 12345)
	l4 := l3.WithField("request_id", "req-abc-123")
	l5 := l4.WithField("operation", "create_user")

	b.ResetTimer()
	for range b.N {
		// New FieldsMap pattern
		fields := l5.FieldsMap()
		_ = fields
	}
}

// BenchmarkFieldsMapCached benchmarks repeated calls to FieldsMap (cached)
func BenchmarkFieldsMapCached(b *testing.B) {
	// Create a loglet with fields
	var base Loglet
	l1 := base.WithField("service", "api")
	l2 := l1.WithField("version", "1.0")
	l3 := l2.WithField("user_id", 12345)
	l4 := l3.WithField("request_id", "req-abc-123")
	l5 := l4.WithField("operation", "create_user")

	// Prime the cache
	_ = l5.FieldsMap()

	b.ResetTimer()
	for range b.N {
		// Cached FieldsMap calls
		fields := l5.FieldsMap()
		_ = fields
	}
}

// createBenchmarkLoglet creates a loglet with 10 fields for benchmarking
func createBenchmarkLoglet() Loglet {
	var base Loglet
	l1 := base.WithField("service", "api")
	l2 := l1.WithField("version", "1.0")
	l3 := l2.WithField("user_id", 12345)
	l4 := l3.WithField("request_id", "req-abc-123")
	l5 := l4.WithField("operation", "create_user")
	l6 := l5.WithField("timestamp", "2025-01-01T00:00:00Z")
	l7 := l6.WithField("ip_address", "192.168.1.1")
	l8 := l7.WithField("user_agent", "Mozilla/5.0")
	l9 := l8.WithField("session_id", "session-xyz-789")
	return l9.WithField("trace_id", "trace-def-456")
}

// BenchmarkIteratorPattern benchmarks field access using iterator
func BenchmarkIteratorPattern(b *testing.B) {
	loglet := createBenchmarkLoglet()
	for range b.N {
		fields := make(map[string]any, loglet.FieldsCount())
		iter := loglet.Fields()
		for iter.Next() {
			k, v := iter.Field()
			fields[k] = v
		}
		_ = fields
	}
}

// BenchmarkFieldsMapFirstCall benchmarks first call to FieldsMap
func BenchmarkFieldsMapFirstCall(b *testing.B) {
	for range b.N {
		freshLoglet := createBenchmarkLoglet()
		fields := freshLoglet.FieldsMap()
		_ = fields
	}
}

// BenchmarkFieldsMapDelegation benchmarks delegation vs iteration for field-less loglets
func BenchmarkFieldsMapDelegation(b *testing.B) {
	var base Loglet
	l1 := base.WithField("key1", "value1")
	l2 := l1.WithField("key2", "value2")
	child := l2.WithLevel(slog.Info) // No fields, should delegate

	b.ResetTimer()
	for range b.N {
		_ = child.FieldsMap()
	}
}

// BenchmarkFieldsMapMultiLevelDelegation benchmarks delegation through multiple intermediate loglets
func BenchmarkFieldsMapMultiLevelDelegation(b *testing.B) {
	var base Loglet
	l1 := base.WithField("service", "api")
	l2 := l1.WithField("version", "1.0")
	intermediate1 := l2.WithLevel(slog.Info)    // No fields, delegates
	intermediate2 := intermediate1.WithStack(1) // No fields, delegates
	child := intermediate2.Copy()               // Final loglet that should delegate

	b.ResetTimer()
	for range b.N {
		_ = child.FieldsMap()
	}
}

// benchmarkFieldsMapIterationHelper benchmarks traditional iteration
func benchmarkFieldsMapIterationHelper(b *testing.B) {
	var base Loglet
	l1 := base.WithField("key1", "value1")
	l2 := l1.WithField("key2", "value2")
	child := l2.WithLevel(slog.Info) // No fields, should delegate

	b.ResetTimer()
	for range b.N {
		// Simulate the old behaviour (without delegation)
		fields := make(map[string]any, child.FieldsCount())
		iter := child.Fields()
		for iter.Next() {
			k, v := iter.Field()
			fields[k] = v
		}
		_ = fields
	}
}

// BenchmarkFieldsMapDelegationVsIteration compares delegation performance vs traditional iteration
func BenchmarkFieldsMapDelegationVsIteration(b *testing.B) {
	b.Run("Delegation", BenchmarkFieldsMapDelegation)
	b.Run("Iteration", benchmarkFieldsMapIterationHelper)
}

// BenchmarkFieldsMapCopy benchmarks the FieldsMapCopy method
func BenchmarkFieldsMapCopy(b *testing.B) {
	var base Loglet
	l1 := base.WithField("key1", "value1")
	l2 := l1.WithField("key2", "value2")
	l3 := l2.WithField("key3", "value3")

	b.ResetTimer()
	for range b.N {
		_ = l3.FieldsMapCopy(0)
	}
}

// BenchmarkFieldsMapCopyWithExcess benchmarks FieldsMapCopy with excess capacity
func BenchmarkFieldsMapCopyWithExcess(b *testing.B) {
	var base Loglet
	l1 := base.WithField("key1", "value1")
	l2 := l1.WithField("key2", "value2")
	l3 := l2.WithField("key3", "value3")

	b.ResetTimer()
	for range b.N {
		_ = l3.FieldsMapCopy(5)
	}
}

// benchmarkFieldsMapCopyFromCachedHelper benchmarks copy from cached source
func benchmarkFieldsMapCopyFromCachedHelper(b *testing.B) {
	var base Loglet
	l1 := base.WithField("key1", "value1")
	l2 := l1.WithField("key2", "value2")
	l3 := l2.WithField("key3", "value3")

	// Prime the cache
	_ = l3.FieldsMap()

	b.ResetTimer()
	for range b.N {
		_ = l3.FieldsMapCopy(0)
	}
}

// benchmarkFieldsMapCopyFromUncachedHelper benchmarks copy from uncached source
func benchmarkFieldsMapCopyFromUncachedHelper(b *testing.B) {
	for range b.N {
		var base Loglet
		l1 := base.WithField("key1", "value1")
		l2 := l1.WithField("key2", "value2")
		l3 := l2.WithField("key3", "value3")
		_ = l3.FieldsMapCopy(0)
	}
}

// BenchmarkFieldsMapCopyVsCached compares FieldsMapCopy performance vs cached FieldsMap
func BenchmarkFieldsMapCopyVsCached(b *testing.B) {
	b.Run("CopyFromCached", benchmarkFieldsMapCopyFromCachedHelper)
	b.Run("CopyFromUncached", benchmarkFieldsMapCopyFromUncachedHelper)
}

// benchmarkManualIterationHelper benchmarks manual iteration to build map
func benchmarkManualIterationHelper(b *testing.B) {
	var base Loglet
	l1 := base.WithField("key1", "value1")
	l2 := l1.WithField("key2", "value2")
	l3 := l2.WithField("key3", "value3")

	b.ResetTimer()
	for range b.N {
		fields := make(map[string]any, l3.FieldsCount())
		iter := l3.Fields()
		for iter.Next() {
			k, v := iter.Field()
			fields[k] = v
		}
		_ = fields
	}
}

// BenchmarkFieldsMapCopyVsIteration compares FieldsMapCopy vs manual iteration
func BenchmarkFieldsMapCopyVsIteration(b *testing.B) {
	b.Run("FieldsMapCopy", BenchmarkFieldsMapCopy)
	b.Run("ManualIteration", benchmarkManualIterationHelper)
}
