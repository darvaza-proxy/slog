# slog, a back-end agnostic interface for structured logs

[![Go Reference](https://pkg.go.dev/badge/darvaza.org/slog.svg)](https://pkg.go.dev/darvaza.org/slog)
[![Codebeat Badge](https://codebeat.co/badges/f96b7fad-4653-4421-9cb9-c2fe7a4678da)](https://codebeat.co/projects/github-com-darvaza-proxy-slog-main)

[`slog.Logger`](#interface) provides a simple standardised interface for structured logs for libraries. It supports [six log levels](#log-levels) and [fields](#fields) with unique string labels (keys).

## Interface
Every method of this interface, with the exception of [`Print()`](#print), returns a `slog.Logger` so it can be daisy chained when composing a log entry.
A log entry begins with setting the [level](#log-levels) followed by optional addition of [fields](#fields) and a [call stack](#call-stack) and ends with a message calling a [Print](#print) method.
Based on the specified [level](#log-levels) an entry can be [enabled or disabled](#enabled). Calls to methods on disabled entries will cause no action unless it's used to create a new entry with a [level](#log-levels) that is enabled.

## Log Levels
An `slog.Logger` entry can have of one of six levels, of which Fatal is expected to end the execution just like the standard `log.Fatal()`right after adding the log entry, and Panic to raise a recoverable panic like `log.Panic()`.

 1. Debug
 2. Info
 3. Warn
 4. Error
 5. Fatal
 6. Panic

New log entries can be created by calling the named shortcut methods (`Debug()`, `Info()`, `Warn()`, `Error()`, `Fatal()`, and `Panic()`) or via `WithLevel(level)`.


## Enabled
A log entry is considered _Enabled_ if the handler would actually log entries of the specified level.
It is always safe to operate on disabled loggers and the cost of should be negletable as when a logger
is not `Enabled()` string formatting operations or fields and stack commands are not performed.

Sometimes it is useful to know if a certain level is *Enabled* so you can decide between two levels with different degree
of detail. For this purpose one can use `WithEnabled()` like this:
```go
if log, ok := logger.Debug().WithEnabled(); ok {
	log.WithField("request", req).Print("Let's write detailed debug stuff")
} else if log, ok := logger.Info().WithEnabled(); ok {
	log.Print("Let's write info stuff instead")
}
```

Logs of Fatal and Panic level are expected to exit/panic regardless of the _Enabled_ state.

## Fields
In `slog` fields are unique key/value pairs where the key is a non-empty string and the value could be any type.

## Call Stack
A Call stack is attached to a log entry considering the given distance to a caller/initiator function.

## Print
`slog.Logger` support three Print methods mimicking their equivalent in the `fmt` package from the standard library. `Print()`, `Println()`, and `Printf()` that finally attempt to emit the log entry with the given message and any previously attached [Field](#fields).

## Standard *log.Logger
In order to be compatible with the standard library's provided `log.Logger`, `slog` provides an `io.Writer` interface connected to a handler function that is expected to parse the entry and call a provided `slog.Logger` as appropriate. This _writer_ is created by calling `NewLogWriter` and passing the logger and the handler function, which is then passed to `log.New()` to create the `*log.Logger`.

Alternatively a generic handler is provided when using `NewStdLogger()`.

## Handlers

A handler is an object that implements the `slog.Logger` interface.
We provide handlers to use popular loggers as _backend_.

* [logrus](https://pkg.go.dev/darvaza.org/slog/handlers/logrus)
* [zap](https://pkg.go.dev/darvaza.org/slog/handlers/zap)
* [zerolog](https://pkg.go.dev/darvaza.org/slog/handlers/zerolog)

We also offer backend independent handlers

* [cblog](https://pkg.go.dev/darvaza.org/slog/handlers/cblog), a implementation
that allows you to receive log entries through a channel.
* [filter](https://pkg.go.dev/darvaza.org/slog/handlers/filter), that can filter by level and also alter log entries before passing them to another slog.Logger.
* [discard](https://pkg.go.dev/darvaza.org/slog/handlers/discard), a placeholder that won't log anything but saves the user from checking if a logger was provided or not every time.
