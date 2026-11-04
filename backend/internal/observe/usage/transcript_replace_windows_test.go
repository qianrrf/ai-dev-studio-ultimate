package usage

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func replaceOpenTranscript(source, target string) error {
	from, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	// ReplaceFileW, unlike legacy MoveFileEx, replaces an open destination
	// whose readers granted delete sharing, preserving their old generation.
	result, _, callErr := windows.NewLazySystemDLL("kernel32.dll").NewProc("ReplaceFileW").Call(
		uintptr(unsafe.Pointer(to)), uintptr(unsafe.Pointer(from)), 0, 0, 0, 0) // #nosec G103 -- documented ReplaceFileW UTF-16 path arguments.
	if result == 0 {
		return callErr
	}
	return nil
}
