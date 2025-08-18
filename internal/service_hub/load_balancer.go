package service_hub

import (
	"math/rand"
	"sync/atomic"
)

type LoadBalancer interface {
	Take([]string) string
}

type RoundRobin struct {
	acc int64
}

func (b *RoundRobin) Take(endpoints []string) string {
	if len(endpoints) == 0 {
		return ""
	}

	n := atomic.AddInt64(&b.acc, 1) // Take needs to support concurrent calls
	index := int(n % int64(len(endpoints)))
	
	return endpoints[index]
}

type RandomSelect struct {
}

func (b *RandomSelect) Take(endpoints []string) string {
	if len(endpoints) == 0 {
		return ""
	}

	index := rand.Intn(len(endpoints)) // random selection

	return endpoints[index]
}
