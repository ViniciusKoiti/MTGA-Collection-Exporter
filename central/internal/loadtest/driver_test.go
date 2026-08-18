package loadtest

import (
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// drive fans total operations over the given worker count and
// collects per-operation latencies; failures count, never panic.
func drive(t *testing.T, workers, total int,
	operation func(index int) error) ([]time.Duration, int64) {
	t.Helper()
	var (
		mu        sync.Mutex
		latencies []time.Duration
		failures  atomic.Int64
		next      atomic.Int64
		wg        sync.WaitGroup
	)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				index := int(next.Add(1)) - 1
				if index >= total {
					return
				}
				start := time.Now()
				err := operation(index)
				elapsed := time.Since(start)
				if err != nil {
					failures.Add(1)
					continue
				}
				mu.Lock()
				latencies = append(latencies, elapsed)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	return latencies, failures.Load()
}

// percentile answers the pth latency percentile of a sample.
func percentile(latencies []time.Duration, p float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), latencies...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	index := int(float64(len(sorted)-1) * p)
	return sorted[index]
}
