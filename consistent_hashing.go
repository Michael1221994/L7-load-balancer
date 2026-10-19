package main

import (
	"fmt"
	"hash/crc32"
	"net/http"
	"sort"
	"sync"
)

type ConsistentHash struct {
	mu       sync.RWMutex
	ring     map[uint32]*Backend
	sorted   []uint32
	replicas int
}// consistent hashing implementation for load balancing, with support for session affinity via cookies.

func NewConsistentHash(backends []*Backend, replicas int) *ConsistentHash {
	c := &ConsistentHash{
		ring:     make(map[uint32]*Backend),
		replicas: replicas,
	}

	for _, b := range backends {
		c.add(b)
	}

	return c
}

func (c *ConsistentHash) add(b *Backend) {
	for i := 0; i < c.replicas; i++ {
		key := fmt.Sprintf("%s#%d", b.URL.String(), i)
		h := crc32.ChecksumIEEE([]byte(key))

		c.ring[h] = b
		c.sorted = append(c.sorted, h)
	}

	sort.Slice(c.sorted, func(i, j int) bool {
		return c.sorted[i] < c.sorted[j]
	})
}

func (c *ConsistentHash) Pick(r *http.Request) *Backend {
	key := r.RemoteAddr

	if cookie, err := r.Cookie("lb_session"); err == nil {
		key = cookie.Value
	}

	h := crc32.ChecksumIEEE([]byte(key))

	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.sorted) == 0 {
		return nil
	}

	idx := sort.Search(len(c.sorted), func(i int) bool {
		return c.sorted[i] >= h
	})
	if idx == len(c.sorted) {
		idx = 0
	}

	// Walk the ring until we find a healthy backend.
	for i := 0; i < len(c.sorted); i++ {
		b := c.ring[c.sorted[(idx+i)%len(c.sorted)]]
		if b.Alive.Load() {
			return b
		}
	}

	return nil
}