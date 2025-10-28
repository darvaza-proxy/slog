# AI Agent Development Guide for slog/handlers/logr

This file provides guidance to AI agents when working with the logr handler.
For general information about the handler, see [README.md](README.md).

## Handler Overview

The logr handler provides bidirectional adapters between `slog.Logger` and
`go-logr/logr`. It consists of two main components:

1. **Logger**: Uses a logr.Logger as the backend for slog.Logger
2. **Sink**: Implements logr.LogSink using a slog.Logger as the backend

## Key Implementation Details

### Logger (logr.go)

- Embeds `internal.Loglet` for field chain management
- Maps slog levels to logr V-levels and Error() calls
- Handles Fatal/Panic by calling Error() then panicking
- Preserves all structured fields through the Loglet chain

### Sink (sink.go)

- Implements all required logr.LogSink methods
- Implements optional CallDepthLogSink and CallStackHelperLogSink interfaces
- Uses `core.SortedKeys` for consistent field ordering
- Accumulates logger names in a "logger" field

## Development Guidelines

### Adding Features

1. Maintain bidirectional compatibility - features should work in both
   directions
2. Preserve the immutable logger pattern
3. Ensure structured fields are properly propagated
4. Test both adapters when making changes

### Testing

The handler includes comprehensive tests:

- `logr_test.go`: Main adapter tests
- `logr_loglet_test.go`: Tests for Loglet integration

Run tests with: `make test-logr`

### Common Tasks

1. **Updating level mappings**: Modify `mapToLogrLevel()` and
   `mapFromLogrLevel()`
2. **Adding fields**: Ensure both adapters handle new field types
3. **Performance improvements**: Focus on the field collection in `print()` and
   `addKeysAndValues()`

## Important Considerations

1. **Level Semantics**: logr uses "verbosity" levels (higher = less important),
   while slog uses severity levels (higher = more important)
2. **Error Handling**: logr's Error() method takes an error parameter, while
   slog embeds errors as fields
3. **No Format Strings**: Both interfaces avoid format strings in favour of
   structured fields
4. **Performance**: The adapters add minimal overhead, mostly in field
   collection

## See Also

- [slog AGENTS.md](../../AGENTS.md) - General slog development guide
- [logr documentation](https://github.com/go-logr/logr) - Understanding logr
  concepts
