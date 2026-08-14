package scanner

import (
	"context"
	"fmt"
	"github.com/undndnwnkk/go-pulse/internal/model"
	"math/rand/v2"
	"sync"
	"time"
)

type Dispatcher struct {
	NumWorkers int
}

func NewDispatcher(workers int) *Dispatcher {
	return &Dispatcher{NumWorkers: workers}
}

func (d *Dispatcher) Start(ctx context.Context, targets []string) <-chan model.Result {
	resCh := make(chan model.Result, d.NumWorkers)
	jobsCh := make(chan model.Job)

	go func() {
		defer close(jobsCh)
		for _, val := range targets {
			select {
			case <-ctx.Done():
				return
			case jobsCh <- model.Job{Address: val}:
			}
		}
	}()

	var wg sync.WaitGroup

	for i := 1; i <= d.NumWorkers; i++ {
		wg.Add(1)
		go worker(ctx, jobsCh, resCh, &wg)
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	return resCh
}

func worker(ctx context.Context, jobs <-chan model.Job, resCh chan<- model.Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for val := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
		}
		start := time.Now()

		time.Sleep(time.Millisecond * time.Duration(rand.IntN(80)+20))

		res := model.Result{
			Address: val.Address,
			Latency: time.Since(start),
		}

		if rand.Float32() < 0.7 {
			res.Err = fmt.Errorf("connect error")
			res.Success = false
		} else {
			res.Success = true
		}

		select {
		case <-ctx.Done():
			return
		case resCh <- res:
		}

	}
}
