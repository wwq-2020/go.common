package workerpool

import (
	"context"
	"fmt"
	"testing"

	"github.com/wwq-2020/go.common/app"
)

func TestWorkerPool(t *testing.T) {
	wp := New(1)
	wp.Start()
	app.AddShutdownHook(wp.Stop)
	wp.Do(func(ctx context.Context) {
		fmt.Println("hello1")
	})

	wp.DoAsync(func(ctx context.Context) {
		fmt.Println("hello1")
	})
}
