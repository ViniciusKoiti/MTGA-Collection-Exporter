//go:build windows

package windowsdetect

import "testing"

func TestSystemProcessSnapshotIsReadableWithoutOpeningProcesses(t *testing.T) {
	names, err := (systemProcesses{}).Names(t.Context())
	if err != nil {
		t.Fatalf("enumerate process snapshot: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("the process snapshot must contain at least the test process")
	}
}
