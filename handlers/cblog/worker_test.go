package cblog_test

import (
	"sync"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/cblog"
)

func TestNewWithCallbackConcurrency(t *testing.T) {
	const numMessages = 1000

	collector := newCallbackCollector(numMessages)
	logger := cblog.NewWithCallback(100, collector.handle)
	if !core.AssertNotNil(t, logger, "NewWithCallback returned nil") {
		return
	}

	sendNumberedLogs(logger, numMessages)
	waitOrTimeout(collector.done, 5*time.Second)

	received := collector.snapshot()
	if !core.AssertEqual(t, numMessages, len(received), "message count") {
		return
	}
	for i, msg := range received {
		core.AssertEqual(t, slog.Info, msg.Level, "message %d level", i)
		core.AssertNotNil(t, msg.Fields["num"], "message %d 'num' field", i)
	}
}

// callbackCollector accumulates messages from a cblog callback and
// signals done once target messages have been received.
type callbackCollector struct {
	done     chan struct{}
	received []cblog.LogMsg
	mu       sync.Mutex
	target   int
}

func newCallbackCollector(target int) *callbackCollector {
	return &callbackCollector{done: make(chan struct{}), target: target}
}

func (c *callbackCollector) handle(msg cblog.LogMsg) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.received = append(c.received, msg)
	if len(c.received) == c.target {
		close(c.done)
	}
}

func (c *callbackCollector) snapshot() []cblog.LogMsg {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]cblog.LogMsg, len(c.received))
	copy(out, c.received)
	return out
}

// sendNumberedLogs fans out one goroutine per message, each tagging its
// output with the index in a "num" field.
func sendNumberedLogs(logger slog.Logger, numMessages int) {
	var wg sync.WaitGroup
	for i := range numMessages {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			logger.Info().WithField("num", n).Printf("message %d", n)
		}(i)
	}
	wg.Wait()
}

func waitOrTimeout(done <-chan struct{}, d time.Duration) {
	select {
	case <-done:
	case <-time.After(d):
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
	if !core.AssertNotNil(t, logger, "NewWithCallback returned nil") {
		return
	}

	// Send messages in order
	const count = 100
	for i := range count {
		logger.Info().WithField("order", i).Printf("message %d", i)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if !core.AssertEqual(t, count, len(messages), "message count") {
		return
	}

	// Verify order
	for i, num := range messages {
		core.AssertEqual(t, i, num, "message %d order", i)
	}
}

func TestNewWithCallbackBlocking(t *testing.T) {
	// Test behaviour when handler is slow
	processed := 0
	var mu sync.Mutex

	handler := func(_ cblog.LogMsg) {
		time.Sleep(10 * time.Millisecond) // Slow handler
		mu.Lock()
		processed++
		mu.Unlock()
	}

	logger := cblog.NewWithCallback(5, handler) // Small buffer
	if !core.AssertNotNil(t, logger, "NewWithCallback returned nil") {
		return
	}

	// Send more messages than buffer size
	const numMessages = 10
	start := time.Now()
	for i := range numMessages {
		logger.Info().Printf("message %d", i)
	}
	elapsed := time.Since(start)

	// Should have taken some time due to blocking
	// With 10 messages, 5 buffer size, and 10ms per message processing,
	// at least 5 messages should have blocked, so minimum ~50ms expected
	if elapsed < 30*time.Millisecond {
		t.Logf("Warning: sending completed in %v, expected blocking behaviour", elapsed)
	}

	// Wait for all to be processed
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	core.AssertEqual(t, numMessages, processed, "processed message count")
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
	for range b.N {
		logger.Info().Print("benchmark message")
	}
}
