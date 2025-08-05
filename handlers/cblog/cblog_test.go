package cblog_test

import (
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/cblog"
	slogtest "darvaza.org/slog/internal/testing"
)

const (
	testHelloWorld = "hello world"
	testValue      = "value"
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

func TestNew(t *testing.T) {
	t.Run("WithNilChannel", testNewWithNilChannel)
	t.Run("WithBufferedChannel", testNewWithBufferedChannel)
}

func testNewWithNilChannel(t *testing.T) {
	logger, ch := cblog.New(nil)
	if logger == nil {
		t.Fatal("New returned nil logger")
	}
	if ch == nil {
		t.Fatal("New returned nil channel")
	}

	// Test that we can send messages
	logger.Info().Print("test message")

	select {
	case msg := <-ch:
		if msg.Message != "test message" {
			t.Errorf("got message %q, want %q", msg.Message, "test message")
		}
		if msg.Level != slog.Info {
			t.Errorf("got level %v, want %v", msg.Level, slog.Info)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func testNewWithBufferedChannel(t *testing.T) {
	ch := make(chan cblog.LogMsg, 100)
	logger, outCh := cblog.New(ch)
	if logger == nil {
		t.Fatal("New returned nil logger")
	}
	if outCh != ch {
		t.Fatal("New returned different channel")
	}

	// Send multiple messages
	logger.Debug().Print("debug")
	logger.Info().Print("info")
	logger.Warn().Print("warn")

	// Verify messages
	messages := []struct {
		level slog.LogLevel
		msg   string
	}{
		{slog.Debug, "debug"},
		{slog.Info, "info"},
		{slog.Warn, "warn"},
	}

	for _, want := range messages {
		select {
		case got := <-ch:
			if got.Level != want.level {
				t.Errorf("got level %v, want %v", got.Level, want.level)
			}
			if got.Message != want.msg {
				t.Errorf("got message %q, want %q", got.Message, want.msg)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}
	}
}

// Helper function to create a cblog logger that records messages for testing
func newCblogWithRecorder() (slog.Logger, *slogtest.Logger) {
	recorder := slogtest.NewLogger()

	// Create cblog with callback that forwards to recorder
	logger := cblog.NewWithCallback(1000, func(msg cblog.LogMsg) {
		recLogger := recorder.WithLevel(msg.Level)
		if msg.Stack != nil {
			recLogger = recLogger.WithStack(0)
		}
		if msg.Fields != nil {
			recLogger = recLogger.WithFields(msg.Fields)
		}
		recLogger.Print(msg.Message)
	})

	return logger, recorder
}

func TestLoggerLevels(t *testing.T) {
	// Use the standard test function with a factory that includes channel draining
	slogtest.TestLevelMethods(t, func() slog.Logger {
		return makeTestLevelMethodsLogger(t)
	})
}

func makeTestLevelMethodsLogger(t *testing.T) slog.Logger {
	logger, ch := cblog.New(nil)
	done := make(chan struct{})

	// Drain channel in background to prevent blocking
	go func() {
		for {
			select {
			case <-ch:
				// Discard messages
			case <-done:
				return
			}
		}
	}()

	// Ensure cleanup when test completes
	t.Cleanup(func() {
		close(done)
	})

	return logger
}

func TestLoggerPrintMethods(t *testing.T) {
	logger, recorder := newCblogWithRecorder()

	slogtest.RunWithLogger(t, "Print", logger, func(t core.T, logger slog.Logger) {
		testCblogPrint(t, logger, recorder)
	})

	slogtest.RunWithLogger(t, "Println", logger, func(t core.T, logger slog.Logger) {
		testCblogPrintln(t, logger, recorder)
	})

	slogtest.RunWithLogger(t, "Printf", logger, func(t core.T, logger slog.Logger) {
		testCblogPrintf(t, logger, recorder)
	})
}

func testCblogPrint(t core.T, logger slog.Logger, recorder *slogtest.Logger) {
	recorder.Clear()
	logger.Info().Print("hello", " ", "world")

	// Give callback time to process
	time.Sleep(10 * time.Millisecond)

	msgs := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)
	slogtest.AssertMessage(t, msgs[0], slog.Info, testHelloWorld)
}

func testCblogPrintln(t core.T, logger slog.Logger, recorder *slogtest.Logger) {
	recorder.Clear()
	logger.Info().Println("hello", "world")

	time.Sleep(10 * time.Millisecond)

	msgs := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)
	slogtest.AssertMessage(t, msgs[0], slog.Info, testHelloWorld)
}

func testCblogPrintf(t core.T, logger slog.Logger, recorder *slogtest.Logger) {
	recorder.Clear()
	logger.Info().Printf("hello %s", "world")

	time.Sleep(10 * time.Millisecond)

	msgs := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)
	slogtest.AssertMessage(t, msgs[0], slog.Info, testHelloWorld)
}

func TestFieldMethods(t *testing.T) {
	// Use the standard TestFieldMethods which tests both WithField and WithFields
	slogtest.TestFieldMethods(t, func() slog.Logger {
		logger, _ := newCblogWithRecorder()
		return logger
	})
}

func TestLoggerWithStack(t *testing.T) {
	logger, _ := newCblogWithRecorder()
	slogtest.TestWithStack(t, logger)
}

func TestLoggerWithLevel(t *testing.T) {
	logger, ch := cblog.New(nil)

	slogtest.RunWithLogger(t, "ValidLevel", logger, func(t core.T, logger slog.Logger) {
		testCblogValidLevel(t, logger, ch)
	})

	slogtest.RunWithLogger(t, "SameLevel", logger, func(t core.T, logger slog.Logger) {
		testCblogSameLevel(t, logger)
	})

	slogtest.RunWithLogger(t, "InvalidLevel", logger, func(t core.T, logger slog.Logger) {
		testCblogInvalidLevel(t, logger, ch)
	})
}

func testCblogValidLevel(t core.T, logger slog.Logger, ch <-chan cblog.LogMsg) {
	l := logger.WithLevel(slog.Error)
	if l == nil {
		t.Fatal("WithLevel returned nil")
	}
	l.Print("error message")

	select {
	case msg := <-ch:
		if msg.Level != slog.Error {
			t.Errorf("got level %v, want %v", msg.Level, slog.Error)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func testCblogSameLevel(t core.T, logger slog.Logger) {
	l1 := logger.Info()
	l2 := l1.WithLevel(slog.Info)
	if l1 != l2 {
		t.Error("WithLevel with same level should return same logger")
	}
}

func testCblogInvalidLevel(t core.T, logger slog.Logger, ch <-chan cblog.LogMsg) {
	// cblog sends a panic-level message for invalid levels instead of actually panicking
	// We need to capture the panic message from the channel
	done := make(chan bool)
	go func() {
		msg := <-ch
		if msg.Level == slog.Panic && strings.Contains(msg.Message, "invalid log level") {
			done <- true
		} else {
			done <- false
		}
	}()

	// This will send a panic-level message
	logger.WithLevel(slog.UndefinedLevel)

	select {
	case ok := <-done:
		if !ok {
			t.Error("expected panic-level message for invalid level")
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for panic message")
	}
}

func TestLoggerEnabled(t *testing.T) {
	logger, _ := cblog.New(nil)

	if !logger.Enabled() {
		t.Error("logger should be enabled")
	}

	l, enabled := logger.WithEnabled()
	if l == nil {
		t.Fatal("WithEnabled returned nil logger")
	}
	if !enabled {
		t.Error("WithEnabled should return enabled=true")
	}
}

func TestConcurrency(t *testing.T) {
	t.Run("BasicConcurrency", testCblogBasicConcurrency)
	t.Run("ConcurrentFields", testCblogConcurrentFields)
	t.Run("ConcurrentWithVerification", testCblogConcurrentWithVerification)
}

func testCblogBasicConcurrency(t *testing.T) {
	logger, ch := cblog.New(nil)

	// Drain channel in background
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-ch:
			case <-done:
				return
			}
		}
	}()

	slogtest.RunConcurrentTest(t, logger, slogtest.DefaultConcurrencyTest())
}

func testCblogConcurrentFields(t *testing.T) {
	slogtest.TestConcurrentFields(t, func() slog.Logger {
		logger, ch := cblog.New(nil)
		// Drain channel in background
		go func() {
			var count int
			for range ch {
				count++
			}
		}()
		return logger
	})
}

func testCblogConcurrentWithVerification(t *testing.T) {
	logger, ch := cblog.New(nil)

	const numGoroutines = 10
	const numMessages = 100

	// Collect messages
	done := make(chan bool)
	var messages []cblog.LogMsg
	var mu sync.Mutex

	go func() {
		for msg := range ch {
			mu.Lock()
			messages = append(messages, msg)
			mu.Unlock()
			if len(messages) == numGoroutines*numMessages {
				done <- true
				return
			}
		}
	}()

	// Send messages concurrently
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				logger.Info().
					WithField("goroutine", id).
					WithField("message", j).
					Printf("message %d from goroutine %d", j, id)
			}
		}(i)
	}

	wg.Wait()

	select {
	case <-done:
		// All messages received
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout: only received %d messages", len(messages))
	}

	// Verify we got all messages
	if len(messages) != numGoroutines*numMessages {
		t.Errorf("got %d messages, want %d", len(messages), numGoroutines*numMessages)
	}
}

