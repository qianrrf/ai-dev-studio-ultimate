// Package gitpath checks the supported Windows Git directory boundary before work starts.
package gitpath

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf16"

	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/apierr"
)

// MaxRootUnits leaves room for Git's Windows process and metadata paths.
// shortcut: physical Git roots above 200 UTF-16 units need a different Git startup strategy.
const MaxRootUnits = 200

// CheckRoot resolves existing directory aliases, including parents of a new directory.
// It does not create, move or remove anything, and does not limit files within a repository.
func CheckRoot(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if runtime.GOOS != "windows" || path == "" {
		return nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	physical := filepath.Clean(abs)
	suffix := ""
	for {
		resolved, resolveErr := physicalPath(physical)
		if resolveErr == nil {
			physical = filepath.Join(resolved, suffix)
			break
		}
		if !errors.Is(resolveErr, os.ErrNotExist) {
			return resolveErr
		}
		parent := filepath.Dir(physical)
		if parent == physical {
			return resolveErr
		}
		suffix = filepath.Join(filepath.Base(physical), suffix)
		physical = parent
	}
	units := len(utf16.Encode([]rune(physical)))
	if units <= MaxRootUnits {
		return nil
	}
	return apierr.Invalid("WINDOWS_GIT_PATH_TOO_LONG",
		"This Windows Git directory is too long. Choose a shorter project or data folder, such as D:\\Projects\\my-app.",
		map[string]any{"path": physical, "utf16Units": units, "maxUtf16Units": MaxRootUnits})
}

// CheckRepository includes linked-worktree and separate Git metadata directories.
func CheckRepository(ctx context.Context, repo string) error {
	_, err := repositoryDirs(ctx, repo)
	return err
}

// CheckWorktree checks both the new root and Git's planned administrative directory.
func CheckWorktree(ctx context.Context, repo, worktree string) error {
	common, err := repositoryDirs(ctx, repo)
	if err != nil {
		return err
	}
	if err := CheckRepository(ctx, worktree); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return CheckRoot(ctx, filepath.Join(common, "worktrees", filepath.Base(worktree)))
	}
	return nil
}

func repositoryDirs(ctx context.Context, repo string) (string, error) {
	if runtime.GOOS != "windows" || repo == "" {
		return "", ctx.Err()
	}
	if err := CheckRoot(ctx, repo); err != nil {
		return "", err
	}
	gitdir := filepath.Join(repo, ".git")
	if err := CheckRoot(ctx, gitdir); err != nil {
		return "", err
	}
	info, err := os.Stat(gitdir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err == nil && !info.IsDir() {
		data, readErr := os.ReadFile(gitdir)
		if readErr != nil {
			return "", readErr
		}
		if target, ok := strings.CutPrefix(strings.TrimSpace(string(data)), "gitdir: "); ok {
			gitdir = relativeTo(repo, target)
			if err := CheckRoot(ctx, gitdir); err != nil {
				return "", err
			}
		}
	}
	common := gitdir
	data, err := os.ReadFile(filepath.Join(gitdir, "commondir"))
	if err == nil {
		common = relativeTo(gitdir, strings.TrimSpace(string(data)))
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	return common, CheckRoot(ctx, common)
}

func relativeTo(parent, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(parent, path)
}
