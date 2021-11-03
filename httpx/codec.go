package httpx

import (
	"encoding/json"

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

type codeMsgJSONCodec struct {
	expectedCode int
}

// CodeMsgJSONCodec CodeMsgJSONCodec
func CodeMsgJSONCodec(expectedCode int) Codec {
	return &codeMsgJSONCodec{
		expectedCode: expectedCode,
	}
}

type respObj struct {
	Code int         `json:"code"`
	Msg  int         `json:"msg"`
	Data interface{} `json:"data"`
}

func (c *codeMsgJSONCodec) Encode(obj interface{}) ([]byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, errorsx.Trace(err)
	}
	return data, nil
}

func (c *codeMsgJSONCodec) Decode(data []byte, obj interface{}) error {
	respObj := &respObj{
		Data: obj,
	}
	err := json.Unmarshal(data, respObj)
	if err != nil {
		return errorsx.Trace(err)
	}
	if respObj.Code != c.expectedCode {
		return errorsx.New("got unexpected code").
			WithField("expectedCode", c.expectedCode).
			WithField("gotCode", respObj.Code)
	}
	return nil
}
