package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (l Level) toZapLevel() zap.AtomicLevel {
	switch l {
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
