package funcx

import (
	"context"
	"fmt"
	"testing"
)

func TestCall(t *testing.T) {
	o := []interface{}{1, 2, 3, 4}
	ctx := context.TODO()
	BatchCallContext(ctx, func(ctx context.Context, a, b int) {
		fmt.Println("hello", a, b)
	}, SliceArgsSpliterFromSlice(2, o...))
}
