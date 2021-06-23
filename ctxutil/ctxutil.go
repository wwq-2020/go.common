package ctxutil

import (
	"context"

	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
)

type loggerKey struct{}

// WithField WithField
func WithField(ctx context.Context, key string, value interface{}) context.Context {
	logger := loggerFromCtx(ctx)
	logger = logger.WithField(key, value)
	return context.WithValue(ctx, loggerKey{}, logger)
}

// WithFields WithFields
func WithFields(ctx context.Context, fields stack.Fields) context.Context {
	logger := loggerFromCtx(ctx)
	logger = logger.WithFields(fields)
	return context.WithValue(ctx, loggerKey{}, logger)
}

// WithError WithError
func WithError(ctx context.Context, err error) context.Context {
	logger := loggerFromCtx(ctx)
	logger = logger.WithError(err)
	return context.WithValue(ctx, loggerKey{}, logger)
}

func loggerFromCtx(ctx context.Context) log.Logger {
	logger := ctx.Value(loggerKey{})
	if logger != nil {
		return logger.(log.Logger)
	}
	return log.Std()
}

// PanicContext PanicContext
func PanicContext(ctx context.Context, msg string) {
	logger := loggerFromCtx(ctx)
	logger.PanicContext(ctx, msg)
}

// PanicfContext PanicfContext
func PanicfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := loggerFromCtx(ctx)
	logger.PanicfContext(ctx, msg, args...)
}

// FatalContext FatalContext
func FatalContext(ctx context.Context, msg string) {
	logger := loggerFromCtx(ctx)
	logger.FatalContext(ctx, msg)
}

// FatalfContext FatalfContext
func FatalfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := loggerFromCtx(ctx)
	logger.FatalfContext(ctx, msg, args...)
}

// ErrorContext ErrorContext
func ErrorContext(ctx context.Context, err error) {
	logger := loggerFromCtx(ctx)
	logger.ErrorContext(ctx, err)
}

// ErrorfContext ErrorfContext
func ErrorfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := loggerFromCtx(ctx)
	logger.ErrorfContext(ctx, msg, args...)
}

// WarnContext WarnContext
func WarnContext(ctx context.Context, msg string) {
	logger := loggerFromCtx(ctx)
	logger.WarnContext(ctx, msg)
}

// WarnfContext WarnfContext
func WarnfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := loggerFromCtx(ctx)
	logger.WarnfContext(ctx, msg, args...)
}

// InfoContext InfoContext
func InfoContext(ctx context.Context, msg string) {
	logger := loggerFromCtx(ctx)
	logger.InfoContext(ctx, msg)
}

// InfofContext InfofContext
func InfofContext(ctx context.Context, msg string, args ...interface{}) {
	logger := loggerFromCtx(ctx)
	logger.InfofContext(ctx, msg, args...)
}

// DebugContext DebugContext
func DebugContext(ctx context.Context, msg string) {
	logger := loggerFromCtx(ctx)
	logger.DebugContext(ctx, msg)
}

// DebugfContext DebugfContext
func DebugfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := loggerFromCtx(ctx)
	logger.DebugfContext(ctx, msg, args...)
}
