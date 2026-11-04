//go:build windows

package conpty

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConPTYChildExitClosesOutputAndRejectsResize(t *testing.T) {
	cmdPath := filepath.Join(os.Getenv("SystemRoot"), "System32", "cmd.exe")
	conn, err := newConPTY(t.TempDir(), cmdPath, []string{
		"/d", "/s", "/c", "echo ao-conpty-exit",
	}, initialConPTYColumns, initialConPTYRows)
	if err != nil {
		t.Fatalf("newConPTY: %v", err)
	}
	defer conn.Close()

	type readResult struct {
		output []byte
		err    error
	}
	resultC := make(chan readResult, 1)
	go func() {
		output, readErr := io.ReadAll(conn)
		resultC <- readResult{output: output, err: readErr}
	}()

	select {
	case result := <-resultC:
		if result.err != nil {
			t.Fatalf("read ConPTY output: %v", result.err)
		}
		if !bytes.Contains(result.output, []byte("ao-conpty-exit")) {
			t.Fatalf("output %q does not contain completion marker", result.output)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ConPTY output did not close after the child exited")
	}
	if err := conn.Resize(120, 40); err == nil {
		t.Fatal("Resize succeeded after the pseudoconsole closed")
	}

	select {
	case <-conn.Done():
	case <-time.After(time.Second):
		t.Fatal("Done was not closed after the child exited")
	}
	if code, exited := conn.ExitCode(); !exited || code != 0 {
		t.Fatalf("ExitCode() = (%d, %v), want (0, true)", code, exited)
	}
}

// TestConPTYStartsAtRequestedGrid is the Windows half of the early-start path:
// a shell opened with a client grid (pty-host --size) starts its ConPTY at that
// grid, and the child sees it from its first output.
func TestConPTYStartsAtRequestedGrid(t *testing.T) {
	cols, rows, deferred := hostStartGrid(HostOptions{StartOnAttach: true, Cols: 146, Rows: 37})
	if deferred {
		t.Fatal("an explicit grid must start the ConPTY immediately")
	}
	cmdPath := filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	conn, err := newConPTY(t.TempDir(), cmdPath, []string{"-NoProfile", "-NonInteractive", "-Command", "[Console]::WriteLine(('AO_GRID:{0}x{1}' -f [Console]::WindowWidth, [Console]::WindowHeight))"}, cols, rows)
	if err != nil {
		t.Fatalf("newConPTY: %v", err)
	}
	defer conn.Close()

	resultC := make(chan []byte, 1)
	go func() {
		output, _ := io.ReadAll(conn)
		resultC <- output
	}()
	var output []byte
	select {
	case output = <-resultC:
	case <-time.After(10 * time.Second):
		t.Fatal("console size probe did not finish")
	}
	if !bytes.Contains(output, []byte("AO_GRID:146x37")) {
		t.Fatalf("console size output %q, want 146 columns and 37 rows", output)
	}
}
