package scanner

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/undndnwnkk/go-pulse/internal/model"
)

func BenchmarkWorker(b *testing.B) {
	// Создаем локальный TCP-сервер
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	ctx := context.Background()
	addr := ln.Addr().String()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		jobsCh := make(chan model.Job, 1)
		resCh := make(chan model.Result, 1)
		stats := &model.Stats{}
		var wg sync.WaitGroup

		jobsCh <- model.Job{Address: addr}
		close(jobsCh)

		wg.Add(1)
		go worker(ctx, jobsCh, resCh, &wg, stats)
		wg.Wait()
	}
}
