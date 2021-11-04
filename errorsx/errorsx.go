package errorsx

import (
	"errors"

	"github.com/wwq-2020/go.common/errcode"
	"github.com/wwq-2020/go.common/stack"
)

type StackError interface {
	error
	WithCode(errcode.ErrCode) StackError
	Code() errcode.ErrCode
	CodeIs(errcode.ErrCode) bool
	WithField(string, interface{}) StackError
	WithFields(stack.Fields) StackError
	Fields() stack.Fields
	FullFields() stack.Fields
	WithTip(string) StackError
	Tip() string
}

// timeout timeout
type timeout interface {
	Timeout() bool
}

// temporary temporary
type temporary interface {
	Temporary() bool
}

type stackError struct {
	fields stack.Fields
	stack  string
	err    error
	code   errcode.ErrCode
	tip    string
}

// New New
func New(msg string) StackError {
	return trace(errors.New(msg), errcode.ErrCode_Unknown, nil)
}

// Trace Trace
func Trace(err error) StackError {
	return As(err)
}

// Std Std
func Std(msg string) error {
	return errors.New(msg)
}

func trace(err error, code errcode.ErrCode, fields stack.Fields) StackError {
	se, ok := err.(*stackError)
	if !ok {
		stackFrame := stack.Callers(stack.StdFilter)
		if fields == nil {
			fields = stack.New()
		}
		return &stackError{
			err:    err,
			fields: fields,
			stack:  stackFrame,
			code:   code,
			tip:    err.Error(),
		}
	}
	if se.fields == nil {
		se.fields = stack.New()
	}
	if fields != nil {
		se.fields = se.fields.Merge(fields)
	}
	return se
}

func (s *stackError) WithField(key string, val interface{}) StackError {
	s.fields.Set(key, val)
	return s
}

func (s *stackError) WithFields(fields stack.Fields) StackError {
	s.fields = s.fields.Merge(fields)
	return s
}

func (s *stackError) Fields() stack.Fields {
	return s.fields
}

func (s *stackError) FullFields() stack.Fields {
	stack := stack.New().
		Set("stack", s.stack).
		Set("tip", s.tip)
	stack = stack.Merge(s.fields)
	return stack
}

func (s *stackError) Unwrap() error {
	return s.err
}

func (s *stackError) Is(err error) bool {
	return s.err == err
}

func (s *stackError) Error() string {
	return s.err.Error()
}

func (s *stackError) CodeIs(code errcode.ErrCode) bool {
	return s.code == code
}

func (s *stackError) WithCode(code errcode.ErrCode) StackError {
	s.code = code
	return s
}

func (s *stackError) Code() errcode.ErrCode {
	return s.code
}

func (s *stackError) WithTip(tip string) StackError {
	s.tip = tip
	return s
}

func (s *stackError) Tip() string {
	return s.tip
}

// Is Is
func StdIs(src, dst error) bool {
	return errors.Is(src, dst)
}

// Unwrap Unwrap
func StdUnwrap(err error) error {
	return errors.Unwrap(err)
}

// As As
func StdAs(err error, target interface{}) bool {
	return errors.As(err, target)
}

func As(err error) StackError {
	if err == nil {
		return nil
	}
	se, ok := err.(StackError)
	if !ok {
		return trace(err, errcode.ErrCode_Unknown, nil)
	}
	return se
}

// IsTimeout IsTimeout
func IsTimeout(err error) bool {
	var timeoutErr timeout
	if StdAs(err, &timeoutErr) {
		return timeoutErr.Timeout()
	}
	return false
}

// IsTemporary IsTemporary
func IsTemporary(err error) bool {
	var temporaryErr temporary
	if StdAs(err, &temporaryErr) {
		return temporaryErr.Temporary()
	}
	return false
}

// Fields Fields
func Fields(err error) stack.Fields {
	return As(err).FullFields()
}

// FullFields FullFields
func FullFields(err error) stack.Fields {
	return As(err).FullFields()
}

// CanAs CanAs
func CanAs(err error) bool {
	var se StackError
	return StdAs(err, &se)
}

// Code Code
func CodeIs(err error, errcode errcode.ErrCode) bool {
	return As(err).CodeIs(errcode)
}

// Code Code
func Code(err error) errcode.ErrCode {
	return As(err).Code()
}

// WithTip WithTip
func WithCode(err error, code errcode.ErrCode) StackError {
	return As(err).WithCode(code)
}

// Tip Tip
func Tip(err error) string {
	return As(err).Tip()
}

// WithTip WithTip
func WithTip(err error, tip string) StackError {
	return As(err).WithTip(tip)
}

func WithField(err error, key string, value interface{}) StackError {
	return As(err).WithField(key, value)
}

func WithFields(err error, stack stack.Fields) StackError {
	return As(err).WithFields(stack)
}
