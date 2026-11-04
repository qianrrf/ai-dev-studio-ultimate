package sessionmanager

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestDeliveryReviewUsesFixedHeadAndSynchronousReadySeed(t *testing.T) {
	workspace := &fakeWorkspace{}
	manager := &Manager{workspace: workspace}
	project := domain.ProjectRecord{ID: "demo", Path: t.TempDir()}
	sha := strings.Repeat("a", 40)
	cfg := ports.SpawnConfig{ProjectID: "demo", Kind: domain.KindWorker, Harness: domain.HarnessCodex, ReadOnly: true, RequestedMode: domain.SessionModeChat, ReviewCommit: sha, Control: &domain.ControlClaim{OwnerID: "desktop", Revision: 1}}
	_, _, err := manager.createSessionWorkspace(context.Background(), project, cfg, "demo-2", "review/demo-2", map[string]string{filepath.Clean(project.Path): "moving-default"})
	if err != nil || workspace.lastCfg.FixedCommit != sha || !workspace.lastCfg.FreshBranch {
		t.Fatalf("review workspace: %+v %v", workspace.lastCfg, err)
	}
	seed := seedRecord(cfg, domain.ProjectConfig{AutoReview: true}, time.Now())
	if seed.ProvisionState.IsProvisioning() || seed.AutoReviewEnabled || !seed.ReadOnly {
		t.Fatalf("synchronous review seed: %+v", seed)
	}
	cfg.ReadOnly = false
	if _, _, err := manager.createSessionWorkspace(context.Background(), project, cfg, "demo-3", "review/demo-3", nil); err == nil {
		t.Fatal("writable fixed-head review accepted")
	}
}
