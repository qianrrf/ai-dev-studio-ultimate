package lifecycle

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/storage/sqlite/sqlitetest"
)

func TestControlledFeedbackBecomesOneDurableControllerReport(t *testing.T) {
	ctx, store := context.Background(), sqlitetest.MustOpen(t)
	now := time.Now().UTC()
	if err := store.UpsertProject(ctx, domain.ProjectRecord{ID: "p", Path: t.TempDir(), RegisteredAt: now}); err != nil {
		t.Fatal(err)
	}
	c, err := store.SetProjectControl(ctx, "p", "external", "desktop", 0, now)
	if err != nil {
		t.Fatal(err)
	}
	worker, err := store.CreateSession(ctx, domain.SessionRecord{ProjectID: "p", Kind: domain.KindWorker, Harness: domain.HarnessDeepSeek, Mode: domain.SessionModeChat, Control: &domain.ControlClaim{OwnerID: c.OwnerID, Revision: c.Revision}, ClientRequestID: "worker-1", ClientRequestHash: "hash", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	m := New(store, nil)
	for i := 0; i < 2; i++ {
		if routed, err := m.routeControlledNudge(ctx, worker.ID, "ci:pr", "checks-v1", "CI failed. Repair required."); err != nil || !routed {
			t.Fatalf("routed=%v err=%v", routed, err)
		}
	}
	reports, err := store.ListReportsByProject(ctx, "p")
	if err != nil || len(reports) != 1 || reports[0].DeliveryState != domain.ReportPending || reports[0].State != domain.ReportNeedsInput || !strings.Contains(reports[0].Note, "Controller decision required") {
		t.Fatalf("reports=%+v err=%v", reports, err)
	}
	if routed, err := m.routeControlledNudge(ctx, worker.ID, "ci:pr", "checks-v2", "New failure"); err != nil || !routed {
		t.Fatal(routed, err)
	}
	reports, err = store.ListReportsByProject(ctx, "p")
	if err != nil || len(reports) != 2 {
		t.Fatalf("new observation lost: %v %v", reports, err)
	}
}
