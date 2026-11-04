//go:build windows

package ptyregistry

import (
	"errors"
	"math"

	"golang.org/x/sys/windows"
)

// defaultPidAlive probes PID liveness via OpenProcess. SUCCESS means alive
// (CloseHandle and return true). ERROR_ACCESS_DENIED mirrors EPERM: the
// process exists but cannot be queried, so treat as alive.
func defaultPidAlive(pid int) bool {
	if pid <= 0 || uint64(pid) > math.MaxUint32 {
		return false
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err == nil {
		_ = windows.CloseHandle(h)
		return true
	}
	return errors.Is(err, windows.ERROR_ACCESS_DENIED)
}
