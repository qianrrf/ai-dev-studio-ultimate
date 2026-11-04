//go:build windows

package browserstream

import (
	"net"
	"path/filepath"
	"regexp"

	"github.com/Microsoft/go-winio"
)

var unsafePipeChars = regexp.MustCompile(`[^a-zA-Z0-9\-]`)

func Listen(runFilePath string) (net.Listener, string, error) {
	suffix := unsafePipeChars.ReplaceAllString(filepath.Base(filepath.Dir(runFilePath)), "-")
	if suffix == "" || suffix == ".ai-dev-studio-ultimate" || suffix == "." {
		suffix = "default"
	}
	name := `\\.\pipe\ai-dev-studio-ultimate-browser-stream-` + suffix
	ln, err := winio.ListenPipe(name, &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;SY)(A;;GA;;;OW)"})
	if err != nil {
		return nil, "", err
	}
	return ln, name, nil
}
