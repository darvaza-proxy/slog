module github.com/darvaza-proxy/slog/handlers/logrus

go 1.19

replace github.com/darvaza-proxy/slog => ../../

require (
	darvaza.org/core v0.9.0
	github.com/darvaza-proxy/slog v0.4.7
	github.com/sirupsen/logrus v1.9.0
)

require (
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
)
