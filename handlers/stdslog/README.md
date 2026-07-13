# darvaza.org/slog/handlers/stdslog

`darvaza.org/slog/handlers/stdslog` provides bidirectional adapters
between `slog.Logger` and the standard library [log/slog][stdslog-pkg]
package.

This package enables:

- Using any standard library `slog.Handler` or `*slog.Logger` as a
  slog.Logger backend
- Using any slog.Logger as a standard library `slog.Handler` backend,
  ready for injection into libraries that take a `*slog.Logger`

Handlers are split into separate Go modules primarily to avoid pulling
additional dependencies; this package needs nothing beyond what the
main `darvaza.org/slog` module already requires (the standard library
and `darvaza.org/core`), so it lives there.

## Installation

```bash
go get darvaza.org/slog
```

## Usage

### Using log/slog as slog.Logger

```go
package main

import (
    stdslog "log/slog"
    "os"

    slogstdslog "darvaza.org/slog/handlers/stdslog"
)

func main() {
    // Wrap a standard library handler...
    h := stdslog.NewJSONHandler(os.Stdout, nil)
    logger := slogstdslog.NewWithHandler(h)

    // ...or a *slog.Logger. New(nil) snapshots slog.Default()
    // at construction.
    logger = slogstdslog.New(stdslog.Default())

    // Use as normal slog.Logger
    logger.Info().
        WithField("component", "main").
        Print("Application started")
}
```

### Using slog.Logger as log/slog

```go
package main

import (
    "darvaza.org/slog"
    slogstdslog "darvaza.org/slog/handlers/stdslog"
)

func run(backend slog.Logger) {
    // *slog.Logger backed by any slog.Logger
    logger := slogstdslog.NewSLogger(backend)

    // Use as normal standard library logger
    logger.WithGroup("req").Info("hello", "id", 42)
    // recorded by the backend as field "req.id" = 42

    // NewHandler returns the slog.Handler itself for
    // handler-level injection.
}
```

## Level Mapping

### slog to log/slog (when using log/slog as backend)

| slog Level | log/slog Level             |
|------------|----------------------------|
| Debug      | LevelDebug                 |
| Info       | LevelInfo                  |
| Warn       | LevelWarn                  |
| Error      | LevelError                 |
| Fatal      | LevelError+4 + os.Exit(1)  |
| Panic      | LevelError+8 + panic       |

Fatal and Panic have no standard library equivalent; they map above
LevelError, preserving slog's severity ordering, and the terminal
behaviour happens at this adapter after the record is delivered.

### log/slog to slog (when using slog as backend)

| log/slog Level                | slog Level |
|-------------------------------|------------|
| below LevelInfo               | Debug      |
| LevelInfo to below LevelWarn  | Info       |
| LevelWarn to below LevelError | Warn       |
| LevelError and above          | Error      |

Inbound levels cap at Error, so records above LevelError never trigger
the backend's Fatal or Panic terminal behaviour.

## Behaviour Notes

- Groups (`WithGroup`, `slog.Group`) become dot-separated key prefixes
  (`req.id`), as slog fields are flat. Inline (empty-name) groups
  expand unprefixed; attributes with an empty key are dropped.
- slog fields form a map, so duplicate keys collapse last-wins on the
  way into a slog.Logger backend; the standard library itself
  preserves duplicate attributes.
- Values normalise through the standard library `slog.Value` on the
  way in: `int` arrives as `int64`, `slog.LogValuer` values are
  resolved.
- The record's PC is discarded and emitted records carry none: slog
  has no caller-attribution concept. `WithStack` attaches the call
  stack as a `"stack"` field instead.
- `Enabled()` on a logger with no level set reports disabled; set a
  level first. Fatal and Panic always report enabled: terminal
  entries are delivered — bypassing the backend's filter — and then
  exit or panic. Level filtering is otherwise the backend's decision
  on both legs.
- Once a level is set, field and stack collection is bound to
  `Enabled()`: a disabled leveled entry skips `WithField`/`WithStack`
  rather than buffering attachments it will never emit, matching the
  `filter` handler. Entries with no level set collect speculatively
  until a level is chosen.
- `Handler.WithAttrs` applies attributes to the backend logger
  eagerly, so attached attributes reach every subsequent record.

## Features

- **Bidirectional Conversion**: Convert between slog and log/slog in
  both directions
- **Field Preservation**: Structured fields are preserved across
  conversions, in sorted key order on the way out
- **Immutable Loggers**: Both adapters follow the immutable logger
  pattern
- **Shared Helpers**: The level mapping and attribute flattening
  helpers (`MapFromSLogLevel`, `MapToSLogLevel`, `SLogRecordAttrs`,
  `AppendSLogAttr`) are exported for reuse by other bidirectional
  adapters

## Implementation Details

The package provides two main types:

1. **Logger**: Adapts a standard library `slog.Handler` to implement
   slog.Logger
   - Uses `internal.Loglet` for field chain management
   - Delivers entries as `slog.Record`s to the wrapped handler

2. **Handler**: Implements the standard library `slog.Handler`
   interface using a slog.Logger
   - Carries only the backend logger and the open group prefix;
     attributes live in the backend's field chain

## See Also

- [darvaza.org/slog][slog-pkg] - Main slog interface
- [log/slog][stdslog-pkg] - The standard library structured logging
  package

[slog-pkg]: https://pkg.go.dev/darvaza.org/slog
[stdslog-pkg]: https://pkg.go.dev/log/slog
