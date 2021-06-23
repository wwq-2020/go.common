package log_test

import (
	"testing"

	"github.com/wwq-2020/go.common/log"
)

func TestLogger(t *testing.T) {
	log.WithField("a", "b").
		WithField("c", "b").
		Info("x")
}
