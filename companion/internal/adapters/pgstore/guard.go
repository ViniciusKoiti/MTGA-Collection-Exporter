package pgstore

import (
	"fmt"
	"strings"
)

// Production-identity rejection (OpenSpec add-graph-workflow-harness,
// task 8.3): the harness store NEVER connects to anything that smells
// like production. The refusal happens on the DSN, before any dial.
var productionMarkers = []string{"prod", "production", "release", "live"}

// GuardDSN refuses production-marked DSNs; the harness may only ever
// open isolated development or CI databases.
func GuardDSN(dsn string) error {
	lower := strings.ToLower(dsn)
	for _, marker := range productionMarkers {
		if strings.Contains(lower, marker) {
			return fmt.Errorf(
				"pgstore: dsn carries the production marker %q — refused before dialing",
				marker)
		}
	}
	return nil
}
