package slog

type Logger interface {
	Debug(string, ...any)
	Info(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
	Fatal(string, ...any)

	Printf(string, ...any)
	WithField(string, any) Logger
	WithFields(map[string]any) Logger
}
