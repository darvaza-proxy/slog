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
)

func TestLevel(t *testing.T) {
	// Test nil receiver
	var nilLogger *cblog.Logger
	core.AssertEqual(t, slog.UndefinedLevel, nilLogger.Level(), "nil logger level")

	// Test normal logger
	logger, _ := cblog.New(nil)
	cblogLogger := core.AssertMustTypeIs[*cblog.Logger](t, logger, "logger type")
	core.AssertEqual(t, slog.UndefinedLevel, cblogLogger.Level(), "default level")

	// Test level-specific logger
	infoLogger := core.AssertMustTypeIs[*cblog.Logger](t, logger.Info(), "info logger type")
	core.AssertEqual(t, slog.Info, infoLogger.Level(), "info level")
}

// Compile-time verification that test case types implement TestCase interface
var _ core.TestCase = messageVerificationTestCase{}
var _ core.TestCase = callbackMessageTestCase{}

type finaliserResult struct {
	closed bool
	count  int
}

func TestFinaliserClosesInternalChannel(t *testing.T) {
	// Test that finaliser closes internally created channels
	resultChan := make(chan finaliserResult, 1)
	runInternalFinaliserScenario(resultChan)

	// Force garbage collection so the finaliser fires deterministically.
	//revive:disable-next-line:call-to-gc
	runtime.GC()
	runtime.Gosched()

	select {
	case res := <-resultChan:
		verifyFinaliserClosed(t, res, 3)
	case <-time.After(2 * time.Second):
		t.Error("finaliser did not close internal channel")
	}
}

// runInternalFinaliserScenario creates a logger backed by an internal
// channel, drains the channel until it is closed, sends a few messages,
// then drops the logger so the finaliser can run.
func runInternalFinaliserScenario(resultChan chan<- finaliserResult) {
	logger, logCh := cblog.New(nil)

	go func() {
		count := 0
		for range logCh {
			count++
		}
		resultChan <- finaliserResult{closed: true, count: count}
	}()

	logger.Info().Print("test message 1")
	logger.Debug().Print("test message 2")
	logger.Warn().Print("test message 3")
}

func verifyFinaliserClosed(t *testing.T, res finaliserResult, want int) {
	t.Helper()
	if !res.closed {
		t.Error("channel was not closed")
	}
	if res.count != want {
		t.Errorf("expected %d messages, got %d", want, res.count)
	}
}

func TestFinaliserDoesNotCloseExternalChannel(t *testing.T) {
	// Test that finaliser does NOT close externally provided channels
	ch := make(chan cblog.LogMsg, 10)
	resultChan := make(chan finaliserResult, 1)

	go monitorExternalChannelUntilManual(ch, resultChan)
	runExternalFinaliserScenario(ch)

	// Force garbage collection so the finaliser fires deterministically.
	//revive:disable-next-line:call-to-gc
	runtime.GC()
	runtime.Gosched()

	// Give time for any potential finaliser to run
	time.Sleep(500 * time.Millisecond)

	if !sendManualMessage(ch) {
		t.Error("channel appears to be closed")
	}

	select {
	case res := <-resultChan:
		verifyFinaliserNotClosed(t, res, 3)
	case <-time.After(time.Second):
		t.Error("timeout waiting for result")
	}

	close(ch) // Clean up
}

// monitorExternalChannelUntilManual counts messages on ch and reports
// once the manual sentinel is received, or when the channel is closed.
func monitorExternalChannelUntilManual(ch <-chan cblog.LogMsg, result chan<- finaliserResult) {
	count := 0
	for msg := range ch {
		count++
		if count == 3 && msg.Message == "manual message" {
			result <- finaliserResult{closed: false, count: count}
			return
		}
	}
	result <- finaliserResult{closed: true, count: count}
}

// runExternalFinaliserScenario creates a logger backed by ch, sends a
// couple of messages, then drops the logger so the finaliser can run.
func runExternalFinaliserScenario(ch chan cblog.LogMsg) {
	logger, _ := cblog.New(ch)
	logger.Info().Print("test message 1")
	logger.Debug().Print("test message 2")
}

