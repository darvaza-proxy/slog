module github.com/darvaza-proxy/slog/handlers/logrus

go 1.19

replace github.com/darvaza-proxy/slog => ../../

require (
	github.com/darvaza-proxy/slog v0.4.2
	github.com/sirupsen/logrus v1.9.0
)

require golang.org/x/sys v0.5.0 // indirect
