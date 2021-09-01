package log

import (
	"context"

	"github.com/wwq-2020/go.common/stack"
	"go.uber.org/zap"
)

// NoopLogger NoopLogger
type NoopLogger struct{}

// Panic Panic
func (NoopLogger) Panic(string) {

}

// Panicf Panicf
func (NoopLogger) Panicf(string, ...interface{}) {}

// Fatal Fatal
func (NoopLogger) Fatal(string) {}

// Fatalf Fatalf
func (NoopLogger) Fatalf(string, ...interface{}) {}

// Error Error
func (NoopLogger) Error(error) {}

// Errorf Errorf
func (NoopLogger) Errorf(string, ...interface{}) {}

// Warn Warn
func (NoopLogger) Warn(string) {}

// Warnf Warnf
func (NoopLogger) Warnf(string, ...interface{}) {}

// Info Info
func (NoopLogger) Info(string) {}

// Infof Infof
func (NoopLogger) Infof(string, ...interface{}) {}

// Debug Debug
func (NoopLogger) Debug(string) {}

// Debugf Debugf
func (NoopLogger) Debugf(string, ...interface{}) {}

// PanicContext PanicContext
func (NoopLogger) PanicContext(context.Context, string) {}

// PanicfContext PanicfContext
func (NoopLogger) PanicfContext(context.Context, string, ...interface{}) {}

// FatalContext FatalContext
func (NoopLogger) FatalContext(context.Context, string) {}

// FatalfContext FatalfContext
func (NoopLogger) FatalfContext(context.Context, string, ...interface{}) {}

// ErrorContext ErrorContext
func (NoopLogger) ErrorContext(context.Context, error) {}

// ErrorfContext ErrorfContext
func (NoopLogger) ErrorfContext(context.Context, string, ...interface{}) {}

// WarnContext WarnContext
func (NoopLogger) WarnContext(context.Context, string) {}

// WarnfContext WarnfContext
func (NoopLogger) WarnfContext(context.Context, string, ...interface{}) {}

// InfoContext InfoContext
func (NoopLogger) InfoContext(context.Context, string) {}

// InfofContext InfofContext
func (NoopLogger) InfofContext(context.Context, string, ...interface{}) {}

// DebugContext DebugContext
func (NoopLogger) DebugContext(context.Context, string) {}

// DebugfContext DebugfContext
func (NoopLogger) DebugfContext(context.Context, string, ...interface{}) {}

// WithFields WithFields
func (l NoopLogger) WithFields(stack.Fields) Logger {
	return l
}

// WithError WithError
func (l NoopLogger) WithError(err error) Logger {
	return l
}

// WithField WithField
func (l NoopLogger) WithField(string, interface{}) Logger {
	return l
}

// WithZapFields WithZapFields
func (l NoopLogger) WithZapFields(fields ...zap.Field) Logger {
	return l
}

// SetLevel SetLevel
func (NoopLogger) SetLevel(Level) {}

// SetStringLevel SetStringLevel
func (NoopLogger) SetStringLevel(string) {}

// Sync Sync
func (NoopLogger) Sync() error {
	return nil
}

// Dup Dup
func (l NoopLogger) Dup() Logger {
	return l
}

// AddDep AddDep
func (l NoopLogger) AddDep(int) Logger {
	return l
}

// Close Close
func (NoopLogger) Close() error {
	return nil
}

// WithStack WithStack
func (l NoopLogger) WithStack() Logger {
	return l
}
