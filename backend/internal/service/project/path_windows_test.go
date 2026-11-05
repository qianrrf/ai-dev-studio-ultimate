//go:build windows

package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/aoagents/agent-orchestrator/backend/internal/service/project"
)

func TestWindowsRootRefusalHasNoProjectOrFilesystemSideEffects(t *testing.T) {
	m := newManager(t)
	parent := t.TempDir()
	long := filepath.Join(parent, strings.Repeat("x", 201-len(utf16.Encode([]rune(parent)))-1))
	ctx := t.Context()
	_, err := m.Add(ctx, project.AddInput{Path: long})
	wantCode(t, err, "WINDOWS_GIT_PATH_TOO_LONG")
	_, err = m.InitializeRepository(ctx, project.InitializeRepositoryInput{Path: long})
	wantCode(t, err, "WINDOWS_GIT_PATH_TOO_LONG")
	_, err = m.CheckCloneDestination(ctx, project.CloneInput{RemoteURL: "https://github.com/example/repo.git", DestinationParent: long})
	wantCode(t, err, "WINDOWS_GIT_PATH_TOO_LONG")
	for _, prepare := range []bool{true, false} {
		in := project.CloneInput{RemoteURL: "https://github.com/example/repo.git", DestinationParent: long}
		if prepare {
			_, err = m.PrepareClone(ctx, in)
		} else {
			_, err = m.Clone(ctx, in)
		}
		wantCode(t, err, "WINDOWS_GIT_PATH_TOO_LONG")
	}
	if _, err := os.Stat(long); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("refusal created destination: %v", err)
	}
	projects, err := m.List(ctx)
	if err != nil || len(projects) != 0 {
		t.Fatalf("refusal registered projects: %v, %v", projects, err)
	}
}

func TestWindowsCloneChecksStagingRootBeforeCreatingParent(t *testing.T) {
	m := newManager(t)
	base := t.TempDir()
	// repo/.git fits, but the private staging root's .git exceeds 200.
	parent := filepath.Join(base, strings.Repeat("x", 180-len(utf16.Encode([]rune(base)))-1))
	_, err := m.PrepareClone(t.Context(), project.CloneInput{
		RemoteURL: "https://github.com/example/r.git", DestinationParent: parent,
	})
	wantCode(t, err, "WINDOWS_GIT_PATH_TOO_LONG")
	if _, err := os.Stat(parent); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("refusal created staging parent: %v", err)
	}
}
