//go:build windows

package systemexec

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"

	aoprocess "github.com/aoagents/agent-orchestrator/backend/internal/process"
)

func TestCommandContextWrapsBatchShimWithComSpec(t *testing.T) {
	dir := t.TempDir()
	shim := filepath.Join(dir, "agent tool.cmd")
	if err := os.WriteFile(shim, []byte("@echo off\r\necho %~1\r\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	comSpec := os.Getenv("ComSpec")
	if comSpec == "" {
		t.Fatal("ComSpec is not set")
	}

	cmd, err := commandContext(context.Background(), shim, "argument with spaces")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(cmd.Path, comSpec) {
		t.Fatalf("Path = %q, want ComSpec", cmd.Path)
	}
	wantCommand := quoteWindowsBatchArg(shim) + " " + quoteWindowsBatchArg("argument with spaces")
	if cmd.SysProcAttr == nil || !strings.Contains(cmd.SysProcAttr.CmdLine, wantCommand) {
		t.Fatalf("CmdLine = %q, want quoted shim and argument", cmd.SysProcAttr.CmdLine)
	}
	aoprocess.ConfigureTreeCancellation(cmd)
	if cmd.SysProcAttr.CmdLine == "" {
		t.Fatal("configureProcessGroup discarded the batch command line")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run batch shim: %v\n%s", err, output)
	}
	if got := strings.TrimSpace(string(output)); got != "argument with spaces" {
		t.Fatalf("output = %q, want batch argument preserved", got)
	}
}

func TestMergeWindowsPathAddsRegistryEntriesWithoutDuplicates(t *testing.T) {
	got := mergeWindowsPath(
		`C:\Windows\System32;C:\Users\tester\AppData\Roaming\npm`,
		`C:\Program Files\Git\cmd;C:\WINDOWS\SYSTEM32`,
		`C:\Users\tester\AppData\Local\Microsoft\WinGet\Links`,
	)
	want := strings.Join([]string{
		`C:\Windows\System32`,
		`C:\Users\tester\AppData\Roaming\npm`,
		`C:\Program Files\Git\cmd`,
		`C:\Users\tester\AppData\Local\Microsoft\WinGet\Links`,
	}, ";")
	if got != want {
		t.Fatalf("mergeWindowsPath() = %q, want %q", got, want)
	}
}

func TestWindowsGitPathsChecksRegistryAndStandardInstallations(t *testing.T) {
	root := t.TempDir()
	registered := filepath.Join(root, "custom Git")
	programFiles := filepath.Join(root, "Program Files")
	localData := filepath.Join(root, "用户 AppData")
	want := []string{filepath.Join(registered, "cmd"), filepath.Join(programFiles, "Git", "cmd"), filepath.Join(localData, "Programs", "Git", "cmd")}
	for _, dir := range want {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "git.exe"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	env := func(key string) string {
		return map[string]string{"ProgramFiles": programFiles, "LOCALAPPDATA": localData}[key]
	}
	read := func(key registry.Key, path, name string) string {
		if key == registry.CURRENT_USER && path == `Software\GitForWindows` && name == "InstallPath" {
			return registered
		}
		return filepath.Join(root, "missing")
	}
	got := windowsGitPaths(env, read)
	if strings.Join(got, ";") != strings.Join(want, ";") {
		t.Fatalf("windowsGitPaths() = %q, want %q", got, want)
	}
	if merged := mergeWindowsPath(want[1], strings.Join(got, ";")); !strings.HasPrefix(merged, want[1]+";") {
		t.Fatalf("existing Git lost priority: %q", merged)
	}
}
