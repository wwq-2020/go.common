package tracing

import "github.com/wwq-2020/go.common/log"

type jaegerLogger struct{}

func (l *jaegerLogger) Error(msg string) {
	log.Errorf(msg)
}

func (l *jaegerLogger) Infof(msg string, args ...interface{}) {
	log.Infof(msg, args)
}
