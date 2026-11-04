//go:build windows

package systemexec

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"

	aoprocess "github.com/aoagents/agent-orchestrator/backend/internal/process"
)

func commandContext(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
	resolved, err := exec.LookPath(name)
	if err != nil {
		return nil, err
	}
	extension := filepath.Ext(resolved)
	if !strings.EqualFold(extension, ".cmd") && !strings.EqualFold(extension, ".bat") {
		return aoprocess.CommandContext(ctx, resolved, args...), nil //nolint:gosec // Callers supply server-owned argv.
	}

	shell := strings.TrimSpace(os.Getenv("ComSpec"))
	if shell == "" {
		shell = "cmd.exe"
	}
	cmd := aoprocess.CommandContext(ctx, shell) //nolint:gosec // ComSpec is Windows' configured batch interpreter.
	cmd.Args = nil
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CmdLine = `/d /s /c "` + windowsBatchCommandLine(resolved, args) + `"`
	return cmd, nil
}

func windowsBatchCommandLine(executable string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, quoteWindowsBatchArg(executable))
	for _, arg := range args {
		parts = append(parts, quoteWindowsBatchArg(arg))
	}
	return strings.Join(parts, " ")
}

func quoteWindowsBatchArg(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func refreshExecutablePath() {
	paths := make([]string, 0, 9)
	paths = append(paths,
		os.Getenv("Path"),
		registryPath(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`),
		registryPath(registry.CURRENT_USER, `Environment`),
	)
	paths = append(paths, windowsGitPaths(os.Getenv, registryString)...)
	if merged := mergeWindowsPath(paths...); merged != "" {
		_ = os.Setenv("Path", merged)
	}
}

func registryPath(root registry.Key, keyPath string) string {
	return registryString(root, keyPath, "Path")
}

func registryString(root registry.Key, keyPath, name string) string {
	key, err := registry.OpenKey(root, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer func() { _ = key.Close() }()
	value, valueType, err := key.GetStringValue(name)
	if err != nil {
		return ""
	}
	if valueType == registry.EXPAND_SZ {
		if expanded, expandErr := registry.ExpandString(value); expandErr == nil {
			value = expanded
		}
	}
	return value
}

func windowsGitPaths(env func(string) string, readRegistry func(registry.Key, string, string) string) []string {
	roots := []string{
		readRegistry(registry.CURRENT_USER, `Software\GitForWindows`, "InstallPath"),
		readRegistry(registry.LOCAL_MACHINE, `Software\GitForWindows`, "InstallPath"),
		readRegistry(registry.LOCAL_MACHINE, `Software\WOW6432Node\GitForWindows`, "InstallPath"),
	}
	for _, name := range []string{"ProgramFiles", "ProgramFiles(x86)"} {
		if root := env(name); root != "" {
			roots = append(roots, filepath.Join(root, "Git"))
		}
	}
	if root := env("LOCALAPPDATA"); root != "" {
		roots = append(roots, filepath.Join(root, "Programs", "Git"))
	}
	paths := make([]string, 0, len(roots))
	for _, root := range roots {
		if root == "" {
			continue
		}
		dir := filepath.Join(root, "cmd")
		if info, err := os.Stat(filepath.Join(dir, "git.exe")); err == nil && info.Mode().IsRegular() {
			paths = append(paths, dir)
		}
	}
	return paths
}

func mergeWindowsPath(values ...string) string {
	seen := make(map[string]struct{})
	merged := make([]string, 0)
	for _, value := range values {
		for _, entry := range strings.Split(value, ";") {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				continue
			}
			key := strings.ToLower(strings.TrimRight(entry, `\/`))
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, entry)
		}
	}
	return strings.Join(merged, ";")
}