func TestNewWithCallback(t *testing.T) {
	t.Run("WithHandler", testNewWithCallbackWithHandler)
	t.Run("WithNilHandler", testNewWithCallbackWithNilHandler)
	t.Run("WithZeroSize", testNewWithCallbackWithZeroSize)
}

func testNewWithCallbackWithHandler(t *testing.T) {
	var messages []cblog.LogMsg
	var mu sync.Mutex

	handler := func(msg cblog.LogMsg) {
		mu.Lock()
		messages = append(messages, msg)
		mu.Unlock()
	}

	logger := cblog.NewWithCallback(10, handler)
	if logger == nil {
		t.Fatal("NewWithCallback returned nil")
	}

	// Send some messages
	logger.Info().Print("message 1")
	logger.Debug().WithField("key", "value").Print("message 2")
	logger.Error().Print("message 3")

	// Wait a bit for handler to process
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(messages) != 3 {
		t.Fatalf("got %d messages, want 3", len(messages))
	}

	// Verify messages
	expected := []struct {
		level   slog.LogLevel
		message string
	}{
		{slog.Info, "message 1"},
		{slog.Debug, "message 2"},
		{slog.Error, "message 3"},
	}

	for i, want := range expected {
		if messages[i].Level != want.level {
			t.Errorf("message %d: got level %v, want %v", i, messages[i].Level, want.level)
		}
		if messages[i].Message != want.message {
			t.Errorf("message %d: got message %q, want %q", i, messages[i].Message, want.message)
		}
	}

	// Check fields on second message
	if messages[1].Fields["key"] != "value" {
		t.Errorf("message 1: missing or incorrect field")
	}
}

