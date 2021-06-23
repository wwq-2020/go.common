package log

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wwq-2020/go.common/errors"
	"github.com/wwq-2020/go.common/stack"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger Logger
type Logger interface {
	Rotate()
	// Panic Panic
	Panic(string)
	// Panicf Panicf
	Panicf(string, ...interface{})
	// Fatal Fatal
	Fatal(string)
	// Fatalf Fatalf
	Fatalf(string, ...interface{})
	// Error Error
	Error(error)
	// Errorf Errorf
	Errorf(string, ...interface{})
	// Warn Warn
	Warn(string)
	// Warnf Warnf
	Warnf(string, ...interface{})
	// Info Info
	Info(string)
	// Infof Infof
	Infof(string, ...interface{})
	// Debug Debug
	Debug(string)
	// Debugf Debugf
	Debugf(string, ...interface{})
	// PanicContext PanicContext
	PanicContext(context.Context, string)
	// PanicfContext PanicfContext
	PanicfContext(context.Context, string, ...interface{})
	// FatalContext FatalContext
	FatalContext(context.Context, string)
	// FatalfContext FatalfContext
	FatalfContext(context.Context, string, ...interface{})
	// ErrorContext ErrorContext
	ErrorContext(context.Context, error)
	// ErrorfContext ErrorfContext
	ErrorfContext(context.Context, string, ...interface{})
	// WarnContext WarnContext
	WarnContext(context.Context, string)
	// WarnfContext WarnfContext
	WarnfContext(context.Context, string, ...interface{})
	// InfoContext InfoContext
	InfoContext(context.Context, string)
	// InfofContext InfofContext
	InfofContext(context.Context, string, ...interface{})
	// DebugContext DebugContext
	DebugContext(context.Context, string)
	// DebugfContext DebugfContext
	DebugfContext(context.Context, string, ...interface{})
	// WithFields WithFields
	WithFields(stack.Fields) Logger
	// WithError WithError
	WithError(err error) Logger
	// WithField WithField
	WithField(string, interface{}) Logger
	// WithZapFields WithZapFields
	WithZapFields(fields ...zap.Field) Logger
	// SetLevel SetLevel
	SetLevel(Level)
	// Sync
	Sync() error
	// Close
	Close() error
}

// Fields Fields
type Fields interface {
	Fields() stack.Fields
}

// Level Level
type Level int

const (
	// PanicLevel PanicLevel
	PanicLevel Level = iota
	// FatalLevel FatalLevel
	FatalLevel
	// ErrorLevel ErrorLevel
	ErrorLevel
	// WarnLevel WarnLevel
	WarnLevel
	// InfoLevel InfoLevel
	InfoLevel
	// DebugLevel DebugLevel
	DebugLevel
)

// Entry Entry
type logger struct {
	l *zap.Logger
	sync.RWMutex
	depth   int
	options *Options
}

// New 初始化Logger
func New(opts ...Option) Logger {
	return NewEx(0, opts...)
}

// WithOutput WithOutput
func WithOutput(output string) Option {
	return func(o *Options) {
		o.output = output
	}
}

// Option Option
type Option func(*Options)

// Options Options
type Options struct {
	output string
	level  Level
}

// SetOutput SetOutput
func SetOutput(output string) {
	defaultOptions.output = output
	std = NewEx(1)
	stdWith = NewEx(0)
}

// SetOutput SetOutput
func (l *logger) Rotate() {
	prevOutput := l.options.output
	l.Lock()
	prev := l.l
	prev.Sync()
	os.Rename(prevOutput, prevOutput+".1")
	zapLogger, _ := genZapConfig(l.options).Build()
	l.l = zapLogger
	l.Unlock()
}

// Std Std
func Std() Logger {
	return std
}

// NewEx 初始化Logger
func NewEx(depth int, opts ...Option) Logger {
	options := defaultOptions
	for _, opt := range opts {
		opt(&options)
	}
	zapLogger, _ := genZapConfig(&options).Build()
	return &logger{
		options: &options,
		depth:   depth,
		l:       zapLogger,
	}
}

func genZapConfig(options *Options) zap.Config {

	return zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
		DisableCaller:     true,
		DisableStacktrace: true,
		Development:       false,
		// Sampling: &zap.SamplingConfig{
		// 	Initial:    100,
		// 	Thereafter: 100,
		// },
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.LowercaseLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				time := t.Format("2006-01-02 15:04:05")
				enc.AppendString(time)
			},
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{options.output},
		ErrorOutputPaths: []string{options.output},
	}
}

var (
	std            = NewEx(1)
	stdWith        = NewEx(0)
	defaultOptions = Options{
		output: "stdout",
		level:  InfoLevel,
	}
)

