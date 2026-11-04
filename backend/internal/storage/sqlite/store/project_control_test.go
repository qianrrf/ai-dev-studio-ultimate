package store_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
)

func controlledWorker(project, request string, control domain.ProjectControl) domain.SessionRecord {
	r := sampleRecord(project)
	r.Harness, r.Mode = domain.HarnessDeepSeek, domain.SessionModeChat
	r.Control = &domain.ControlClaim{OwnerID: control.OwnerID, Revision: control.Revision}
	r.ClientRequestID, r.ClientRequestHash = request, "hash:"+request
	return r
}

func TestProjectControlConcurrentTakeoverAndRelease(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	seedProject(t, s, "control")
	first, err := s.SetProjectControl(ctx, "control", "external", "desktop", 0, time.Now())
	if err != nil || first.Revision != 1 {
		t.Fatalf("initial: %+v %v", first, err)
	}
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, owner := range []string{"a", "b"} {
		wg.Add(1)
		go func(i int, owner string) {
			defer wg.Done()
			_, errs[i] = s.SetProjectControl(ctx, "control", "external", owner, 1, time.Now())
		}(i, owner)
	}
	wg.Wait()
	wins, conflicts := 0, 0
	for _, err := range errs {
		if err == nil {
			wins++
		} else if errors.Is(err, domain.ErrProjectControlConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("takeover errors: %v", errs)
	}
	released, err := s.SetProjectControl(ctx, "control", "unowned", "", 2, time.Now())
	if err != nil || released.Revision != 3 {
		t.Fatalf("release: %+v %v", released, err)
	}
	if _, err := s.CreateSession(ctx, controlledWorker("control", "after-release", first)); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("released owner dispatched: %v", err)
	}
	if _, err := s.SetProjectControl(ctx, "control", "external", "desktop", 0, time.Now()); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("revision reset: %v", err)
	}
}

func TestControlledDispatchReplayFencingAndCapacity(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	seedProject(t, s, "control")
	c, err := s.SetProjectControl(ctx, "control", "external", "desktop", 0, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r := controlledWorker("control", "one", c)
	first, fresh, err := s.CreateClientRequestSession(ctx, r)
	if err != nil || !fresh {
		t.Fatalf("dispatch: %+v %v %v", first, fresh, err)
	}
	second, fresh, err := s.CreateClientRequestSession(ctx, r)
	if err != nil || fresh || second.ID != first.ID {
		t.Fatalf("duplicate: %+v %v %v", second, fresh, err)
	}
	r.ClientRequestHash = "changed"
	if _, _, err := s.CreateClientRequestSession(ctx, r); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("payload changed: %v", err)
	}
	if _, _, err := s.CreateClientRequestSession(ctx, controlledWorker("control", "two", c)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.CreateClientRequestSession(ctx, controlledWorker("control", "three", c)); !errors.Is(err, domain.ErrProjectWorkerCapacity) {
		t.Fatalf("third worker: %v", err)
	}
	if _, err := s.SetProjectControl(ctx, "control", "external", "new", 1, time.Now()); err != nil {
		t.Fatal(err)
	}
	r.ClientRequestHash = "hash:one"
	if _, _, err := s.CreateClientRequestSession(ctx, r); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("old replay: %v", err)
	}
	if _, found, err := s.GetSession(ctx, first.ID); err != nil || !found {
		t.Fatalf("takeover lost worker: %v %v", found, err)
	}
}

func TestControlledDispatchUncertainSeedCannotBeRepeated(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	seedProject(t, s, "control")
	c, err := s.SetProjectControl(ctx, "control", "external", "desktop", 0, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	r := controlledWorker("control", "uncertain", c)
	r.Metadata.WorkspacePath = ""
	created, _, err := s.CreateClientRequestSession(ctx, r)
	if err != nil {
		t.Fatal(err)
	}
	if deleted, err := s.DeleteSession(ctx, created.ID); err != nil || !deleted {
		t.Fatalf("delete seed: %v %v", deleted, err)
	}
	if _, _, err := s.CreateClientRequestSession(ctx, r); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("uncertain request relaunched: %v", err)
	}
}

func TestProjectRequestIdempotencyAndPolicyBypasses(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	seedProject(t, s, "control")
	seedProject(t, s, "other")
	prep := sampleRecord("control")
	prep.IsTaskPreparation = true
	prepared, err := s.CreateSession(ctx, prep)
	if err != nil {
		t.Fatal(err)
	}
	c, err := s.SetProjectControl(ctx, "control", "external", "desktop", 0, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateSession(ctx, sampleRecord("control")); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("direct bypass: %v", err)
	}
	if _, err := s.CreateSession(ctx, prep); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("preparation bypass: %v", err)
	}
	if _, err := s.PromoteTaskPreparation(ctx, prepared.ID, controlledWorker("control", "promote", c)); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("promotion bypass: %v", err)
	}
	in := domain.ProjectRequest{ID: "request", ProjectID: "control", Brief: "fix calculator", CreatedAt: time.Now()}
	first, err := s.AddProjectRequest(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	in.CreatedAt = time.Now().Add(time.Hour)
	again, err := s.AddProjectRequest(ctx, in)
	if err != nil || !again.CreatedAt.Equal(first.CreatedAt) {
		t.Fatalf("replay: %+v %v", again, err)
	}
	in.Brief = "different"
	if _, err := s.AddProjectRequest(ctx, in); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("changed brief: %v", err)
	}
	in.Brief, in.ProjectID = "fix calculator", "other"
	if _, err := s.AddProjectRequest(ctx, in); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("cross-project collision: %v", err)
	}
	list, err := s.ListProjectRequests(ctx, "control")
	if err != nil || len(list) != 1 {
		t.Fatalf("intake: %+v %v", list, err)
	}
}

func TestProjectControlInternalOwnerAndConcurrentCapacity(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	seedProject(t, s, "control")
	seedProject(t, s, "other")
	orch := sampleRecord("other")
	orch.Kind, orch.Harness, orch.Mode = domain.KindOrchestrator, domain.HarnessCodex, domain.SessionModeChat
	wrong, err := s.CreateSession(ctx, orch)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetProjectControl(ctx, "control", "internal", string(wrong.ID), 0, time.Now()); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("cross-project controller accepted: %v", err)
	}
	orch.ProjectID = "control"
	owner, err := s.CreateSession(ctx, orch)
	if err != nil {
		t.Fatal(err)
	}
	c, err := s.SetProjectControl(ctx, "control", "internal", string(owner.ID), 0, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	p, _, err := s.GetProject(ctx, "control")
	if err != nil || p.Config.Worker.Harness != domain.HarnessDeepSeek || p.Config.Orchestrator.Harness != domain.HarnessCodex || p.Config.ResolveReviewerHarness(domain.HarnessDeepSeek) != domain.ReviewerCodex {
		t.Fatalf("role configuration: %+v %v", p.Config, err)
	}
	var wg sync.WaitGroup
	errs := make([]error, 3)
	for i, request := range []string{"one", "two", "three"} {
		wg.Add(1)
		go func(i int, request string) {
			defer wg.Done()
			_, _, errs[i] = s.CreateClientRequestSession(ctx, controlledWorker("control", request, c))
		}(i, request)
	}
	wg.Wait()
	wins, full := 0, 0
	for _, err := range errs {
		if err == nil {
			wins++
		} else if errors.Is(err, domain.ErrProjectWorkerCapacity) {
			full++
		} else {
			t.Fatal(err)
		}
	}
	if wins != 2 || full != 1 {
		t.Fatalf("concurrent capacity: %v", errs)
	}
}
