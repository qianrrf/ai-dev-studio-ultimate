//go:build windows

package process

import (
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func configureHidden(cmd *exec.Cmd) {
	// Git parses -C before core.longpaths. Let Windows enter the directory first.
	name := filepath.Base(cmd.Path)
	if (strings.EqualFold(name, "git.exe") || strings.EqualFold(name, "git")) &&
		len(cmd.Args) >= 3 && cmd.Args[1] == "-C" && filepath.IsAbs(cmd.Args[2]) && len(cmd.Args[2]) >= 248 {
		cmd.Dir = cmd.Args[2]
		cmd.Args = append(cmd.Args[:1], cmd.Args[3:]...)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NO_WINDOW,
		HideWindow:    true,
	}
}
