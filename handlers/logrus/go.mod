module darvaza.org/slog/handlers/logrus

go 1.23.0

replace darvaza.org/slog => ../../

require (
	darvaza.org/core v0.18.3
	darvaza.org/slog v0.8.1
)

require github.com/sirupsen/logrus v1.9.4

require (
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)
