package scanner

import (
	"context"
	"github.com/undndnwnkk/go-pulse/internal/model"
	"net"
	"sync"
	"time"
)

type Dispatcher struct {
	NumWorkers int
	Stats      *model.Stats
}

func NewDispatcher(workers int, stats *model.Stats) *Dispatcher {
	return &Dispatcher{NumWorkers: workers, Stats: stats}
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
		go worker(ctx, jobsCh, resCh, &wg, d.Stats)
	}

	go func() {
		wg.Wait()
		close(resCh)
	}()

	return resCh
}

func worker(ctx context.Context, jobs <-chan model.Job, resCh chan<- model.Result, wg *sync.WaitGroup, stats *model.Stats) {
	defer wg.Done()
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	for val := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
		}
		start := time.Now()

		// time.Sleep(time.Millisecond * time.Duration(rand.IntN(80)+20))
		res := model.Result{
			Address: val.Address,
		}
		conn, err := dialer.DialContext(ctx, "tcp", val.Address)
		res.Latency = time.Since(start)

		if err != nil {
			res.Success = false
			res.Err = err
			stats.IncFailed()

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				res.ErrType = model.ErrTimeout
			} else {
				res.ErrType = model.ErrConnectionRefused
			}
		} else {
			res.Success = true
			stats.IncSuccessfull()
			conn.Close()
		}

		select {
		case <-ctx.Done():
			return
		case resCh <- res:
		}

	}
}
