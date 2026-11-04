//go:build !windows

package session

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
)

// POSIX directory search permission is checked during stat. Windows metadata
// queries do not enforce this bit, so this syscall contract runs on Unix.
func TestWorkspaceLocationReportsUnreadableWorkspace(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	parent := t.TempDir()
	workspace := filepath.Join(parent, "workspace")
	if err := os.Mkdir(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(parent, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o700) })
	store := newFakeStore()
	store.sessions["ao-1"] = domain.SessionRecord{
		ID:       "ao-1",
		Metadata: domain.SessionMetadata{WorkspacePath: workspace},
	}

	_, err := (&Service{store: store}).WorkspaceLocation(context.Background(), "ao-1")
	assertAPIErrorCode(t, err, "SESSION_WORKSPACE_UNREADABLE")
}
