package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/undndnwnkk/go-pulse/internal/model"
	"github.com/undndnwnkk/go-pulse/internal/scanner"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			slog.Error("pprof server error", "error", err)
		}
	}()

	stats := model.Stats{}
	dispatcher := scanner.NewDispatcher(100, &stats, 500)

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

	if ctx.Err() != nil {
		fmt.Println("\nScan interrupted by user (Ctrl+C). Finalizing...")
	} else {
		fmt.Println("\nScan completed successfully!")
	}

	fmt.Printf("[FINAL STATS] Total: %d | Success: %d | Failed: %d\n",
		stats.Total(),
		stats.Successfull(),
		stats.Failed(),
	)

	time.Sleep(100 * time.Millisecond)
	slog.Info("active goroutines", "count", runtime.NumGoroutine())
}
