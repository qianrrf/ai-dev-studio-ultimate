//go:build !windows

package testutil

import (
	"os"
	"testing"
)

// AssertPrivatePath checks native permissions, rather than Windows' synthetic mode bits.
func AssertPrivatePath(t testing.TB, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != want || info.IsDir() != (want == 0o700) {
		t.Fatalf("%s: mode %v, want private mode %o", path, info.Mode(), want)
	}
}
