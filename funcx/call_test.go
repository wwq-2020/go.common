package funcx

import (
	"fmt"
	"testing"
)

func TestCall(t *testing.T) {
	BatchCall(func(a, b int) {
		fmt.Println("hello", a, b)
	}, [][]interface{}{{1, 2}, {3, 4}})
}
