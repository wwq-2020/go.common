package conf

import (
	"strconv"

	"github.com/wwq-2020/go.common/errors"
)

// errs
var (
	ErrKeyNil = errors.Std("key nil")
)

// Walker Walker
type Walker func(*Field) error

// NewWalker NewWalker
func NewWalker(validateFunc ValidateFunc) Walker {
	return func(field *Field) error {
		nullableTag := field.Field.Tag.Get("nullable")
		nullable, err := strconv.ParseBool(nullableTag)
		if err != nil {
			nullable = false
		}
		if nullable {
			return nil
		}
		if err := validateFunc(field); err != nil {
			return err
		}
		return nil
	}
}
