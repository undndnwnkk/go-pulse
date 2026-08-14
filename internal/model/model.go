package model

import (
	"time"
)

type Job struct {
	Address string
}

type Result struct {
	Address string
	Success bool
	Latency time.Duration
	Err     error
}
