# non-Logger for slog.Logger

[![Go Reference](https://pkg.go.dev/badge/darvaza.org/slog/handlers/discard.svg)](https://pkg.go.dev/darvaza.org/slog/handlers/discard)

The `discard` handler is a placeholder to avoid having to conditionally decide if using a logger
or not. `discard` will handle Panic() and Fatal() correctly, but everything else will be discarded.

for Panic/Fatal messages, the [Go standard logger](https://pkg.go.dev/log#Output) will be called. fields and call stack are lost.

## See also

* [darvaza.org/slog](https://pkg.go.dev/darvaza.org/slog)
