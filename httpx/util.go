package httpx

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/wwq-2020/go.common/errors"
)

// DrainBody DrainBody
func DrainBody(src io.ReadCloser) ([]byte, io.ReadCloser, error) {
	defer src.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(src); err != nil {
		return nil, nil, errors.Trace(err)
	}
	return buf.Bytes(), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func toByteStr() {

}
