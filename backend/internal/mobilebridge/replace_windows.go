package mobilebridge

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

func replaceStateFile(source, target string) error {
	// A concurrent reader or virus scanner can briefly hold the destination
	// without delete sharing. Keep the old file intact until replacement works.
	deadline := time.Now().Add(250 * time.Millisecond)
	for {
		err := os.Rename(source, target)
		if (!errors.Is(err, windows.ERROR_SHARING_VIOLATION) && !errors.Is(err, windows.ERROR_ACCESS_DENIED)) || !time.Now().Before(deadline) {
			return err
		}
		time.Sleep(5 * time.Millisecond)
	}
}
