package env

import (
	"fmt"
	"testing"
)

func TestT(t *testing.T) {
	fmt.Println(Env(2000).Number())
	fmt.Println(IsValid(2000))
}
