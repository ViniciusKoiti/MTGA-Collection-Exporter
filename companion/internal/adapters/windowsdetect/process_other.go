//go:build !windows

package windowsdetect

import "context"

type systemProcesses struct{}

func (systemProcesses) Names(context.Context) ([]string, error) {
	return nil, ErrUnsupported
}
