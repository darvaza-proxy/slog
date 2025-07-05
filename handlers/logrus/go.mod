module darvaza.org/slog/handlers/logrus

go 1.23.0

replace darvaza.org/slog => ../../

require (
	darvaza.org/core v0.17.1
	darvaza.org/slog v0.7.0
)

require github.com/sirupsen/logrus v1.9.3

require (
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)
