package errors

import (
	"errors"

	"github.com/wwq-2020/go.common/stack"
)

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
}

// New New
func New(msg string) error {
	return trace(errors.New(msg), nil)
}

// NewWithFields NewWithFields
func NewWithFields(msg string, fields stack.Fields) error {
	return trace(errors.New(msg), fields)
}

// NewWithField NewWithField
func NewWithField(msg string, key string, val interface{}) error {
	fields := stack.New().Set(key, val)
	return trace(errors.New(msg), fields)
}

// Trace Trace
func Trace(err error) error {
	if err == nil {
		return nil
	}
	return trace(err, nil)
}

// Std Std
func Std(msg string) error {
	return errors.New(msg)
}

func trace(err error, fields stack.Fields) error {
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
		}
	}
	if se.fields == nil {
		se.fields = stack.New()
	}
	if fields != nil {
		se.fields.Merge(fields)
	}
	return se
}

// WithField WithField
func (s *stackError) WithField(key string, val interface{}) {
	s.fields.Set(key, val)
}

// WithField WithFields
func (s *stackError) WithFields(fields stack.Fields) {
	s.fields.Merge(fields)
}

// Fields Fields
func (s *stackError) Fields() stack.Fields {
	return s.fields
}

// StackFields StackFields
func (s *stackError) StackFields() stack.Fields {
	stack := stack.New().Set("stack", s.stack)
	stack = stack.Merge(s.fields)
	return stack
}

// FullFields FullFields
func (s *stackError) AllFields() stack.Fields {
	stack := stack.New().
		Set("stack", s.stack).
		Set("err", s.err.Error())
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

// TraceWithFields TraceWithFields
func TraceWithFields(err error, fields stack.Fields) error {
	return trace(err, fields)
}

// TraceWithField TraceWithField
func TraceWithField(err error, key string, val interface{}) error {
	fields := stack.New().Set(key, val)
	return TraceWithFields(err, fields)
}

// Is Is
func Is(src, dst error) bool {
	return errors.Is(src, dst)
}

// Unwrap Unwrap
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// As As
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// IsTimeout IsTimeout
func IsTimeout(err error) bool {
	timeoutErr, ok := err.(timeout)
	return ok && timeoutErr.Timeout()
}

// IsTemporary IsTemporary
func IsTemporary(err error) bool {
	temporaryErr, ok := err.(temporary)
	return ok && temporaryErr.Temporary()
}

// Fields Fields
func Fields(err error) stack.Fields {
	if stackErr, ok := err.(*stackError); ok {
		return stackErr.Fields()
	}
	return stack.New()
}

// AllFields AllFields
func AllFields(err error) stack.Fields {
	if stackErr, ok := err.(*stackError); ok {
		return stackErr.AllFields()
	}
	return stack.New().Set("err", err.Error())
}

// StackFields StackFields
func StackFields(err error) stack.Fields {
	if stackErr, ok := err.(*stackError); ok {
		return stackErr.StackFields()
	}
	s := stack.New()
	s.Set("stack", stack.Callers(stack.StdFilter))
	return s
}

// CanAs CanAs
func CanAs(err error) bool {
	var se *stackError
	return As(err, &se)
}
