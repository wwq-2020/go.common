package errorsx

import (
	"errors"

	"github.com/wwq-2020/go.common/stack"
)

// consts
const (
	UnknownCode int = -1
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
	code   int
	tip    string
}

// New New
func New(msg string) error {
	return trace(errors.New(msg), UnknownCode, nil)
}

// NewWithFields NewWithFields
func NewWithFields(msg string, fields stack.Fields) error {
	return trace(errors.New(msg), UnknownCode, fields)
}

// NewWithField NewWithField
func NewWithField(msg string, key string, val interface{}) error {
	fields := stack.New().Set(key, val)
	return trace(errors.New(msg), UnknownCode, fields)
}

// NewWithCode NewWithCode
func NewWithCode(msg string, code int) error {
	return trace(errors.New(msg), code, nil)
}

// NewWithCodeWithFields NewWithCodeWithFields
func NewWithCodeWithFields(msg string, code int, fields stack.Fields) error {
	return trace(errors.New(msg), code, fields)
}

// NewWithCodeWithField NewWithCodeWithField
func NewWithCodeWithField(msg string, code int, key string, val interface{}) error {
	fields := stack.New().Set(key, val)
	return trace(errors.New(msg), code, fields)
}

// Trace Trace
func Trace(err error) error {
	if err == nil {
		return nil
	}
	return trace(err, UnknownCode, nil)
}

// Std Std
func Std(msg string) error {
	return errors.New(msg)
}

func trace(err error, code int, fields stack.Fields) error {
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

// WithField WithField
func (s *stackError) WithField(key string, val interface{}) {
	s.fields.Set(key, val)
}

// WithField WithFields
func (s *stackError) WithFields(fields stack.Fields) {
	s.fields = s.fields.Merge(fields)
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

func (s *stackError) CodeIs(code int) bool {
	return s.code == code
}

func (s *stackError) Code() int {
	return s.code
}

// TraceWithFields TraceWithFields
func TraceWithFields(err error, fields stack.Fields) error {
	return trace(err, UnknownCode, fields)
}

// TraceWithField TraceWithField
func TraceWithField(err error, key string, val interface{}) error {
	fields := stack.New().Set(key, val)
	return TraceWithFields(err, fields)
}

// TraceWithCode TraceWithCode
func TraceWithCode(err error, code int) error {
	return trace(err, code, nil)
}

// TraceWithCodeWithField TraceWithCodeWithField
func TraceWithCodeWithField(err error, code int, key string, val interface{}) error {
	return trace(err, code, nil)
}

// TraceWithCodeWithFields TraceWithCodeWithFields
func TraceWithCodeWithFields(err error, code int, fields stack.Fields) error {
	return trace(err, code, fields)
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

// Replace Replace
func Replace(raw, err error) error {
	se, ok := raw.(*stackError)
	if !ok {
		return trace(err, UnknownCode, nil)
	}
	se.err = err
	return se
}

// CodeIs CodeIs
func CodeIs(err error, code int) bool {
	se, ok := err.(*stackError)
	return ok && se.CodeIs(code)
}

// Code Code
func Code(err error) int {
	se, ok := err.(*stackError)
	if !ok {
		return UnknownCode
	}
	return se.Code()
}

// ReplaceCode ReplaceCode
func ReplaceCode(raw error, code int) error {
	se, ok := raw.(*stackError)
	if !ok {
		return trace(raw, code, nil)
	}
	se.code = code
	return se
}

// ReplaceCodeWithFields ReplaceCodeWithFields
func ReplaceCodeWithFields(raw error, code int, fields stack.Fields) error {
	se, ok := raw.(*stackError)
	if !ok {
		return trace(raw, code, fields)
	}
	se.code = code
	se.fields = se.fields.Merge(fields)
	return se
}

// ReplaceCodeWithField ReplaceCodeWithField
func ReplaceCodeWithField(raw error, code int, key, value string) error {
	se, ok := raw.(*stackError)
	if !ok {
		fields := stack.New().Set(key, value)
		return trace(raw, code, fields)
	}
	se.code = code
	se.fields.Set(key, value)
	return se
}

// Tip Tip
func Tip(raw error) string {
	se, ok := raw.(*stackError)
	if !ok {
		return raw.Error()
	}
	return se.tip
}

// ReplaceTip ReplaceTip
func ReplaceTip(raw error, tip string) error {
	se, ok := raw.(*stackError)
	if !ok {
		return errors.New(tip)
	}
	se.tip = tip
	return se
}

// ReplaceTipWithFields ReplaceTipWithFields
func ReplaceTipWithFields(raw error, tip string, fields stack.Fields) error {
	se, ok := raw.(*stackError)
	if !ok {
		return NewWithFields(tip, fields)
	}
	se.tip = tip
	se.fields = se.fields.Merge(fields)
	return se
}

// ReplaceTipWithField ReplaceTipWithField
func ReplaceTipWithField(raw error, tip, key, value string) error {
	se, ok := raw.(*stackError)
	if !ok {
		return NewWithField(tip, key, value)
	}
	se.tip = tip
	se.fields.Set(key, value)
	return se
}
