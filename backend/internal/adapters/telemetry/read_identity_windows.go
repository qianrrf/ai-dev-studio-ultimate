package telemetry

import (
	"io"
	"os"

	"golang.org/x/sys/windows"
)

func readIdentityFile(path string) ([]byte, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	// Concurrent identity recovery must be able to remove an empty legacy
	// record while another reader finishes reading that same generation.
	handle, err := windows.CreateFile(name, windows.GENERIC_READ,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	file := os.NewFile(uintptr(handle), path)
	defer func() { _ = file.Close() }()
	return io.ReadAll(file)
}
