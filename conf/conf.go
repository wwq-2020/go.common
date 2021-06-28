package conf

import (
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/wwq-2020/go.common/errors"
	"github.com/wwq-2020/go.common/log"
)

var (
	flagger = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	cfgPath = flagger.String("conf", "conf.toml", "-conf=conf.toml")
)

func init() {
	flagger.Parse(os.Args[1:])
}

// Field Field
type Field struct {
	Field     reflect.StructField
	Value     reflect.Value
	Ancestors []string
}

// MustLoad MustLoad
func MustLoad(dest interface{}) {
	MustParseFile(*cfgPath, dest)
}

// MustParseFile MustParseFile
func MustParseFile(file string, dest interface{}) {
	if err := ParseFile(file, dest); err != nil {
		log.WithError(err).
			Fatal("failed to ParseFile")
	}
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

// MustParse MustParse
func MustParse(data []byte, dest interface{}) {
	if err := Parse(data, dest); err != nil {
		log.WithError(err).
			Fatal("failed to Parse")
	}
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
		keys := append(field.Ancestors, field.Field.Tag.Get("toml"))
		key := strings.Join(keys, ".")
		if !metadata.IsDefined(keys...) {
			return errors.TraceWithField(ErrKeyNil, key, strings.Join(append(field.Ancestors, field.Field.Name), "."))
		}
		return nil
	})
	if err := ParseBy(data, dest, "toml", decodeFunc, walker); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// MustParseBy MustParseBy
func MustParseBy(data []byte, dest interface{}, tag string, decodeFunc DecodeFunc, walker Walker) {
	if err := ParseBy(data, dest, tag, decodeFunc, walker); err != nil {
		log.WithError(err).
			Fatal("failed to ParseBy")
	}
}

// ParseBy ParseBy
func ParseBy(data []byte, dest interface{}, tag string, decodeFunc DecodeFunc, walker Walker) error {
	if err := decodeFunc(data, dest); err != nil {
		return errors.Trace(err)
	}
	if err := Walk(dest, tag, walker); err != nil {
		return errors.Trace(err)
	}
	return nil
}

// MustWalk MustWalk
func MustWalk(dest interface{}, tag string, walker Walker) {
	if err := Walk(dest, tag, walker); err != nil {
		log.WithError(err).
			Fatal("failed to Walk")
	}
}

// Walk Walk
func Walk(dest interface{}, tag string, walker Walker) error {
	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)
	if err := walk(t, v, nil, tag, walker); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func walk(t reflect.Type, v reflect.Value, ancestors []string, tag string, walker Walker) error {
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
		ancestors = append(ancestors, field.Tag.Get(tag))
		if err := walk(fieldType, fieldValue, ancestors, tag, walker); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}
