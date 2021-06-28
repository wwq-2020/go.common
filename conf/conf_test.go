package conf

import (
	"testing"

	"github.com/wwq-2020/go.common/log"
)

type conf struct {
	A  int     `toml:"a"`
	B  *string `toml:"b"`
	XX struct {
		C int `toml:"c" `
	} `toml:"xx"`
}

func TestParse(t *testing.T) {
	c := &conf{}
	if err := ParseFile("./testdata/conf.toml", c); err != nil {
		log.Error(err)
	}
}
