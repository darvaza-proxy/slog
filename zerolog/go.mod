module github.com/darvaza-proxy/slog/zerolog

go 1.19

replace github.com/darvaza-proxy/slog => ../

require (
	github.com/darvaza-proxy/slog v0.0.0-20230102194403-372027bb9066
	github.com/rs/zerolog v1.28.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	golang.org/x/sys v0.3.0 // indirect
)