func sendManualMessage(ch chan<- cblog.LogMsg) bool {
	select {
	case ch <- cblog.LogMsg{Level: slog.Info, Message: "manual message"}:
		return true
	default:
		return false
	}
}

func verifyFinaliserNotClosed(t *testing.T, res finaliserResult, want int) {
	t.Helper()
	if res.closed {
		t.Error("finaliser incorrectly closed external channel")
	}
	if res.count != want {
		t.Errorf("expected %d messages, got %d", want, res.count)
	}
}

func TestNew(t *testing.T) {
	t.Run("WithNilChannel", testNewWithNilChannel)
	t.Run("WithBufferedChannel", testNewWithBufferedChannel)
}

func testNewWithNilChannel(t *testing.T) {
	logger, ch := cblog.New(nil)
	if !core.AssertNotNil(t, logger, "New returned nil logger") {
		return
	}
	if !core.AssertNotNil(t, ch, "New returned nil channel") {
		return
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
	if !core.AssertNotNil(t, logger, "New returned nil logger") {
		return
	}
	if !core.AssertEqual(t, ch, outCh, "returned channel") {
		return
	}

	// Send multiple messages
	logger.Debug().Print("debug")
	logger.Info().Print("info")
	logger.Warn().Print("warn")

	// Verify messages using TestCase pattern
	core.RunTestCases(t, messageVerificationTestCases(ch))
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
	if !core.AssertNotNil(t, l, "WithLevel returned nil") {
		return
	}
	l.Print("error message")

	select {
	case msg := <-ch:
		core.AssertEqual(t, slog.Error, msg.Level, "message level")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func testCblogSameLevel(t core.T, logger slog.Logger) {
	l1 := logger.Info()
	l2 := l1.WithLevel(slog.Info)
	core.AssertSame(t, l1, l2, "WithLevel same level")
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
		core.AssertTrue(t, ok, "expected panic-level message for invalid level")
	case <-time.After(time.Second):
		if !core.AssertTrue(t, false, "timeout waiting for panic message") {
			return
		}
	}
}

func TestLoggerEnabled(t *testing.T) {
	logger, _ := cblog.New(nil)

	core.AssertTrue(t, logger.Enabled(), "logger enabled")

	l, enabled := logger.WithEnabled()
	if !core.AssertNotNil(t, l, "WithEnabled returned nil logger") {
		return
	}
	core.AssertTrue(t, enabled, "WithEnabled result")
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
	const numGoroutines = 10
	const numMessages = 100
	const total = numGoroutines * numMessages

	logger, ch := cblog.New(nil)
	done := make(chan bool, 1)
	collector := &cblogMessageCollector{}

	go collector.collectUntil(ch, total, done)
	sendConcurrentLogs(logger, numGoroutines, numMessages)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		if !core.AssertTrue(t, false, "timeout: only received %d messages", collector.count()) {
			return
		}
	}

	core.AssertEqual(t, total, collector.count(), "message count")
}

type cblogMessageCollector struct {
	messages []cblog.LogMsg
	mu       sync.Mutex
}

// collectUntil appends messages from ch until total are gathered, then
// reports done. It exits when ch closes.
func (c *cblogMessageCollector) collectUntil(ch <-chan cblog.LogMsg, total int, done chan<- bool) {
	for msg := range ch {
		c.mu.Lock()
		c.messages = append(c.messages, msg)
		finished := len(c.messages) == total
		c.mu.Unlock()
		if finished {
			done <- true
			return
		}
	}
}

func (c *cblogMessageCollector) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.messages)
}

