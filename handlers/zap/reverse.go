package zap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/slog"
)

var (
	_ zapcore.WriteSyncer  = (*Reverse)(nil)
	_ zapcore.LevelEnabler = (*Reverse)(nil)
)

// Reverse implements a json decoder for zapcore
type Reverse struct {
	l slog.Logger
}

func (*Reverse) Write(p []byte) (int, error) {
	n := len(p)
	_, err := fmt.Fprintf(os.Stderr, "%s\n%s\n", "----", hex.Dump(p))
	return n, err
}

// Sync is expected to flush buffered log data into
// persistent storage.
func (*Reverse) Sync() error {
	return nil
}

// Enabled returns true if the given level is at or above this level.
func (r *Reverse) Enabled(zl zapcore.Level) bool {
	level, ok := fromZapLevel(zl)
	if !ok {
		return false
	}

	return r.l.WithLevel(level).Enabled()
}

// NewReversed returns a [zap.Logger] built on top of a [slog.Logger].
func NewReversed(parent slog.Logger) (*zap.Logger, error) {
	switch l := parent.(type) {
	case nil:
		// parent-less unacceptable
		return nil, errors.New("invalid parent logger")
	case *Logger:
		// unwrap
		return l.logger, nil
	default:
		// reverse wrapper
		r := &Reverse{l: parent}
		enc := zapcore.NewJSONEncoder(newDefaultEncoderConfig())
		log := zap.New(zapcore.NewCore(enc, r, r))

		return log, nil
	}
}
