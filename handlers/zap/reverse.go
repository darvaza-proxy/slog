package zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"darvaza.org/core"
	"darvaza.org/slog"
)

var (
	_ zapcore.LevelEnabler = (*reverseLevels)(nil)
)

// errInvalidParent rejects parents [NewReversed] can't build on.
// It matches [core.ErrInvalid] for errors.Is tests.
var errInvalidParent = core.QuietWrap(core.ErrInvalid,
	"invalid parent logger")

// reverseLevels delegates zap level decisions to a parent slog.Logger.
type reverseLevels struct {
	l slog.Logger
}

// Enabled returns true if the parent logger emits entries at the
// given level.
func (r *reverseLevels) Enabled(zl zapcore.Level) bool {
	level, ok := fromZapLevel(zl)
	if !ok {
		return false
	}

	return r.l.WithLevel(level).Enabled()
}

// NewReversed returns a [zap.Logger] backed by the given [slog.Logger].
// Entries flow to the parent through [SlogCore], and level decisions
// are delegated to the parent. When the parent is itself zap-backed,
// the underlying [zap.Logger] is returned directly, carrying over any
// fields accumulated on the slog side. Optional [zap.Option] values
// are applied to the returned logger either way.
//
// NewReversed fails with an error matching [core.ErrInvalid] when
// parent is nil or not backed by a usable logger.
//
// Note that wrapped DPanic entries are logged at [slog.Error]; zap's
// development-mode panic still applies at the [zap.Logger] layer.
func NewReversed(parent slog.Logger, opts ...zap.Option) (*zap.Logger, error) {
	switch l := parent.(type) {
	case *Logger:
		// unwrap
		return l.unwrapReversed(opts)
	default:
		if core.IsNil(parent) {
			// parent-less unacceptable
			return nil, errInvalidParent
		}
		// reverse wrapper
		zc := NewCore(parent, &reverseLevels{l: parent})
		return zap.New(zc, opts...), nil
	}
}

// unwrapReversed returns the wrapped [zap.Logger], applying options
// and carrying over fields accumulated on the slog side. Pending
// call-stack context is not carried over.
func (zpl *Logger) unwrapReversed(opts []zap.Option) (*zap.Logger, error) {
	if zpl == nil || zpl.logger == nil {
		return nil, errInvalidParent
	}

	zl := zpl.logger
	if len(opts) > 0 {
		zl = zl.WithOptions(opts...)
	}
	if fields := zapFields(zpl.loglet.FieldsMap()); len(fields) > 0 {
		zl = zl.With(fields...)
	}
	return zl, nil
}