// Infof Infof
func Infof(msg string, args ...interface{}) {
	std.Infof(msg, args...)
}

// Info Info
func Info(msg string) {
	std.Info(msg)
}

// Sync Sync
func Sync() error {
	if err := std.Sync(); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Errorf Errorf
func Errorf(msg string, args ...interface{}) {
	std.Errorf(msg, args...)
}

// Error Error
func Error(err error) {
	std.Error(err)
}

// Warnf Warnf
func Warnf(msg string, args ...interface{}) {
	std.Warnf(msg, args...)
}

// Warn Warn
func Warn(msg string) {
	std.Warn(msg)
}

// Debugf Debugf
func Debugf(msg string, args ...interface{}) {
	std.Debugf(msg, args...)
}

// Debug Debug
func Debug(msg string) {
	std.Debug(msg)
}

// Fatalf Fatalf
func Fatalf(msg string, args ...interface{}) {
	std.Fatalf(msg, args...)
}

// Fatal Fatal
func Fatal(msg string) {
	std.Fatal(msg)
}

// Panicf Panicf
func Panicf(msg string, args ...interface{}) {
	std.Panicf(msg, args...)
}

// Panic Panic
func Panic(msg string) {
	std.Panic(msg)
}

// InfofContext InfofContext
func InfofContext(ctx context.Context, msg string, args ...interface{}) {
	std.InfofContext(ctx, msg, args...)
}

// InfoContext InfoContext
func InfoContext(ctx context.Context, msg string) {
	std.InfoContext(ctx, msg)
}

// ErrorfContext ErrorfContext
func ErrorfContext(ctx context.Context, msg string, args ...interface{}) {
	std.ErrorfContext(ctx, msg, args...)
}

// ErrorContext ErrorContext
func ErrorContext(ctx context.Context, err error) {
	std.ErrorContext(ctx, err)
}

// WarnfContext WarnfContext
func WarnfContext(ctx context.Context, msg string, args ...interface{}) {
	std.WarnfContext(ctx, msg, args...)
}

// WarnContext WarnContext
func WarnContext(ctx context.Context, msg string) {
	std.WarnContext(ctx, msg)
}

// DebugfContext DebugfContext
func DebugfContext(ctx context.Context, msg string, args ...interface{}) {
	std.DebugfContext(ctx, msg, args...)
}

// DebugContext DebugContext
func DebugContext(ctx context.Context, msg string) {
	std.DebugContext(ctx, msg)
}

// FatalfContext FatalfContext
func FatalfContext(ctx context.Context, msg string, args ...interface{}) {
	std.FatalfContext(ctx, msg, args...)
}

// FatalContext FatalContext
func FatalContext(ctx context.Context, msg string) {
	std.FatalContext(ctx, msg)
}

// PanicfContext PanicfContext
func PanicfContext(ctx context.Context, msg string, args ...interface{}) {
	std.PanicfContext(ctx, msg, args...)
}

// PanicContext PanicContext
func PanicContext(ctx context.Context, msg string) {
	std.PanicContext(ctx, msg)
}

// SetLevel SetLevel
func SetLevel(level Level) {
	std.SetLevel(level)
}

// WithFields WithFields
func WithFields(fields stack.Fields) Logger {
	return stdWith.WithFields(fields)
}

// WithZapFields WithZapFields
func WithZapFields(fields ...zap.Field) Logger {
	return stdWith.WithZapFields(fields...)
}

// WithFieldsFromErr WithFieldsFromErr
func WithFieldsFromErr(err error) Logger {
	return stdWith.WithFields(errors.Fields(err))
}

// WithFieldsFrom WithFieldsFrom
func WithFieldsFrom(fields Fields) Logger {
	return stdWith.WithFields(fields.Fields())
}

// WithField WithField
func WithField(key string, val interface{}) Logger {
	return stdWith.WithField(key, val)
}

// WithError WithError
func WithError(err error) Logger {
	return stdWith.WithError(err)
}

// Infof Infof
func (l *logger) Infof(msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Info(fmt.Sprintf(msg, args...))
}

// Info Info
func (l *logger) Info(msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Info(msg)
}

// Errorf Errorf
func (l *logger) Errorf(msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Error(fmt.Sprintf(msg, args...))
}

// Error Error
func (l *logger) Error(err error) {
	zapFields := extractZapFieldsFromError(err)
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		With(zapFields...).
		Error(err.Error())
}

// Warnf Warnf
func (l *logger) Warnf(msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Warn(fmt.Sprintf(msg, args...))
}

// Warn Warn
func (l *logger) Warn(msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Warn(msg)
}

// Debugf Debugf
func (l *logger) Debugf(msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Debug(fmt.Sprintf(msg, args...))
}

// Debug Debug
func (l *logger) Debug(msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Debug(msg)
}

// Fatalf Fatalf
func (l *logger) Fatalf(msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Fatal(fmt.Sprintf(msg, args...))
}

// Fatal Fatal
func (l *logger) Fatal(msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Fatal(msg)
}

// Panicf Panicf
func (l *logger) Panicf(msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Panic(fmt.Sprintf(msg, args...))
}

// Panic Panic
func (l *logger) Panic(msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Panic(msg)
}

// InfofContext InfofContext
func (l *logger) InfofContext(ctx context.Context, msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).Info(fmt.Sprintf(msg, args...))
}

