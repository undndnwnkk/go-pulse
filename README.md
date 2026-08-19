# go-pulse

A concurrent TCP scanner written in Go: checks the availability of a list of addresses (`host:port`) using a worker-goroutine pool, request-rate limiting (RPS), and automatic retries on failed connections.

## Features

- **Worker pool** — parallel address checks via a configurable number of goroutines (`NumWorkers`)
- **Rate limiting** — caps requests per second using a `time.Ticker`
- **Retry logic** — up to 3 connection attempts with a 100ms delay between them
- **Live stats** — atomic total/success/failed counters, updated in the console in real time
- **Graceful shutdown** — clean stop on Ctrl+C (SIGINT/SIGTERM) via `context`
- **Built-in pprof** — `localhost:6060` endpoint for live CPU/memory profiling
- **File-based targets** — one address per line, read from `target.txt`

## Project structure

```
go-pulse/
├── cmd/
│   └── pulse/
│       └── main.go          # entry point, reads target.txt, prints stats
├── internal/
│   ├── model/
│   │   ├── model.go          # Job, Result, error types
│   │   └── stats.go          # atomic stats counters (Stats)
│   └── scanner/
│       ├── dispatcher.go     # Dispatcher: job distribution, workers
│       └── worker_test.go    # worker benchmark
├── target.txt                 # list of addresses to scan
└── go.mod
```

## How it works

1. `Dispatcher` reads `target.txt` line by line via `bufio.Scanner` and pushes addresses into the `jobsCh` channel, throttling itself with a ticker if RPS is set
2. `NumWorkers` worker goroutines pull jobs off the channel and attempt a TCP connection (`net.Dialer` with a 3s timeout), retrying up to 3 times on failure
3. Each result (success/failure, latency, error type) is sent to `resCh`, read back in `main.go` and printed as live progress
4. On completion or Ctrl+C, final stats are printed along with the number of active goroutines

## Installation and usage

```bash
git clone https://github.com/undndnwnkk/go-pulse.git
cd go-pulse
go run ./cmd/pulse
```

A `target.txt` file must exist in the project root before running — one `host:port` address per line.

pprof is available at `http://localhost:6060/debug/pprof/` while the app is running.

## Example console output

```
[Progress: 187] | Success: 42 | Failed: 145 | Current: metrics.monitor.io:5000

Scan completed successfully!
[FINAL STATS] Total: 700 | Success: 163 | Failed: 537
```

If interrupted mid-scan with Ctrl+C:

```
[Progress: 412] | Success: 98 | Failed: 314 | Current: api.github.com:8080

Scan interrupted by user (Ctrl+C). Finalizing...
[FINAL STATS] Total: 412 | Success: 98 | Failed: 314
```

The progress line is rewritten in place (`\r\033[K`) as each result comes in, so only the latest state is visible until the scan finishes.

## Benchmarks

Tested locally on:

```
OS:   Windows
Arch: amd64
CPU:  AMD Ryzen 5 7535HS with Radeon Graphics
```

```
$ go test -bench=. -benchmem ./internal/scanner
goos: windows
goarch: amd64
pkg: github.com/undndnwnkk/go-pulse/internal/scanner
cpu: AMD Ryzen 5 7535HS with Radeon Graphics
BenchmarkWorker-12      2084      561655 ns/op      2784 B/op      43 allocs/op
PASS
ok      github.com/undndnwnkk/go-pulse/internal/scanner  1.911s
```

The benchmark runs a single worker against a local TCP listener — ~560µs per connection, 43 allocations / ~2.7KB per operation.

### Memory profiling (pprof)

```
$ go tool pprof http://localhost:6060/debug/pprof/heap
(pprof) top
Showing nodes accounting for 3080.40kB, 100% of 3080.40kB total
Showing top 10 nodes out of 24
      flat  flat%   sum%        cum   cum%
    1026kB 33.31% 33.31%     1026kB 33.31%  runtime.mallocgc
 1024.02kB 33.24% 66.55%  1024.02kB 33.24%  bufio.(*Scanner).Text (inline)
  515.19kB 16.72% 83.28%   515.19kB 16.72%  context.(*cancelCtx).propagateCancel
  515.19kB 16.72%   100%   515.19kB 16.72%  scanner.(*Dispatcher).Start
```

Main memory consumers during an active run:
- `bufio.(*Scanner).Text()` — allocates a new string for every line read from the target file
- `context.WithDeadline` / `propagateCancel` — context-tree overhead on every `DialContext` call

Potential optimization points: reuse a read buffer instead of `Scanner.Text()`, or reduce how often child contexts are created per connection.

## Requirements

- Go 1.26+