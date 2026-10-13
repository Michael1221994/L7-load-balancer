package main

import (
	"math"
	"net/http"
)

type LeastConnections struct {
	backends []*Backend
}

func (s *LeastConnections) Pick(r *http.Request) *Backend {
	var best *Backend
	var bestActive int64 = math.MaxInt64

	for _, b := range s.backends {
		if !b.Alive.Load() {
			continue
		}

		active := b.Active.Load()
		if best == nil || active < bestActive {
			best = b
			bestActive = active
		}
	}

	return best
}