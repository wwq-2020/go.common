package bus

import (
	"context"
	"fmt"
	"testing"
)

type a interface {
	A()
}

type aa struct{}

func (a *aa) A() {

}

func TestBatchSubscribeChans(t *testing.T) {
	ch1 := make(chan *aa)
	ch2 := make(chan int)
	BatchSubscribeChansContext(context.TODO(), MapChanSpliterFromObj(map[interface{}]interface{}{
		ch1: func(i a) {
			fmt.Println("hello", i)
		},
		ch2: func(i string) {
			fmt.Println("hello", i)
		},
	}))
	a := &aa{}
	ch1 <- a
}
