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
	logger.Panic(msg)
}

// PanicfContext PanicfContext
func PanicfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.Panicf(msg, args...)
}

// FatalContext FatalContext
func FatalContext(ctx context.Context, msg string) {
	logger := LoggerFromCtx(ctx)
	logger.Fatal(msg)
}

// FatalfContext FatalfContext
func FatalfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.Fatalf(msg, args...)
}

// ErrorContext ErrorContext
func ErrorContext(ctx context.Context, err error) {
	logger := LoggerFromCtx(ctx)
	logger.Error(err)
}

// ErrorfContext ErrorfContext
func ErrorfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.Errorf(msg, args...)
}

// WarnContext WarnContext
func WarnContext(ctx context.Context, msg string) {
	logger := LoggerFromCtx(ctx)
	logger.Warn(msg)
}

// WarnfContext WarnfContext
func WarnfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.Warnf(msg, args...)
}

// InfoContext InfoContext
func InfoContext(ctx context.Context, msg string) {
	logger := LoggerFromCtx(ctx)
	logger.Info(msg)
}

// InfofContext InfofContext
func InfofContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.Infof(msg, args...)
}

// DebugContext DebugContext
func DebugContext(ctx context.Context, msg string) {
	logger := LoggerFromCtx(ctx)
	logger.Debug(msg)
}

// DebugfContext DebugfContext
func DebugfContext(ctx context.Context, msg string, args ...interface{}) {
	logger := LoggerFromCtx(ctx)
	logger.Debugf(msg, args...)
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
