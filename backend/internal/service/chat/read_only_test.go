package chat_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
	chatsvc "github.com/aoagents/agent-orchestrator/backend/internal/service/chat"
)

func TestReadOnlySessionSurvivesMetadataUpdatesAndControllerRecovery(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	rec := domain.SessionRecord{ID: "p1-2", ProjectID: testProject, Kind: domain.KindWorker,
		Harness: domain.HarnessCodex, Mode: domain.SessionModeChat, ReadOnly: true,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if _, err := st.CreateSession(ctx, rec); err != nil {
		t.Fatal(err)
	}
	// Ordinary metadata writers cannot clear a lifetime launch constraint.
	rec.ReadOnly = false
	if err := st.UpdateSession(ctx, rec); err != nil {
		t.Fatal(err)
	}
	saved, found, err := st.GetSession(ctx, rec.ID)
	if err != nil || !found || !saved.ReadOnly {
		t.Fatalf("stored constraint lost: %+v, %v", saved, err)
	}

	workspace := t.TempDir()
	for _, resume := range []bool{false, true} {
		var startCfg ports.ChatStartConfig
		var resumeCfg ports.ChatResumeConfig
		driver := fakeDriver{conv: newFakeConversation(), startCfg: &startCfg, resumeCfg: &resumeCfg}
		svc := chatsvc.New(chatsvc.Options{Store: st, Sessions: st, Drivers: fakeRegistry{driver: driver}, NewID: uuid.NewString})
		cfg := chatsvc.StartConfig{SessionID: rec.ID, ProjectID: testProject, Harness: domain.HarnessCodex, WorkspacePath: workspace}
		if resume {
			cfg.ProviderConversationID = "thread-1"
		}
		controller, err := svc.Start(ctx, cfg)
		if err != nil {
			t.Fatal(err)
		}
		if (!resume && (!startCfg.ReadOnly || !startCfg.DisableExternalTools || !startCfg.DisableNetwork)) || (resume && (!resumeCfg.ReadOnly || !resumeCfg.DisableExternalTools || !resumeCfg.DisableNetwork)) {
			t.Fatal("provider launched without persisted read-only constraint")
		}
		settings := controller.Settings()
		settings.ApprovalMode = ports.PermissionModeBypassPermissions
		if _, err := svc.SetTurnSettingsForOwner(ctx, domain.SessionConversationOwner(rec.ID), settings); !errors.Is(err, chatsvc.ErrReviewerPermissionsFixed) {
			t.Fatalf("permission change: %v", err)
		}
		if err := svc.Stop(ctx, rec.ID); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReadOnlySessionRefusesUnsupportedProviderBeforeLaunch(t *testing.T) {
	ctx := context.Background()
	st := openStore(t)
	rec := domain.SessionRecord{ID: "p1-2", ProjectID: testProject, Kind: domain.KindWorker,
		Harness: domain.HarnessCodex, Mode: domain.SessionModeChat, ReadOnly: true,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if _, err := st.CreateSession(ctx, rec); err != nil {
		t.Fatal(err)
	}
	svc := chatsvc.New(chatsvc.Options{Store: st, Sessions: st})
	if _, err := svc.Start(ctx, chatsvc.StartConfig{SessionID: rec.ID, Harness: domain.HarnessDeepSeek}); !errors.Is(err, ports.ErrChatUnsupported) {
		t.Fatalf("unsupported start: %v", err)
	}
	var refused *domain.ProviderSwitchError
	if err := svc.SwitchProvider(ctx, rec.ID, domain.ProviderSwitchRequest{TargetHarness: domain.HarnessDeepSeek}); !errors.As(err, &refused) || refused.Kind != domain.ProviderSwitchRefusalUnsupported {
		t.Fatalf("unsupported provider switch: %v", err)
	}
}
