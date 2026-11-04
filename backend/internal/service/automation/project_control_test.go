package automation

import (
	"context"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/storage/sqlite/sqlitetest"
)

func TestControlledAutomationDeliversIntakeWithoutStartingAnAgent(t *testing.T) {
	ctx, store := context.Background(), sqlitetest.MustOpen(t)
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	if err := store.UpsertProject(ctx, domain.ProjectRecord{ID: "p", Path: t.TempDir(), RegisteredAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetProjectControl(ctx, "p", "external", "desktop", 0, now); err != nil {
		t.Fatal(err)
	}
	definition, err := store.CreateAutomation(ctx, domain.Automation{ID: "daily", ProjectID: "p", DisplayName: "Review", Prompt: "Review calculator", Kind: domain.KindWorker, Harness: domain.HarnessCodex, RRuleText: "DTSTART:20261104T120000Z\nRRULE:FREQ=DAILY", Timezone: "UTC", Enabled: true, NextRunAt: now, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	// No spawner: this path must only deliver the requirement to the inbox.
	svc := New(Deps{Store: store, Clock: func() time.Time { return now }})
	for i := 0; i < 2; i++ {
		if err := svc.Tick(ctx, now); err != nil {
			t.Fatal(err)
		}
	}
	requests, err := store.ListProjectRequests(ctx, "p")
	if err != nil || len(requests) != 1 || requests[0].Brief != definition.Prompt {
		t.Fatalf("requests=%+v err=%v", requests, err)
	}
	runs, err := store.ListAutomationRuns(ctx, domain.AutomationRunFilter{AutomationID: definition.ID, Limit: 10})
	if err != nil || len(runs.Items) != 1 || runs.Items[0].Status != domain.AutomationRunCompleted || runs.Items[0].SessionID != nil {
		t.Fatalf("runs=%+v err=%v", runs, err)
	}
	sessions, err := store.ListSessions(ctx, "p")
	if err != nil || len(sessions) != 0 {
		t.Fatalf("automation bypassed controller: %+v %v", sessions, err)
	}
}
