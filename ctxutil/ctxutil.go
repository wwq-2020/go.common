package ctxutil

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
)

type loggerKey struct{}

// WithField WithField
func WithField(ctx context.Context, key string, value interface{}) context.Context {
	logger := LoggerFromCtx(ctx)
	logger = logger.WithField(key, value)
	return WithLogger(ctx, logger)
}

// WithFields WithFields
func WithFields(ctx context.Context, fields stack.Fields) context.Context {
	logger := LoggerFromCtx(ctx)
	logger = logger.WithFields(fields)
	return WithLogger(ctx, logger)
}

// WithError WithError
func WithError(ctx context.Context, err error) context.Context {
	logger := LoggerFromCtx(ctx)
	logger = logger.WithError(err)
	return WithLogger(ctx, logger)
}

// WithLogger WithLogger
func WithLogger(ctx context.Context, logger log.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// LoggerFromCtx LoggerFromCtx
func LoggerFromCtx(ctx context.Context) log.Logger {
	logger := ctx.Value(loggerKey{})
	if logger != nil {
		return logger.(log.Logger)
	}
	return log.Std()
}

// PanicContext PanicContext
func PanicContext(ctx context.Context, msg string) {
	logger := LoggerFromCtx(ctx)
	logger.PanicContext(ctx, msg)
}

// PanicfContext PanicfContext
func PanicfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.PanicfContext(ctx, msg, args...)
}

// FatalContext FatalContext
func FatalContext(ctx context.Context, msg string) {
	logger := LoggerFromCtx(ctx)
	logger.FatalContext(ctx, msg)
}

// FatalfContext FatalfContext
func FatalfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.FatalfContext(ctx, msg, args...)
}

// ErrorContext ErrorContext
func ErrorContext(ctx context.Context, err error) {
	logger := LoggerFromCtx(ctx)
	logger.ErrorContext(ctx, err)
}

// ErrorfContext ErrorfContext
func ErrorfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.ErrorfContext(ctx, msg, args...)
}

// WarnContext WarnContext
func WarnContext(ctx context.Context, msg string) {
	logger := LoggerFromCtx(ctx)
	logger.WarnContext(ctx, msg)
}

// WarnfContext WarnfContext
func WarnfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.WarnfContext(ctx, msg, args...)
}

// InfoContext InfoContext
func InfoContext(ctx context.Context, msg string) {
	logger := LoggerFromCtx(ctx)
	logger.InfoContext(ctx, msg)
}

// InfofContext InfofContext
func InfofContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.InfofContext(ctx, msg, args...)
}

// DebugContext DebugContext
func DebugContext(ctx context.Context, msg string) {
	logger := LoggerFromCtx(ctx)
	logger.DebugContext(ctx, msg)
}

// DebugfContext DebugfContext
func DebugfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.DebugfContext(ctx, msg, args...)
}

// EnsureRequestID EnsureRequestID
func EnsureRequestID(ctx context.Context) context.Context {
	return EnsureRequestIDWithFun(ctx, GenRequestID)
}

type requestIDKey struct{}

// EnsureRequestIDWithFun EnsureRequestIDWithFun
func EnsureRequestIDWithFun(ctx context.Context, fn func() string) context.Context {
	requestID := fn()
	return WithField(ctx, "requestID", requestID)
}

var (
	seq uint64
	pid = os.Getpid()
)

// GenRequestID GenRequestID
func GenRequestID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d.%d.%d", timestamp, pid, atomic.AddUint64(&seq, 1))
}
