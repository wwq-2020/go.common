package conf

import (
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/wwq-2020/go.common/errors"
)

// Field Field
type Field struct {
	Field     reflect.StructField
	Value     reflect.Value
	Ancestors []string
}

// ParseFile ParseFile
func ParseFile(file string, dest interface{}) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.Trace(err)
	}
	if err := Parse(data, dest); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Parse Parse
func Parse(data []byte, dest interface{}) error {
	var metadata toml.MetaData
	decodeFunc := func(data []byte, dest interface{}) error {
		var err error
		metadata, err = toml.Decode(string(data), dest)
		if err != nil {
			return errors.Trace(err)
		}
		return nil
	}
	walker := NewWalker(func(field *Field) error {
		key := strings.Join(append(field.Ancestors, field.Field.Tag.Get("toml")), ".")
		if !metadata.IsDefined(key) {
			return errors.TraceWithField(ErrNotExist, key, strings.Join(append(field.Ancestors, field.Field.Name), "."))
		}
		return nil
	})
	if err := ParseBy(data, dest, decodeFunc, walker); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// ParseBy ParseBy
func ParseBy(data []byte, dest interface{}, decodeFunc DecodeFunc, walker Walker) error {
	if err := decodeFunc(data, dest); err != nil {
		return errors.Trace(err)
	}
	if err := Walk(dest, walker); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// Walk Walk
func Walk(dest interface{}, walker Walker) error {
	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)
	if err := walk(t, v, nil, walker); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func walk(t reflect.Type, v reflect.Value, ancestors []string, walker Walker) error {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type
		fieldValue := v.Field(i)
		if err := walker(&Field{
			Field:     field,
			Value:     v.Field(i),
			Ancestors: ancestors,
		}); err != nil {
			return errors.Trace(err)
		}

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() != reflect.Struct {
			continue
		}
		ancestors = append(ancestors, field.Name)
		if err := walk(fieldType, fieldValue, ancestors, walker); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}
