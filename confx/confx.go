package confx

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/log"
	etcd "go.etcd.io/etcd/client/v3"
)

// Field Field
type Field struct {
	Field     reflect.StructField
	Value     reflect.Value
	Ancestors []string
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
		return errorsx.Trace(err)
	}
	if err := Parse(data, dest); err != nil {
		return errorsx.Trace(err)
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
			return errorsx.Trace(err)
		}
		return nil
	}
	walker := NewWalker(func(field *Field) error {
		parts := strings.Split(field.Field.Tag.Get("toml"), ",")
		curKey := parts[0]
		keys := field.Ancestors
		if !hasInlineTag(parts[1:]) {
			keys = append(field.Ancestors, curKey)
		}

		key := strings.Join(keys, ".")
		if !metadata.IsDefined(keys...) {
			return errorsx.TraceWithField(ErrKeyNil, "key", key)
		}
		return nil
	})
	if err := ParseBy(data, dest, "toml", decodeFunc, walker); err != nil {
		return errorsx.Trace(err)
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
	data, err := Render(data)
	if err != nil {
		return errorsx.Trace(err)
	}
	if err := decodeFunc(data, dest); err != nil {
		return errorsx.Trace(err)
	}
	if err := Walk(dest, tag, walker); err != nil {
		return errorsx.Trace(err)
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
		return errorsx.Trace(err)
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
			return errorsx.Trace(err)
		}

		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() != reflect.Struct {
			continue
		}
		curAncestors := ancestors
		parts := strings.Split(field.Tag.Get(tag), ",")
		if !hasInlineTag(parts[1:]) {
			curAncestors = append(curAncestors, parts[0])
		}

		if err := walk(fieldType, fieldValue, curAncestors, tag, walker); err != nil {
			return errorsx.Trace(err)
		}
	}
	return nil
}

// Render Render
func Render(data []byte) ([]byte, error) {
	tpl, err := template.New("conf").Funcs(template.FuncMap{
		"kv": KV,
	}).Parse(string(data))
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, ""); err != nil {
		return nil, errorsx.Trace(err)
	}
	return buf.Bytes(), nil
}

// KV KV
func KV(key string) string {
	endpointsStr := os.Getenv("ETCD_ENDPOINTS")
	if len(endpointsStr) == 0 {
		endpointsStr = "127.0.0.1:2379"
	}
	endpoints := strings.Split(endpointsStr, ",")
	client, err := etcd.New(etcd.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 1,
	})
	if err != nil {
		log.WithError(err).
			Fatal("failed to New")
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*1)
	defer cancel()
	kvResp, err := client.KV.Get(ctx, key)
	if err != nil {
		log.WithError(err).
			Fatal("failed to Get")
	}
	if len(kvResp.Kvs) == 0 {
		log.WithField("key", key).
			Fatal("empty key")
	}
	return string(kvResp.Kvs[0].Value)
}

func hasInlineTag(parts []string) bool {
	return stringsContains(parts, "inline")
}

func stringsContains(parts []string, check string) bool {
	for _, part := range parts {
		if part == check {
			return true
		}
	}
	return false
}
