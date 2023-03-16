module github.com/darvaza-proxy/slog/handlers/zap

go 1.19

replace github.com/darvaza-proxy/slog => ../../

require (
	github.com/darvaza-proxy/slog v0.4.6
	go.uber.org/zap v1.24.0
)

require (
	github.com/darvaza-proxy/core v0.6.1 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)
