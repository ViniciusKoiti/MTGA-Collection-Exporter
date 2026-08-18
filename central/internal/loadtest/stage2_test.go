package loadtest

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestStageTwoBurstAtCIScale runs only after Stage 1 is green (task
// 8.2): sustain a burst ABOVE 700 events/s through the real telemetry
// path and prove at least 30 percent of the PostgreSQL connection
// budget stays free the whole time. CI-scale caveats are documented
// in docs/load-stage1.md; the staging-scale rerun stays with 8.5.
func TestStageTwoBurstAtCIScale(t *testing.T) {
	// 20 connections model the Capacity budget of one worker plus API
	// replicas; the burst must leave 30% of it untouched.
	base, pool := startStackWith(t, 20)
	client := &http.Client{Timeout: 10 * time.Second}
	const (
		eventsPerBatch = 4
		totalBatches   = 2500 // 10k events total
		workers        = 12   // sized to leave the pool its headroom
	)
	// Sample pool pressure while the burst runs; the maximum acquired
	// count is the headroom evidence.
	var maxAcquired atomic.Int32
	stopSampling := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stopSampling:
				return
			case <-ticker.C:
				if acquired := pool.Stat().AcquiredConns(); acquired > maxAcquired.Load() {
					maxAcquired.Store(acquired)
				}
			}
		}
	}()

	events := `[{"name":"scan_completed","attrs":{"result":"ok"}},` +
		`{"name":"scan_completed","attrs":{"result":"ok"}},` +
		`{"name":"scan_completed","attrs":{"result":"ok"}},` +
		`{"name":"scan_completed","attrs":{"result":"ok"}}]`
	start := time.Now()
	_, failures := drive(t, workers, totalBatches, func(index int) error {
		body := fmt.Sprintf(`{"batch_id":"burst-%d","sequence":%d,"events":%s}`,
			index, index+1_000_000, events)
		req, err := http.NewRequest(http.MethodPost, base+"/v1/telemetry",
			strings.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer load-token")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode != http.StatusAccepted {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	})
	elapsed := time.Since(start)
	close(stopSampling)

	if failures != 0 {
		t.Fatalf("the burst must complete error-free: %d failures", failures)
	}
	rate := float64(totalBatches*eventsPerBatch) / elapsed.Seconds()
	if rate < 700 {
		t.Fatalf("the burst must sustain 700 events/s, got %.0f", rate)
	}
	budget := pool.Stat().MaxConns()
	acquired := maxAcquired.Load()
	if float64(acquired) > 0.7*float64(budget) {
		t.Fatalf("PostgreSQL headroom below 30%%: peak %d of %d connections",
			acquired, budget)
	}
	t.Logf("stage2 evidence: %.0f events/s over %s, peak %d/%d connections "+
		"(headroom %.0f%%)", rate, elapsed.Round(time.Millisecond),
		acquired, budget, 100*(1-float64(acquired)/float64(budget)))
}