// InfoContext InfoContext
func (l *logger) InfoContext(ctx context.Context, msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Info(msg)
}

// ErrorfContext ErrorfContext
func (l *logger) ErrorfContext(ctx context.Context, msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Error(fmt.Sprintf(msg, args...))
}

// ErrorContext ErrorContext
func (l *logger) ErrorContext(ctx context.Context, err error) {
	zapFields := extractZapFieldsFromError(err)
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		With(zapFields...).Error(err.Error())
}

// WarnfContext WarnfContext
func (l *logger) WarnfContext(ctx context.Context, msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Warn(fmt.Sprintf(msg, args...))
}

// WarnContext WarnContext
func (l *logger) WarnContext(ctx context.Context, msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Warn(msg)
}

// DebugfContext DebugfContext
func (l *logger) DebugfContext(ctx context.Context, msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Debug(fmt.Sprintf(msg, args...))
}

// DebugContext DebugContext
func (l *logger) DebugContext(ctx context.Context, msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Debug(msg)
}

// FatalfContext FatalfContext
func (l *logger) FatalfContext(ctx context.Context, msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Fatal(fmt.Sprintf(msg, args...))
}

// FatalContext FatalContext
func (l *logger) FatalContext(ctx context.Context, msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Fatal(msg)
}

// PanicfContext PanicfContext
func (l *logger) PanicfContext(ctx context.Context, msg string, args ...interface{}) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Panic(fmt.Sprintf(msg, args...))
}

// PanicContext PanicContext
func (l *logger) PanicContext(ctx context.Context, msg string) {
	l.l.With(zap.String("caller", stack.Caller(l.depth+1))).
		Panic(msg)
}

func convToZapLevel(level Level) zap.AtomicLevel {
	switch level {
	case PanicLevel:
		return zap.NewAtomicLevelAt(zapcore.PanicLevel)
	case FatalLevel:
		return zap.NewAtomicLevelAt(zapcore.FatalLevel)
	case ErrorLevel:
		return zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case WarnLevel:
		return zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case InfoLevel:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case DebugLevel:
		return zap.NewAtomicLevelAt(zapcore.DebugLevel)
	default:
		return zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}
}

// SetLevel SetLevel
func (l *logger) SetLevel(level Level) {
	cfg := genZapConfig(l.options)
	switch level {
	case PanicLevel:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.PanicLevel)
	case FatalLevel:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	case ErrorLevel:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case WarnLevel:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case InfoLevel:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case DebugLevel:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}
	l.l, _ = cfg.Build()
}

// WithFields WithFields
func (l *logger) WithFields(fields stack.Fields) Logger {
	return &logger{
		l:     l.l.With(fields2ZapFields(fields)...),
		depth: l.depth,
	}
}

// WithFields WithFields
func (l *logger) WithZapFields(fields ...zap.Field) Logger {
	return &logger{
		l:     l.l.With(fields...),
		depth: l.depth,
	}
}

// WithError WithError
func (l *logger) WithError(err error) Logger {
	stack := errors.AllFields(err)
	return &logger{
		l:     l.l.With(fields2ZapFields(stack)...),
		depth: l.depth,
	}
}

// WithField WithField
func (l *logger) WithField(key string, val interface{}) Logger {
	return &logger{
		l:     l.l.With(zap.Any(key, val)),
		depth: l.depth,
	}
}

func (l *logger) Sync() error {
	return l.l.Sync()
}

func (l *logger) Close() error {
	return nil

}

func extractZapFieldsFromError(err error) []zap.Field {
	fields := errors.AllFields(err)
	return fields2ZapFields(fields)
}

func fields2ZapFields(fields stack.Fields) []zap.Field {
	kvs := fields.KVs()
	zapFields := make([]zap.Field, 0, len(kvs)+1)
	for k, v := range kvs {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

// String2Level String2Level
func String2Level(level string) Level {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "panic":
		return PanicLevel
	case "fatal":
		return FatalLevel
	}
	return InfoLevel
}