// sendConcurrentLogs fans out numGoroutines workers, each sending
// numMessages messages tagged with their goroutine and message index.
func sendConcurrentLogs(logger slog.Logger, numGoroutines, numMessages int) {
	var wg sync.WaitGroup
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range numMessages {
				logger.Info().
					WithField("goroutine", id).
					WithField("message", j).
					Printf("message %d from goroutine %d", j, id)
			}
		}(i)
	}
	wg.Wait()
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
	if !core.AssertNotNil(t, logger, "NewWithCallback returned nil") {
		return
	}

	// Send some messages
	logger.Info().Print("message 1")
	logger.Debug().WithField("key", "value").Print("message 2")
	logger.Error().Print("message 3")

	// Wait a bit for handler to process
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if !core.AssertEqual(t, 3, len(messages), "message count") {
		return
	}

	// Verify messages using TestCase pattern
	core.RunTestCases(t, callbackMessageTestCases(messages))

	// Check fields on second message
	core.AssertEqual(t, "value", messages[1].Fields["key"], "message 1 field")
}

func testNewWithCallbackWithNilHandler(t *testing.T) {
	logger := cblog.NewWithCallback(10, nil)
	core.AssertNil(t, logger, "NewWithCallback nil handler")
}

func testNewWithCallbackWithZeroSize(t *testing.T) {
	var called atomic.Int32
	handler := func(_ cblog.LogMsg) {
		called.Store(1)
	}

	logger := cblog.NewWithCallback(0, handler)
	if !core.AssertNotNil(t, logger, "NewWithCallback returned nil") {
		return
	}

	logger.Info().Print("test")
	time.Sleep(100 * time.Millisecond)

	core.AssertNotEqual(t, int32(0), called.Load(), "handler was not called")
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
		_, ok := msgs[0].Fields[k]
		core.AssertTrue(t, ok, "missing field %q", k)
	}
}

type messageVerificationTestCase struct {
	channel <-chan cblog.LogMsg
	name    string
	message string
	level   slog.LogLevel
}

func (tc messageVerificationTestCase) Name() string {
	return tc.name
}

func (tc messageVerificationTestCase) Test(t *testing.T) {
	t.Helper()
	select {
	case got := <-tc.channel:
		core.AssertEqual(t, tc.level, got.Level, "message level")
		core.AssertEqual(t, tc.message, got.Message, "message text")
	case <-time.After(time.Second):
		core.AssertTrue(t, false, "timeout waiting for message")
	}
}

func newMessageVerificationTestCase(
	name string, level slog.LogLevel, message string, channel <-chan cblog.LogMsg,
) messageVerificationTestCase {
	return messageVerificationTestCase{
		name:    name,
		level:   level,
		message: message,
		channel: channel,
	}
}

func messageVerificationTestCases(channel <-chan cblog.LogMsg) []messageVerificationTestCase {
	return []messageVerificationTestCase{
		newMessageVerificationTestCase("Debug", slog.Debug, "debug", channel),
		newMessageVerificationTestCase("Info", slog.Info, "info", channel),
		newMessageVerificationTestCase("Warn", slog.Warn, "warn", channel),
	}
}

type callbackMessageTestCase struct {
	name     string
	message  string
	messages []cblog.LogMsg
	index    int
	level    slog.LogLevel
}

func (tc callbackMessageTestCase) Name() string {
	return tc.name
}

func (tc callbackMessageTestCase) Test(t *testing.T) {
	t.Helper()
	if tc.index >= len(tc.messages) {
		core.AssertTrue(t, false, "message index %d out of bounds", tc.index)
		return
	}
	core.AssertEqual(t, tc.level, tc.messages[tc.index].Level, "message %d level", tc.index)
	core.AssertEqual(t, tc.message, tc.messages[tc.index].Message, "message %d text", tc.index)
}

func newCallbackMessageTestCase(
	name string, index int, level slog.LogLevel, message string, messages []cblog.LogMsg,
) callbackMessageTestCase {
	return callbackMessageTestCase{
		name:     name,
		index:    index,
		level:    level,
		message:  message,
		messages: messages,
	}
}

func callbackMessageTestCases(messages []cblog.LogMsg) []callbackMessageTestCase {
	return []callbackMessageTestCase{
		newCallbackMessageTestCase("Message1", 0, slog.Info, "message 1", messages),
		newCallbackMessageTestCase("Message2", 1, slog.Debug, "message 2", messages),
		newCallbackMessageTestCase("Message3", 2, slog.Error, "message 3", messages),
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
