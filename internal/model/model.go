package model

import (
	"time"
)

const (
	ErrTimeout           = "Timeout"
	ErrConnectionRefused = "Connection refused"
)

type ErrorType string

type Job struct {
	Address string
}

type Result struct {
	Address string
	Success bool
	Latency time.Duration
	ErrType ErrorType
	Err     error
}
