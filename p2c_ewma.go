package main

import (
	"math/rand/v2"
	"net/http"
	"sync"
)

type P2C struct {
	backends []*Backend
	rnd      *rand.Rand
	mu       sync.Mutex
}

func (p *P2C) Pick(r *http.Request) *Backend {
	healthy := make([]*Backend, 0, len(p.backends))
	for _, b := range p.backends {
		if b.Alive.Load() {
			healthy = append(healthy, b)
		}
	}

	n := len(healthy)
	if n == 0 {
		return nil
	}
	if n == 1 {
		return healthy[0]
	}

	p.mu.Lock()
	i := p.rnd.Intn(n)
	j := p.rnd.Intn(n)
	p.mu.Unlock()

	for j == i {
		p.mu.Lock()
		j = p.rnd.Intn(n)
		p.mu.Unlock()
	}

	a, b := healthy[i], healthy[j]
	if p.score(a) <= p.score(b) {
		return a
	}
	return b
}

func (p *P2C) score(b *Backend) float64 {
	// Active connections plus latency penalty.
	// Tune the weight: 10 means 100ms adds 1.0 to the score.
	return float64(b.Active.Load()) + b.EWMA().Seconds()*10
}