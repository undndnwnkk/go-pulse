package model

import (
	"sync/atomic"
)

type Stats struct {
	total       atomic.Int64
	successfull atomic.Int64
	failed      atomic.Int64
}

func (s *Stats) IncSuccessfull() {
	s.total.Add(1)
	s.successfull.Add(1)
}

func (s *Stats) IncFailed() {
	s.total.Add(1)
	s.failed.Add(1)
}

func (s *Stats) Total() int64 {
	return s.total.Load()
}

func (s *Stats) Successfull() int64 {
	return s.successfull.Load()
}

func (s *Stats) Failed() int64 {
	return s.failed.Load()
}
