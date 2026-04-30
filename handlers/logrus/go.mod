module darvaza.org/slog/handlers/logrus

go 1.24.0

replace darvaza.org/slog => ../../

require (
	darvaza.org/core v0.19.0
	darvaza.org/slog v0.8.1
)

require github.com/sirupsen/logrus v1.9.3

require (
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
