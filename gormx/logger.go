package gormx

import (
	"context"
	"fmt"
	"time"

	"github.com/wwq-2020/go.common/log"
	"github.com/wwq-2020/go.common/stack"
	"gorm.io/gorm/logger"
)

// GormLogger GormLogger
type GormLogger struct {
	log.Logger
}

// LogMode LogMode
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := l.Logger.Dup()
	switch level {
	case logger.Silent:
		newLogger = log.NoopLogger{}
	case logger.Error:
		newLogger.SetLevel(log.ErrorLevel)
	case logger.Warn:
		newLogger.SetLevel(log.WarnLevel)
	case logger.Info:
		newLogger.SetLevel(log.InfoLevel)
	}
	return &GormLogger{
		Logger: newLogger,
	}
}

// Info Info
func (l *GormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.InfoContext(ctx, fmt.Sprintf(msg, args...))
}

// Warn Warn
func (l *GormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.InfoContext(ctx, fmt.Sprintf(msg, args...))
}

// Error Error
func (l *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	l.Logger.ErrorfContext(ctx, fmt.Sprintf(msg, args...))
}

// Trace Trace
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, rowsAffected := fc()
	stack := stack.New().
		Set("begin", begin.Format("2006-01-02 15:04:05")).
		Set("sql", sql).
		Set("rowsAffected", rowsAffected)
	if err != nil {
		l.Logger.WithFields(stack).ErrorContext(ctx, err)
		return
	}
	l.Logger.WithFields(stack).InfoContext(ctx, "trace")
}
