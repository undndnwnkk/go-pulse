package main

import (
	"context"
	"fmt"
	"github.com/undndnwnkk/go-pulse/internal/model"
	"github.com/undndnwnkk/go-pulse/internal/scanner"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Duration(100*time.Millisecond))
	// defer cancel()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	stats := model.Stats{}

	dispatcher := scanner.NewDispatcher(3, &stats, 2)

	file, err := os.Open("target.txt")
	if err != nil {
		slog.Error("error reading file", "error", err)
		return
	}
	defer file.Close()

	res := dispatcher.Start(ctx, file)
	for val := range res {
		fmt.Printf("\r\033[K[Progress: %d] | Success: %d | Failed: %d | Current: %s",
			stats.Total(),
			stats.Successfull(),
			stats.Failed(),
			val.Address,
		)
	}
	fmt.Println()

	time.Sleep(1 * time.Second)
	slog.Info("active goroutines", "count", runtime.NumGoroutine())

	<-ctx.Done()

	fmt.Println("\nFINAL STATS")
	fmt.Printf("\r\033[K[Progress: %d] | Success: %d | Failed: %d\n",
		stats.Total(),
		stats.Successfull(),
		stats.Failed(),
	)

	fmt.Println("Scan interrupting by user. Finalizing...")
}
