//go:build !windows

package usage

import "os"

func openTranscriptFile(path string) (*os.File, error) { return os.Open(path) }
