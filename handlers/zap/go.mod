module darvaza.org/slog/handlers/zap

go 1.23.0

replace darvaza.org/slog => ../../

require (
	darvaza.org/core v0.18.3
	darvaza.org/slog v0.8.1
)

require go.uber.org/zap v1.27.1

require (
	github.com/stretchr/testify v1.9.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)
