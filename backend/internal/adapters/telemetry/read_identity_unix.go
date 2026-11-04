//go:build !windows

package telemetry

import "os"

func readIdentityFile(path string) ([]byte, error) { return os.ReadFile(path) }
