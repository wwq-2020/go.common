package rpc

import (
	"math/rand"
	"sync"
	"time"

	"github.com/wwq-2020/go.common/errorsx"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Balancer Balancer
type Balancer interface {
	Add(string)
	Del(string)
	Pick() (string, error)
}

type randomBalancer struct {
	endpoints []string
	m         sync.Mutex
}

// NewRandomBalancer NewRandomBalancer
func NewRandomBalancer() Balancer {
	return &randomBalancer{}
}

func (b *randomBalancer) Add(endpoint string) {
	b.m.Lock()
	defer b.m.Unlock()
	b.endpoints = append(b.endpoints, endpoint)
	rand.Shuffle(len(b.endpoints), func(i, j int) {
		b.endpoints[j], b.endpoints[i] = b.endpoints[i], b.endpoints[j]
	})
}

func (b *randomBalancer) Del(endpoint string) {
	b.m.Lock()
	defer b.m.Unlock()
	for i, each := range b.endpoints {
		if each == endpoint {
			b.endpoints = append(b.endpoints[:i], b.endpoints[i+1:]...)
			return
		}
	}
	rand.Shuffle(len(b.endpoints), func(i, j int) {
		b.endpoints[j], b.endpoints[i] = b.endpoints[i], b.endpoints[j]
	})
}

func (b *randomBalancer) Pick() (string, error) {
	b.m.Lock()
	defer b.m.Unlock()
	if len(b.endpoints) == 0 {
		return "", errorsx.New("no endpoint")
	}
	n := rand.Intn(len(b.endpoints))
	return b.endpoints[n], nil
}
