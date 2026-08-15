package windowsdetect

import (
	"path/filepath"
	"strings"
)

func containsProcess(names []string, target string) bool {
	for _, name := range names {
		if strings.EqualFold(filepath.Base(name), target) {
			return true
		}
	}
	return false
}
