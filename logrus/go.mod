module github.com/darvaza-proxy/slog/logrus

go 1.19

replace github.com/darvaza-proxy/slog => ../

require (
	github.com/darvaza-proxy/slog v0.0.0-20230102194403-372027bb9066
	github.com/sirupsen/logrus v1.9.0
)

require golang.org/x/sys v0.3.0 // indirect
