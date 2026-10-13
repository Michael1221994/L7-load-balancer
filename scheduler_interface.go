package main

import "net/http"

type Scheduler interface {
	Pick(r *http.Request) *Backend
}