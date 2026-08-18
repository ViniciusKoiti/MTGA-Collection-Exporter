package windowsdetect

import (
	"strings"
)

func containsProcess(names []string, target string) bool {
	for _, name := range names {
		if separator := strings.LastIndexAny(name, `/\\`); separator >= 0 {
			name = name[separator+1:]
		}
		if strings.EqualFold(name, target) {
			return true
		}
	}
	return false
}
