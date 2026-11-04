//go:build !windows

package usage

import "os"

func replaceOpenTranscript(source, target string) error { return os.Rename(source, target) }
