package cblog_test

import (
	"runtime"
	"testing"
	"time"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/cblog"
)


func TestFinaliserClosesInternalChannel(t *testing.T) {
	// Test that finaliser closes internally created channels
	type result struct {
		closed bool
		count  int
	}
	resultChan := make(chan result, 1)

	func() {
		logger, logCh := cblog.New(nil)

		// Set up goroutine to detect channel closure and count messages
		go func() {
			count := 0
			for range logCh {
				count++
			}
			resultChan <- result{closed: true, count: count}
		}()

		// Send test messages
		const expectedMessages = 3
		logger.Info().Print("test message 1")
		logger.Debug().Print("test message 2")
		logger.Warn().Print("test message 3")

		// logger goes out of scope here
	}()

	// Force garbage collection
	runtime.GC()
	runtime.Gosched()

	// Check if channel was closed and verify message count
	select {
	case res := <-resultChan:
		if !res.closed {
			t.Error("channel was not closed")
		}
		if res.count != 3 {
			t.Errorf("expected 3 messages, got %d", res.count)
		}
		// Success - channel was closed by finaliser with correct message count
	case <-time.After(2 * time.Second):
		t.Error("finaliser did not close internal channel")
	}
}

func TestFinaliserDoesNotCloseExternalChannel(t *testing.T) {
	// Test that finaliser does NOT close externally provided channels
	ch := make(chan cblog.LogMsg, 10)
	type result struct {
		closed bool
		count  int
	}
	resultChan := make(chan result, 1)

	// Monitor the channel
	go func() {
		count := 0
		for msg := range ch {
			count++
			// Exit after receiving expected messages plus one manual message
			if count == 3 && msg.Message == "manual message" {
				resultChan <- result{closed: false, count: count}
				return
			}
		}
		// Channel was closed unexpectedly
		resultChan <- result{closed: true, count: count}
	}()

	func() {
		logger, _ := cblog.New(ch)

		// Send test messages
		logger.Info().Print("test message 1")
		logger.Debug().Print("test message 2")

		// logger goes out of scope here
	}()

	// Force garbage collection
	runtime.GC()
	runtime.Gosched()

	// Give time for any potential finaliser to run
	time.Sleep(500 * time.Millisecond)

	// Verify we can still send to the channel
	select {
	case ch <- cblog.LogMsg{Level: slog.Info, Message: "manual message"}:
		// Success - channel is still open
	default:
		t.Error("channel appears to be closed")
	}

	// Check the result
	select {
	case res := <-resultChan:
		if res.closed {
			t.Error("finaliser incorrectly closed external channel")
		}
		if res.count != 3 {
			t.Errorf("expected 3 messages, got %d", res.count)
		}
		// Success - received all messages and channel remains open
	case <-time.After(time.Second):
		t.Error("timeout waiting for result")
	}

	close(ch) // Clean up
}
