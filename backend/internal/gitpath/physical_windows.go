//go:build windows

package gitpath

import (
	"strings"

	"golang.org/x/sys/windows"
)

// Resolve junctions and volume aliases as well as symbolic links. Go's
// EvalSymlinks does not resolve junctions with the current Windows semantics.
func physicalPath(path string) (string, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	handle, err := windows.CreateFile(name, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", err
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	var size uint32 = 512
	for {
		buffer := make([]uint16, size)
		n, err := windows.GetFinalPathNameByHandle(handle, &buffer[0], size, 0)
		if err != nil {
			return "", err
		}
		if n >= size {
			size = n + 1
			continue
		}
		resolved := windows.UTF16ToString(buffer[:n])
		if rest, ok := strings.CutPrefix(resolved, `\\?\UNC\`); ok {
			return `\\` + rest, nil
		}
		return strings.TrimPrefix(resolved, `\\?\`), nil
	}
}
