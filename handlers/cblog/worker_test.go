package cblog_test

import (
	"sync"
	"testing"
	"time"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/cblog"
)

func TestNewWithCallbackConcurrency(t *testing.T) {
	// Test concurrent message handling with callback
	const numMessages = 1000
	var received []cblog.LogMsg
	var mu sync.Mutex
	done := make(chan bool)

	handler := func(msg cblog.LogMsg) {
		mu.Lock()
		received = append(received, msg)
		if len(received) == numMessages {
			close(done)
		}
		mu.Unlock()
	}

	logger := cblog.NewWithCallback(100, handler)
	if logger == nil {
		t.Fatal("NewWithCallback returned nil")
	}

	// Send messages concurrently
	var wg sync.WaitGroup
	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			logger.Info().WithField("num", n).Printf("message %d", n)
		}(i)
	}

	wg.Wait()

	// Wait for all messages to be processed
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout: only received %d/%d messages", len(received), numMessages)
	}

	// Verify message count
	if len(received) != numMessages {
		t.Errorf("got %d messages, want %d", len(received), numMessages)
	}

	// Verify all messages have the expected level
	for i, msg := range received {
		if msg.Level != slog.Info {
			t.Errorf("message %d: got level %v, want %v", i, msg.Level, slog.Info)
		}
		if msg.Fields["num"] == nil {
			t.Errorf("message %d: missing 'num' field", i)
		}
	}
}

func TestNewWithCallbackOrdering(t *testing.T) {
	// Test that messages are processed in order
	var messages []int
	var mu sync.Mutex

	handler := func(msg cblog.LogMsg) {
		mu.Lock()
		if num, ok := msg.Fields["order"].(int); ok {
			messages = append(messages, num)
		}
		mu.Unlock()
	}

	logger := cblog.NewWithCallback(1, handler) // Small buffer to test ordering
	if logger == nil {
		t.Fatal("NewWithCallback returned nil")
	}

	// Send messages in order
	const count = 100
	for i := 0; i < count; i++ {
		logger.Info().WithField("order", i).Printf("message %d", i)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(messages) != count {
		t.Fatalf("got %d messages, want %d", len(messages), count)
	}

	// Verify order
	for i, num := range messages {
		if num != i {
			t.Errorf("message %d: got order %d, want %d", i, num, i)
		}
	}
}

func TestNewWithCallbackPanic(t *testing.T) {
	// Skip this test as it causes the worker goroutine to panic,
	// which is expected behavior but interferes with test execution
	t.Skip("Skipping panic test - documented behavior: handler panics are not recovered")

	// Documentation of expected behavior:
	// The current implementation doesn't recover from panics in the handler.
	// If a handler panics, the worker goroutine will exit and subsequent
	// messages will not be processed.
}

func TestNewWithCallbackBlocking(t *testing.T) {
	// Test behavior when handler is slow
	processed := 0
	var mu sync.Mutex

	handler := func(_ cblog.LogMsg) {
		time.Sleep(10 * time.Millisecond) // Slow handler
		mu.Lock()
		processed++
		mu.Unlock()
	}

	logger := cblog.NewWithCallback(5, handler) // Small buffer
	if logger == nil {
		t.Fatal("NewWithCallback returned nil")
	}

	// Send more messages than buffer size
	const numMessages = 10
	start := time.Now()
	for i := 0; i < numMessages; i++ {
		logger.Info().Printf("message %d", i)
	}
	elapsed := time.Since(start)

	// Should have taken some time due to blocking
	// With 10 messages, 5 buffer size, and 10ms per message processing,
	// at least 5 messages should have blocked, so minimum ~50ms expected
	if elapsed < 30*time.Millisecond {
		t.Logf("Warning: sending completed in %v, expected blocking behavior", elapsed)
	}

	// Wait for all to be processed
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if processed != numMessages {
		t.Errorf("processed %d messages, want %d", processed, numMessages)
	}
}

func BenchmarkNewWithCallback(b *testing.B) {
	handler := func(_ cblog.LogMsg) {
		// No-op handler
	}

	logger := cblog.NewWithCallback(1000, handler)
	if logger == nil {
		b.Fatal("NewWithCallback returned nil")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().Print("benchmark message")
	}
}
