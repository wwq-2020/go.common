package funcx

import (
	"fmt"
	"testing"
)

func TestCall(t *testing.T) {
	o := []interface{}{1, 2, 3, 4}
	BatchCall(func(a, b int) {
		fmt.Println("hello", a, b)
	}, SliceArgsSpliterFromSlice(2, o...))
}
