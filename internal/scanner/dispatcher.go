package scanner

import (
	"bufio"
	"context"
	"io"
	"log/slog"
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

func (d *Dispatcher) Start(ctx context.Context, r io.Reader) <-chan model.Result {
	resCh := make(chan model.Result, d.NumWorkers)
	jobsCh := make(chan model.Job)

	go func() {
		defer close(jobsCh)

		var ticker *time.Ticker
		var tickerChan <-chan time.Time

		if d.RPS > 0 {
			ticker = time.NewTicker(time.Second / time.Duration(d.RPS))
			defer ticker.Stop()
			tickerChan = ticker.C
		}

		scanner := bufio.NewScanner(r)

		for scanner.Scan() {
			val := scanner.Text()

			if val == "" {
				continue
			}

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

		if err := scanner.Err(); err != nil {
			slog.Error("scanner error", "error", err)
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

	const maxRetries = 3
	const retryDelay = 100 * time.Millisecond

	for val := range jobs {
		if ctx.Err() != nil {
			return
		}

		start := time.Now()
		res := model.Result{
			Address: val.Address,
		}

		var conn net.Conn
		var err error

		for attempt := 1; attempt <= maxRetries; attempt++ {
			conn, err = dialer.DialContext(ctx, "tcp", val.Address)
			if err == nil {
				break
			}

			if attempt == maxRetries || ctx.Err() != nil {
				break
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(retryDelay):
			}
		}

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
