package report

import (
	"context"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
)

type controlledCoordinatorStore struct {
	*coordinatorStore
	control domain.ProjectControl
}

func (s *controlledCoordinatorStore) GetProjectControl(context.Context, domain.ProjectID) (domain.ProjectControl, bool, error) {
	return s.control, true, nil
}

func TestControlledReportsWaitForExternalOwnerAndRouteOnlyToCurrentCodex(t *testing.T) {
	now := time.Now()
	store := &controlledCoordinatorStore{coordinatorStore: newCoordinatorStore(now), control: domain.ProjectControl{ProjectID: "p", OwnerKind: "external", OwnerID: "desktop", Revision: 1}}
	store.sessions = []domain.SessionRecord{
		{ID: "old", ProjectID: "p", Kind: domain.KindOrchestrator, Harness: domain.HarnessCodex, Mode: domain.SessionModeChat, CreatedAt: now.Add(time.Hour)},
		{ID: "current", ProjectID: "p", Kind: domain.KindOrchestrator, Harness: domain.HarnessCodex, Mode: domain.SessionModeChat, CreatedAt: now},
	}
	store.add(reportAt("r", "worker", now, now, "test failed"))
	delivery := &coordinatorDelivery{}
	c := NewCoordinator(CoordinatorDeps{Store: store, Delivery: delivery, Now: func() time.Time { return now }})
	_, _ = c.RunDue(context.Background())
	if len(delivery.submissions) != 0 || store.acked != 0 || store.reports[0].DeliveryState != domain.ReportPending {
		t.Fatalf("external owner report consumed: %+v", store.reports)
	}
	store.control.OwnerKind, store.control.OwnerID, store.control.Revision = "internal", "current", 2
	if _, err := c.RunDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(delivery.submissions) != 1 || delivery.submissions[0].target != "current" {
		t.Fatalf("wrong controller: %+v", delivery.submissions)
	}
}
