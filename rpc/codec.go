package rpc

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/wwq-2020/go.common/errorsx"
	"github.com/wwq-2020/go.common/httpx"
	"github.com/wwq-2020/go.common/log"
)

// Codec Codec
type Codec = httpx.Codec

type jsonCodec struct {
	r io.Reader
}

// JSONCodec JSONCodec
var JSONCodec = httpx.JSONCodec

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
	log.WithField("reqData", string(reqData)).
		InfoContext(c.ctx, "got reqData")
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
	log.WithField("respData", string(respData)).
		InfoContext(c.ctx, "send respData")
	if _, err := c.w.Write(respData); err != nil {
		return errorsx.Trace(err)
	}
	return nil
}
