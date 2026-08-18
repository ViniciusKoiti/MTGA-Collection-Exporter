package observability

import (
	"fmt"
	"maps"
	"sort"
	"strings"
	"sync"
)

// Metrics is a bounded counter registry: label keys come from a closed
// allowlist, identifier-shaped values are refused, and the series
// budget is hard — cardinality can never explode silently.
type Metrics struct {
	AllowedLabels map[string]bool
	MaxSeries     int

	mu     sync.Mutex
	series map[string]int64
}

// Inc adds one to a series or refuses it; refusal never grows memory.
func (m *Metrics) Inc(name string, labels map[string]string) error {
	if name == "" {
		return fmt.Errorf("observability: metric name required")
	}
	keys := make([]string, 0, len(labels))
	for key, value := range labels {
		if !m.AllowedLabels[key] {
			return fmt.Errorf("observability: label %q outside the allowlist", key)
		}
		if identifierShaped(value) {
			return fmt.Errorf("observability: label %q carries an identifier", key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(name)
	for _, key := range keys {
		fmt.Fprintf(&b, ",%s=%s", key, labels[key])
	}
	seriesKey := b.String()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.series == nil {
		m.series = map[string]int64{}
	}
	if _, exists := m.series[seriesKey]; !exists && len(m.series) >= m.MaxSeries {
		return fmt.Errorf("observability: series budget %d spent", m.MaxSeries)
	}
	m.series[seriesKey]++
	return nil
}

// Snapshot copies every counter for scraping.
func (m *Metrics) Snapshot() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return maps.Clone(m.series)
}

// identifierShaped flags values that look like IDs or hashes: long
// hex/uuid-like strings are cardinality bombs and privacy leaks.
func identifierShaped(value string) bool {
	if len(value) < 16 {
		return false
	}
	for _, r := range strings.ToLower(strings.ReplaceAll(value, "-", "")) {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}
