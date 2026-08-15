//go:build windows

package windowsdetect

import (
	"context"
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

type systemProcesses struct{}

func (systemProcesses) Names(ctx context.Context) ([]string, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer func() { _ = windows.CloseHandle(snapshot) }()
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	if err := windows.Process32First(snapshot, &entry); err != nil {
		return nil, err
	}
	names := make([]string, 0, 64)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		names = append(names, windows.UTF16ToString(entry.ExeFile[:]))
		err = windows.Process32Next(snapshot, &entry)
		if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
			return names, nil
		}
		if err != nil {
			return nil, err
		}
	}
}
