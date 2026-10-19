package main

import "net/http"

type LoadBalancer struct {
	backends  []*Backend
	scheduler Scheduler
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b := lb.scheduler.Pick(r)
	if b == nil {
		http.Error(w, "no healthy backends", http.StatusServiceUnavailable)
		return
	}

	b.ServeHTTP(w, r)
}