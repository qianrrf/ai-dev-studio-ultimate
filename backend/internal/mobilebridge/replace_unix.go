//go:build !windows

package mobilebridge

import "os"

func replaceStateFile(source, target string) error { return os.Rename(source, target) }
