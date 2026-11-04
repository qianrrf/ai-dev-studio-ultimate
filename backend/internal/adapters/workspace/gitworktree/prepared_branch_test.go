package gitworktree

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestFreshPreparedBranchSkipsRemoteCollision(t *testing.T) {
	git := requireGit(t)
	tmp := t.TempDir()
	repo := setupOriginClone(t, git, tmp)
	mainSHA := gitOutput(t, git, repo, "rev-parse", "refs/heads/main")
	runGit(t, git, repo, "update-ref", "refs/remotes/origin/ao/sess/root", mainSHA)
	ws, err := New(Options{Binary: git, ManagedRoot: filepath.Join(tmp, "managed"), RepoResolver: StaticRepoResolver{"proj": repo}})
	if err != nil {
		t.Fatal(err)
	}
	info, err := ws.Create(context.Background(), ports.WorkspaceConfig{
		ProjectID: "proj", SessionID: "sess", Branch: "ao/sess/root", FreshBranch: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if info.Branch != "ao/sess/root-2" {
		t.Fatalf("prepared branch = %q, want fresh suffix", info.Branch)
	}
	if got := gitOutput(t, git, info.Path, "rev-parse", "HEAD"); got != mainSHA {
		t.Fatalf("prepared HEAD = %q, want main %q", got, mainSHA)
	}
}

func TestFixedCommitReviewChecksOutWorkerCommitInsteadOfDefault(t *testing.T) {
	git := requireGit(t)
	tmp := t.TempDir()
	repo := setupOriginClone(t, git, tmp)
	base := gitOutput(t, git, repo, "rev-parse", "origin/main")
	runGit(t, git, repo, "checkout", "-b", "worker-change")
	runGit(t, git, repo, "commit", "--allow-empty", "-m", "worker change")
	head := gitOutput(t, git, repo, "rev-parse", "HEAD")
	if base == head {
		t.Fatal("fixture has no worker change")
	}
	ws, err := New(Options{Binary: git, ManagedRoot: filepath.Join(tmp, "managed"), RepoResolver: StaticRepoResolver{"proj": repo}})
	if err != nil {
		t.Fatal(err)
	}
	cfg := ports.WorkspaceConfig{ProjectID: "proj", SessionID: "review", Branch: "review", FreshBranch: true, RequireNewBranch: true, FixedCommit: head}
	info, err := ws.Create(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if got := gitOutput(t, git, info.Path, "rev-parse", "HEAD"); got != head || info.BaseSHA != head {
		t.Fatalf("review HEAD=%s base=%s want worker %s (default %s)", got, info.BaseSHA, head, base)
	}
	cfg.SessionID, cfg.Branch, cfg.FixedCommit = "invalid", "invalid", "worker-change"
	if _, err := ws.Create(context.Background(), cfg); err == nil {
		t.Fatal("moving ref accepted as immutable review commit")
	}
}