func testNewWithCallbackWithNilHandler(t *testing.T) {
	logger := cblog.NewWithCallback(10, nil)
	if logger != nil {
		t.Error("NewWithCallback with nil handler should return nil")
	}
}

func testNewWithCallbackWithZeroSize(t *testing.T) {
	var called int32
	handler := func(_ cblog.LogMsg) {
		atomic.StoreInt32(&called, 1)
	}

	logger := cblog.NewWithCallback(0, handler)
	if logger == nil {
		t.Fatal("NewWithCallback returned nil")
	}

	logger.Info().Print("test")
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&called) == 0 {
		t.Error("handler was not called")
	}
}

func TestFieldChaining(t *testing.T) {
	logger, recorder := newCblogWithRecorder()

	// Create a logger with base fields
	baseLogger := logger.Info().
		WithField("app", "test").
		WithField("version", "1.0")

	// Add more fields in derived logger
	derivedLogger := baseLogger.
		WithField("component", "auth").
		WithField("user", "john")

	// Log from derived logger
	derivedLogger.Print("test message")

	// Give callback time to process
	time.Sleep(10 * time.Millisecond)

	msgs := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)
	// Check log level
	slogtest.AssertMessage(t, msgs[0], slog.Info, "test message")

	// Check all fields are present
	slogtest.AssertField(t, msgs[0], "app", "test")
	slogtest.AssertField(t, msgs[0], "version", "1.0")
	slogtest.AssertField(t, msgs[0], "component", "auth")
	slogtest.AssertField(t, msgs[0], "user", "john")
}

func TestComplexFieldTypes(t *testing.T) {
	logger, recorder := newCblogWithRecorder()

	// Test various field types
	type customStruct struct {
		Name  string
		Value int
	}

	fields := map[string]any{
		"string":  "hello",
		"int":     42,
		"int64":   int64(9223372036854775807),
		"float32": float32(3.14),
		"float64": 3.14159265359,
		"bool":    true,
		"nil":     nil,
		"slice":   []int{1, 2, 3},
		"map":     map[string]int{"a": 1, "b": 2},
		"struct":  customStruct{Name: "test", Value: 123},
		"pointer": &customStruct{Name: "ptr", Value: 456},
		"frame":   core.Here(),
	}

	logger.Info().WithFields(fields).Print("complex fields test")

	// Give callback time to process
	time.Sleep(10 * time.Millisecond)

	msgs := recorder.GetMessages()
	slogtest.AssertMustMessageCount(t, msgs, 1)
	slogtest.AssertMessage(t, msgs[0], slog.Info, "complex fields test")

	// Verify all fields are present
	for k := range fields {
		if _, ok := msgs[0].Fields[k]; !ok {
			t.Errorf("missing field %q", k)
		}
	}
}

func BenchmarkLogger(b *testing.B) {
	// Create a logger with a handler that discards messages
	discardHandler := func(_ cblog.LogMsg) {
		// No-op - just discard
	}

	b.Run("SimpleMessage", func(b *testing.B) {
		logger := cblog.NewWithCallback(1000, discardHandler)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().Print("benchmark message")
		}
	})

	b.Run("WithFields", func(b *testing.B) {
		logger := cblog.NewWithCallback(1000, discardHandler)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().
				WithField("key1", "value1").
				WithField("key2", 42).
				WithField("key3", true).
				Print("benchmark message")
		}
	})

	b.Run("WithFieldsMap", func(b *testing.B) {
		logger := cblog.NewWithCallback(1000, discardHandler)
		fields := map[string]any{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info().WithFields(fields).Print("benchmark message")
		}
	})
}
