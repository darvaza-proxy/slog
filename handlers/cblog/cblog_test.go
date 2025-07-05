package cblog_test

import (
	"strings"
	"sync"
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

func TestNew(t *testing.T) {
	t.Run("WithNilChannel", func(t *testing.T) {
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
	})

	t.Run("WithBufferedChannel", func(t *testing.T) {
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
	})
}

func TestLoggerLevels(t *testing.T) {
	logger, ch := cblog.New(nil)

	testCases := []struct {
		name   string
		method func() slog.Logger
		level  slog.LogLevel
	}{
		{"Debug", logger.Debug, slog.Debug},
		{"Info", logger.Info, slog.Info},
		{"Warn", logger.Warn, slog.Warn},
		{"Error", logger.Error, slog.Error},
		{"Fatal", logger.Fatal, slog.Fatal},
		{"Panic", logger.Panic, slog.Panic},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := tc.method()
			if l == nil {
				t.Fatal("level method returned nil")
			}

			l.Print("test")

			select {
			case msg := <-ch:
				if msg.Level != tc.level {
					t.Errorf("got level %v, want %v", msg.Level, tc.level)
				}
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for message")
			}
		})
	}
}

func TestLoggerPrintMethods(t *testing.T) {
	logger, ch := cblog.New(nil)

	t.Run("Print", func(t *testing.T) {
		logger.Info().Print("hello", " ", "world")
		select {
		case msg := <-ch:
			if msg.Message != testHelloWorld {
				t.Errorf("got message %q, want %q", msg.Message, testHelloWorld)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}
	})

	t.Run("Println", func(t *testing.T) {
		logger.Info().Println("hello", "world")
		select {
		case msg := <-ch:
			if msg.Message != testHelloWorld {
				t.Errorf("got message %q, want %q", msg.Message, testHelloWorld)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}
	})

	t.Run("Printf", func(t *testing.T) {
		logger.Info().Printf("hello %s", "world")
		select {
		case msg := <-ch:
			if msg.Message != testHelloWorld {
				t.Errorf("got message %q, want %q", msg.Message, testHelloWorld)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}
	})
}

func TestLoggerWithField(t *testing.T) {
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

	slogtest.TestWithField(t, logger)
}

func TestLoggerWithFields(t *testing.T) {
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

	slogtest.TestWithFields(t, logger)
}

func TestLoggerWithStack(t *testing.T) {
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

	slogtest.TestWithStack(t, logger)
}

func TestLoggerWithLevel(t *testing.T) {
	logger, ch := cblog.New(nil)

	t.Run("ValidLevel", func(t *testing.T) {
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
	})

	t.Run("SameLevel", func(t *testing.T) {
		l1 := logger.Info()
		l2 := l1.WithLevel(slog.Info)
		if l1 != l2 {
			t.Error("WithLevel with same level should return same logger")
		}
	})

	t.Run("InvalidLevel", func(t *testing.T) {
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
	})
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
	t.Run("BasicConcurrency", func(t *testing.T) {
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
	})

	t.Run("ConcurrentFields", func(t *testing.T) {
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
	})

	t.Run("ConcurrentWithVerification", func(t *testing.T) {
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
	})
}

func TestNewWithCallback(t *testing.T) {
	t.Run("WithHandler", func(t *testing.T) {
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
	})

	t.Run("WithNilHandler", func(t *testing.T) {
		logger := cblog.NewWithCallback(10, nil)
		if logger != nil {
			t.Error("NewWithCallback with nil handler should return nil")
		}
	})

	t.Run("WithZeroSize", func(t *testing.T) {
		var called bool
		handler := func(_ cblog.LogMsg) {
			called = true
		}

		logger := cblog.NewWithCallback(0, handler)
		if logger == nil {
			t.Fatal("NewWithCallback returned nil")
		}

		logger.Info().Print("test")
		time.Sleep(100 * time.Millisecond)

		if !called {
			t.Error("handler was not called")
		}
	})
}

func TestFieldChaining(t *testing.T) {
	logger, ch := cblog.New(nil)

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

	select {
	case msg := <-ch:
		if len(msg.Fields) != 4 {
			t.Errorf("got %d fields, want 4", len(msg.Fields))
		}

		expectedFields := map[string]any{
			"app":       "test",
			"version":   "1.0",
			"component": "auth",
			"user":      "john",
		}

		for k, v := range expectedFields {
			if msg.Fields[k] != v {
				t.Errorf("field %s: got %v, want %v", k, msg.Fields[k], v)
			}
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestComplexFieldTypes(t *testing.T) {
	logger, ch := cblog.New(nil)

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

	select {
	case msg := <-ch:
		if len(msg.Fields) != len(fields) {
			t.Errorf("got %d fields, want %d", len(msg.Fields), len(fields))
		}

		// Verify all fields are present
		for k := range fields {
			if _, ok := msg.Fields[k]; !ok {
				t.Errorf("missing field %q", k)
			}
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
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
