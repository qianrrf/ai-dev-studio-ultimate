//go:build windows

package gitpath

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/apierr"
	aoprocess "github.com/aoagents/agent-orchestrator/backend/internal/process"
)

func rootOfUnits(t *testing.T, parent string, units int) string {
	t.Helper()
	base := filepath.Join(parent, "中文 spaced 🙂")
	padding := units - len(utf16.Encode([]rune(base)))
	if padding < 0 {
		t.Fatalf("test parent exceeds boundary: %s", parent)
	}
	return base + strings.Repeat("x", padding)
}

func requirePathError(t *testing.T, err error) {
	t.Helper()
	var rejected *apierr.Error
	if !errors.As(err, &rejected) || rejected.Code != "WINDOWS_GIT_PATH_TOO_LONG" {
		t.Fatalf("error = %v, want Windows path refusal", err)
	}
}

func TestRootCountsUTF16AndUncreatedParents(t *testing.T) {
	parent := t.TempDir()
	for _, units := range []int{199, 200, 201} {
		path := rootOfUnits(t, parent, units)
		err := CheckRoot(t.Context(), path)
		if units > MaxRootUnits {
			requirePathError(t, err)
		} else if err != nil {
			t.Fatalf("%d UTF-16 units: %v", units, err)
		}
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("validation created directory: %v", err)
		}
	}
}

func TestRootResolvesJunctionBeforeCheckingLength(t *testing.T) {
	parent := t.TempDir()
	long := rootOfUnits(t, parent, 201)
	if err := os.Mkdir(long, 0700); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(parent, "alias")
	if output, err := aoprocess.CommandContext(t.Context(), "cmd.exe", "/d", "/c", "mklink", "/J", alias, long).CombinedOutput(); err != nil {
		t.Fatalf("create directory junction: %v\n%s", err, output)
	}
	t.Cleanup(func() { _ = os.Remove(alias) })
	requirePathError(t, CheckRoot(t.Context(), alias))
	requirePathError(t, CheckRoot(t.Context(), filepath.Join(alias, "not-created", "child")))
}

func TestRepositoryChecksGitdirAndCommonDirectory(t *testing.T) {
	for _, kind := range []string{"gitdir", "commondir"} {
		t.Run(kind, func(t *testing.T) {
			parent := t.TempDir()
			repo := filepath.Join(parent, "repo")
			gitdir := filepath.Join(parent, "metadata")
			long := rootOfUnits(t, parent, 201)
			for _, dir := range []string{repo, gitdir, long} {
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
			}
			target := gitdir
			if kind == "gitdir" {
				target = long
			} else if err := os.WriteFile(filepath.Join(gitdir, "commondir"), []byte(long), 0600); err != nil {
				t.Fatal(err)
			}
			marker := []byte("gitdir: " + target + "\n")
			if err := os.WriteFile(filepath.Join(repo, ".git"), marker, 0600); err != nil {
				t.Fatal(err)
			}
			requirePathError(t, CheckRepository(t.Context(), repo))
			after, err := os.ReadFile(filepath.Join(repo, ".git"))
			if err != nil || string(after) != string(marker) {
				t.Fatalf("validation changed metadata: %v", err)
			}
		})
	}
}

func TestWorktreeChecksFutureAdministrativeDirectory(t *testing.T) {
	parent := t.TempDir()
	repo := rootOfUnits(t, parent, 180)
	worktree := filepath.Join(parent, "worker-with-long-id")
	requirePathError(t, CheckWorktree(t.Context(), repo, worktree))
	if _, err := os.Stat(repo); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("validation created repository: %v", err)
	}
}
