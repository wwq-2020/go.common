package rpc

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/wwq-2020/go.common/errorsx"
)

// Codec Codec
type Codec interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
}

type jsonCodec struct{}

// JSONCodec JSONCodec
func JSONCodec() Codec {
	return &jsonCodec{}
}

func (c *jsonCodec) Encode(obj interface{}) ([]byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return data, nil
}

func (c *jsonCodec) Decode(data []byte, obj interface{}) error {
	err := json.Unmarshal(data, obj)
	if err != nil {
		return errorsx.Trace(err)
	}
	return nil
}

type respObj struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// ServerCodec ServerCodec
type ServerCodec interface {
	Encode(obj interface{}) error
	Decode(obj interface{}) error
}

type serverCodec struct {
	ctx   context.Context
	r     io.Reader
	w     io.Writer
	codec Codec
}

func serverCodecFactory(ctx context.Context, r io.Reader, w io.Writer, codec Codec) ServerCodec {
	return &serverCodec{
		ctx:   ctx,
		r:     r,
		w:     w,
		codec: codec,
	}
}

func (c *serverCodec) Decode(obj interface{}) error {
	reqData, err := ioutil.ReadAll(c.r)
	if err != nil {
		return errorsx.TraceWithCode(err, http.StatusBadRequest)
	}
	if err := c.codec.Decode(reqData, obj); err != nil {
		return errorsx.ReplaceCode(err, http.StatusBadRequest)
	}
	return nil
}

func (c *serverCodec) Encode(obj interface{}) error {
	respData, err := c.codec.Encode(obj)
	if err != nil {
		return errorsx.Trace(err)
	}
	if _, err := c.w.Write(respData); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}
