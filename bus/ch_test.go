package bus

import (
	"context"
	"fmt"
	"testing"
)

func TestBatchSubscribeChans(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	BatchSubscribeChans(context.TODO(), MapChanSpliterFromObj(map[interface{}]interface{}{
		ch1: func(i int) {
			fmt.Println("hello", i)
		},
		ch2: func(i string) {
			fmt.Println("hello", i)
		},
	}))
	ch1 <- 1
}
