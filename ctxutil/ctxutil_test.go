package ctxutil_test

import (
	"context"
	"testing"

	"github.com/wwq-2020/go.common/ctxutil"
	"github.com/wwq-2020/go.common/errors"
)

func a() error {
	return errors.New("ss")
}

func TestInfo(t *testing.T) {
	ctx := ctxutil.WithField(context.TODO(), "a", "b")
	err := a()
	ctxutil.ErrorContext(ctx, err)
}
