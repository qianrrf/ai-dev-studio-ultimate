//go:build windows

package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestCommandConstructorsPreserveGitDirectoryArguments(t *testing.T) {
	dir := filepath.Join(t.TempDir(), strings.Repeat("segment", 24), strings.Repeat("nested", 20))
	for _, name := range []string{"git", "git.exe", "other.exe"} {
		for _, command := range []*exec.Cmd{
			Command(name, "-C", dir, "status", "--porcelain"),
			CommandContext(t.Context(), name, "-C", dir, "status", "--porcelain"),
		} {
			wantArgs := []string{name, "-C", dir, "status", "--porcelain"}
			if command.Dir != "" || !reflect.DeepEqual(command.Args, wantArgs) {
				t.Fatalf("%s: dir=%q args=%q", name, command.Dir, command.Args)
			}
		}
	}
	if cmd := Command("git", "-C", t.TempDir(), "status"); cmd.Dir != "" || cmd.Args[1] != "-C" {
		t.Fatalf("short directory changed: %v", cmd)
	}
}

func TestCommandContextHidesConsoleWindow(t *testing.T) {
	cmd := CommandContext(context.Background(), "git", "--version")
	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr = nil, want hidden Windows process attributes")
	}
	if got := cmd.SysProcAttr.CreationFlags; got&windows.CREATE_NO_WINDOW == 0 {
		t.Fatalf("CreationFlags = %#x, want CREATE_NO_WINDOW", got)
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatal("HideWindow = false, want true")
	}
}

func TestCommandContextRunsBatchFileWithoutConsole(t *testing.T) {
	script := filepath.Join(t.TempDir(), "console-probe.cmd")
	contents := "@echo off\r\n\"%~1\" -test.run=TestHiddenConsoleHelper\r\n"
	if err := os.WriteFile(script, []byte(contents), 0o600); err != nil {
		t.Fatalf("write batch probe: %v", err)
	}

	cmd := CommandContext(t.Context(), script, os.Args[0])
	cmd.Env = append(os.Environ(), "AO_HIDDEN_CONSOLE_HELPER=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run batch probe: %v\n%s", err, output)
	}
}

func TestHiddenConsoleHelper(t *testing.T) {
	if os.Getenv("AO_HIDDEN_CONSOLE_HELPER") != "1" {
		return
	}
	getConsoleWindow := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleWindow")
	window, _, callErr := getConsoleWindow.Call()
	if window != 0 {
		fmt.Fprintf(os.Stderr, "helper inherited console window %#x: %v\n", window, callErr)
		os.Exit(1)
	}
}
