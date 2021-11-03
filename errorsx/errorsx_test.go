package errorsx_test

import (
	"errors"
	"testing"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
)

func demo() error {
	return errorsx.Trace(errors.New("sss"))
}

func TestTrace(t *testing.T) {
	log.Error(errorsx.Trace(demo()))
}
