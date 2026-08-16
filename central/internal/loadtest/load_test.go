package loadtest

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestStageOneLoadAtCIScale drives the manifest and telemetry paths
// at CI scale (task 8.1): 2000 manifest reads over 16 workers and 400
// unique telemetry batches over 8 workers — a compressed burst well
// above the 10k-DAU steady rates derived in docs/load-stage1.md.
// Evidence: p50/p95/p99 per path, zero errors, the pool never grows
// past its budget, and goroutine/GC counters as the CPU proxy.
func TestStageOneLoadAtCIScale(t *testing.T) {
	base, pool := startStack(t)
	client := &http.Client{Timeout: 10 * time.Second}

	manifestLat, manifestErrs := drive(t, 16, 2000, func(int) error {
		resp, err := client.Get(base + "/v1/manifest/current")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status %d", resp.StatusCode)
		}
		return nil
	})
	telemetryLat, telemetryErrs := drive(t, 8, 400, func(index int) error {
		body := fmt.Sprintf(`{"batch_id":"load-%d","sequence":%d,`+
			`"events":[{"name":"scan_completed","attrs":{"result":"ok"}}]}`,
			index, index+1)
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

	if manifestErrs != 0 || telemetryErrs != 0 {
		t.Fatalf("the load must complete error-free: manifest=%d telemetry=%d",
			manifestErrs, telemetryErrs)
	}
	manifestP95 := percentile(manifestLat, 0.95)
	telemetryP95 := percentile(telemetryLat, 0.95)
	if manifestP95 > 100*time.Millisecond {
		t.Fatalf("manifest p95 blew its budget: %s", manifestP95)
	}
	if telemetryP95 > 250*time.Millisecond {
		t.Fatalf("telemetry p95 blew its budget: %s", telemetryP95)
	}
	stat := pool.Stat()
	if stat.TotalConns() > 8 {
		t.Fatalf("the pool grew past its budget: %d", stat.TotalConns())
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	t.Logf("stage1 evidence: manifest p50=%s p95=%s p99=%s (n=%d) | "+
		"telemetry p50=%s p95=%s p99=%s (n=%d) | pool total=%d idle=%d | "+
		"goroutines=%d gc_pause_total=%s",
		percentile(manifestLat, 0.50), manifestP95,
		percentile(manifestLat, 0.99), len(manifestLat),
		percentile(telemetryLat, 0.50), telemetryP95,
		percentile(telemetryLat, 0.99), len(telemetryLat),
		stat.TotalConns(), stat.IdleConns(), runtime.NumGoroutine(),
		time.Duration(mem.PauseTotalNs))
}
