package scanner

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/undndnwnkk/go-pulse/internal/model"
)

type Dispatcher struct {
	NumWorkers int
	Stats      *model.Stats
	RPS        int
}

func NewDispatcher(workers int, stats *model.Stats, rps int) *Dispatcher {
	return &Dispatcher{NumWorkers: workers, Stats: stats, RPS: rps}
}

func (d *Dispatcher) Start(ctx context.Context, targets []string) <-chan model.Result {
	resCh := make(chan model.Result, d.NumWorkers)
	jobsCh := make(chan model.Job)

	go func() {
		defer close(jobsCh)

		var ticker *time.Ticker
		var tickerChan <-chan time.Time

		if d.RPS > 0 {
			ticker = time.NewTicker(time.Second / time.Duration(d.RPS))
			defer ticker.Stop() // defer внутри горутины, где создан тикер!
			tickerChan = ticker.C
		}

		for _, val := range targets {
			if tickerChan != nil {
				select {
				case <-ctx.Done():
					return
				case <-tickerChan:
				}
			}

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
		if ctx.Err() != nil {
			return
		}

		start := time.Now()
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
