package errorsx

import (
	"github.com/wwq-2020/go.common/errcode"
	"github.com/wwq-2020/go.common/stack"
)

type noop struct {
}

func (n *noop) WithField(key string, val interface{}) StackError {
	return n
}

func (n *noop) WithFields(fields stack.Fields) StackError {
	return n
}

func (n *noop) Fields() stack.Fields {
	return stack.New()
}

func (n *noop) FullFields() stack.Fields {
	return stack.New()
}

func (n *noop) Unwrap() error {
	return n
}

func (n *noop) Is(err error) bool {
	return err == nil
}

func (n *noop) Error() string {
	return ""
}

func (n *noop) CodeIs(code errcode.ErrCode) bool {
	return code == errcode.ErrCode_Unknown
}

func (n *noop) WithCode(code errcode.ErrCode) StackError {
	return n
}

func (n *noop) Code() errcode.ErrCode {
	return errcode.ErrCode_Unknown
}

func (n *noop) WithTip(tip string) StackError {
	return n
}

func (n *noop) Tip() string {
	return ""
}
