package log

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/stack"
	"go.uber.org/zap"
)

type traceInfoKey struct{}

// ContextWithTraceID ContextWithTraceID
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceInfoKey{}, traceID)
}

// ContextWithTraceIDWithFun ContextWithTraceIDWithFun
func ContextWithTraceIDWithFun(ctx context.Context, fun func() string) context.Context {
	return ContextWithTraceID(ctx, fun())
}

// ContextWithTraceIDWithFunx ContextWithTraceIDWithFunx
func ContextWithTraceIDWithFunx(ctx context.Context, fun func() string) (context.Context, string) {
	traceID := fun()
	return ContextWithTraceID(ctx, traceID), traceID
}

// ContextEnsureTraceID ContextEnsureTraceID
func ContextEnsureTraceID(ctx context.Context) context.Context {
	return ContextEnsureTraceIDWithGen(ctx, GenTraceID)
}

// ContextEnsureTraceIDx ContextEnsureTraceIDx
func ContextEnsureTraceIDx(ctx context.Context) (context.Context, string) {
	return ContextEnsureTraceIDWithGenx(ctx, GenTraceID)
}

// ContextEnsureTraceIDWithGen ContextEnsureTraceIDWithGen
func ContextEnsureTraceIDWithGen(ctx context.Context, fun func() string) context.Context {
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		return ContextWithTraceIDWithFun(ctx, fun)
	}
	return ctx
}

// ContextEnsureTraceIDWithGenx ContextEnsureTraceIDWithGenx
func ContextEnsureTraceIDWithGenx(ctx context.Context, fun func() string) (context.Context, string) {
	traceID := TraceIDFromContext(ctx)
	if traceID == "" {
		return ContextWithTraceIDWithFunx(ctx, fun)
	}
	return ctx, traceID
}

// TraceIDFromContext TraceIDFromContext
func TraceIDFromContext(ctx context.Context) string {
	traceIDObj := ctx.Value(traceInfoKey{})
	if traceIDObj == nil {
		return ""
	}
	return traceIDObj.(string)
}

var (
	seq uint64
	pid = os.Getpid()
)

// GenTraceID GenTraceID
func GenTraceID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%d.%d.%d", timestamp, pid, atomic.AddUint64(&seq, 1))
}

// DupContext DupContext
func DupContext(ctx context.Context) context.Context {
	logger := LoggerFromContext(ctx)
	traceID := TraceIDFromContext(ctx)
	ctx = ContextWithLogger(context.TODO(), logger)
	return ContextWithTraceID(ctx, traceID)
}

func zapFieldsFromContext(ctx context.Context) []zap.Field {
	return []zap.Field{
		zap.String("traceID", TraceIDFromContext(ctx)),
	}
}

func fields2ZapFields(fields stack.Fields) []zap.Field {
	kvs := fields.KVs()
	zapFields := make([]zap.Field, 0, len(kvs))
	for k, v := range kvs {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

func zapFieldsFromError(err error) []zap.Field {
	fields := errorsx.FullFields(err)
	return fields2ZapFields(fields)
}
