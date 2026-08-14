package main

import (
	"context"
	"github.com/undndnwnkk/go-pulse/internal/scanner"
	"log/slog"
	"runtime"
	"time"
)

func main() {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(100*time.Millisecond))
	// defer cancel()
	ctx := context.Background()

	dispatcher := scanner.NewDispatcher(3)
	addresses := []string{
		"google.com:80",
		"google.com:12345",
		"8.8.8:53",
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
