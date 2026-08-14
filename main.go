package main

import (
	"context"
	"github.com/undndnwnkk/go-pulse/service"
	"log/slog"
	"runtime"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(100*time.Millisecond))
	defer cancel()

	dispatcher := service.NewDispatcher(3)
	addresses := []string{
		"192.168.1.1",
		"10.0.0.15",
		"172.16.0.1",
		"8.8.8.8",
		"1.1.1.1",
		"93.184.216.34",
		"140.82.121.4",
		"127.0.0.1",
	}

	res := dispatcher.Start(ctx, addresses)
	for val := range res {
		slog.Info("worker processed address",
			"address", val.Address,
			"success", val.Success,
		)
	}

	time.Sleep(1 * time.Second)
	slog.Info("active goroutines", "count", runtime.NumGoroutine())
}
