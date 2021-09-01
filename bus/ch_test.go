package bus

import (
	"fmt"
	"testing"
	"time"
)

type a interface {
	A()
}

type aa struct{}

func (a *aa) A() {

}

func TestBatchSubscribeChans(t *testing.T) {
	// ch1 := make(chan *aa)
	// ch2 := make(chan int)
	// BatchSubscribeChansContext(context.TODO(), map[interface{}]interface{}{
	// 	ch1: func(ctx context.Context, i a) {
	// 		fmt.Println("hello", i)
	// 	},
	// 	ch2: func(i string) {
	// 		fmt.Println("hello", i)
	// 	},
	// })
	// // a := &aa{}
	// close(ch1)
	// time.Sleep(time.Second * 1)
	// ch1 <- a
	ch1 := make(chan int)
	ch2 := make(chan string)
	ch3 := make(chan float64)
	ch4 := make(chan aa)
	ch5 := make(chan a)
	BroadcastChan(ch1, ch2, ch3, ch4, ch5)
	ch1 <- 1
	fmt.Println(<-ch2)
	fmt.Println(<-ch3)
	fmt.Println(<-ch4)
	fmt.Println(<-ch5)
	time.Sleep(time.Second)
}
