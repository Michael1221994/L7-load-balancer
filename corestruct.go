package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL   *url.URL
	Proxy *httputil.ReverseProxy

	Alive    atomic.Bool
	Active   atomic.Int64
	Total    atomic.Uint64
	Failures atomic.Uint64

	mu   sync.Mutex
	ewma time.Duration
} 

func (b *Backend) observeLatency(d time.Duration) {
	const alpha = 0.2

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.ewma == 0 {
		b.ewma = d
	} else {
		b.ewma = time.Duration(
			float64(b.ewma)*(1-alpha) + float64(d)*alpha,
		)
	}
}

func (b *Backend) EWMA() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ewma
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.Active.Add(1)
	defer b.Active.Add(-1)

	start := time.Now()
	defer func() {
		b.observeLatency(time.Since(start))
	}()

	b.Proxy.ServeHTTP(w, r)
}
