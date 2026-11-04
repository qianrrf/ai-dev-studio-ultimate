//go:build windows

package agent

import (
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func assertPrivateTestPath(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.IsDir() != (want == 0o700) {
		t.Fatalf("%s has the wrong file type", path)
	}
	handle, _, security, err := openCodexWindowsPath(path, info.IsDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = windows.CloseHandle(handle) }()
	if !security.ownerCurrent || !security.aclSafe {
		t.Errorf("%s does not have a private current-user ACL", path)
	}
}
