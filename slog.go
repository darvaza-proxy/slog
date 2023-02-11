// Package slog provides a backend agnostic interface for structured logs
package slog

// LogLevel represents the level of criticality of a log entry
type LogLevel int8

const (
	// UndefinedLevel is a placeholder for the zero-value when no level has been set
	UndefinedLevel LogLevel = iota

	Panic // Panic represents a log entry for a fatal problem that could be stoped by defer/recover
	Fatal // Fatal represents a log entry for a problem we can't recover
	Error // Error represents a log entry for a problem we can recover
	Warn  // Warn represents a log entry for something that might not a problem but it's worth mentioning
	Info  // Info represents a log entry just to tell what we are doing
	Debug // Debug represents a log entry that contains information important mostly only to developers

	// ErrorFieldName is the preferred field label for errors
	ErrorFieldName = "error"
)

// Logger is a backend agnostic interface for structured logs
type Logger interface {
	Debug() Logger // Debug is an alias of WithLevel(Debug)
	Info() Logger  // Info is an alias of WithLevel(Info)
	Warn() Logger  // Warn is an alias of WithLevel(Warn)
	Error() Logger // Error is an alias of WithLevel(Error)
	Fatal() Logger // Fatal is an alias of WithLevel(Fatal)
	Panic() Logger // Panic is an alias of WithLevel(Panic)

	Print(...any)          // Print adds a log entry handled in the manner of fmt.Print
	Println(...any)        // Println adds a log entry handled in the manner of fmt.Println
	Printf(string, ...any) // Printf adds a log entry handled in the manner of fmt.Printf

	WithLevel(LogLevel) Logger        // WithLevel returns a new log context set to add entries to the specified level
	WithStack(int) Logger             // WithStack attaches a call stack a log context
	WithField(string, any) Logger     // WithField attaches a field to a log context
	WithFields(map[string]any) Logger // WithFields attaches a set of fields to a log context

	// Enabled tells if the Logger would actually log
	Enabled() bool

	// WithEnabled tells if Enabled but also passes a reference to
	// the logger for convenience when choosing what to log
	//
	// e.g.
	// if log, ok := logger.Debug().WithEnabled(); ok {
	//    log.Print("Let's write detailed debug stuff")
	// } elseif log, ok := logger.Info().WithEnabled(); ok {
	//    log.Print("Let's write info stuff instead")
	// }
	WithEnabled() (Logger, bool)
}

// Fields is sugar syntax for WithFields() for those
// who believe log.WithFields(slog.Fields{foo: bar}) is
// nicer than log.WithFields(map[string]any{foo: var})
type Fields map[string]any
