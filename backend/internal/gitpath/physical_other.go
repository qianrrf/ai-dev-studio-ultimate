//go:build !windows

package gitpath

import "path/filepath"

func physicalPath(path string) (string, error) { return filepath.EvalSymlinks(path) }
