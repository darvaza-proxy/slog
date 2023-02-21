# slog.Logger adapter for logrus

[![Go Reference](https://pkg.go.dev/badge/github.com/darvaza-proxy/slog/handlers/logrus.svg)](https://pkg.go.dev/github.com/darvaza-proxy/slog/handlers/logrus)

This package implements a wrapper around a `*logrus.Logger` so
it can be used as a `slog.Logger`.

It is important `SetReportCaller()` is disabled otherwise `logrus` will
set a useless `"method"` field pointing to our `Print()` handler.
`WithStack()` will set the `"method"` field considering the provided `skip` value

`WithStack()` will also create a `"call-stack"` field with the complete
call stack from the caller upward.

## See also

* [github.com/darvaza-proxy/slog](https://pkg.go.dev/github.com/darvaza-proxy/slog)
* [github.com/sirupsen/logrus](https://pkg.go.dev/github.com/sirupsen/logrus)
