package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
)

func TestDeliveryReceiptsFenceReworkAndRequireMatchingEvidence(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	seedProject(t, s, "control")
	control, err := s.SetProjectControl(ctx, "control", "external", "desktop", 0, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	seed := controlledWorker("control", "worker", control)
	seed.Activity.State = domain.ActivityIdle
	worker, _, err := s.CreateClientRequestSession(ctx, seed)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CommitClientRequestSession(ctx, worker.ID); err != nil {
		t.Fatal(err)
	}
	claim := domain.ControlClaim{OwnerID: control.OwnerID, Revision: control.Revision}
	conversation, err := s.CreateConversation(ctx, "controlled-conversation", domain.ConversationScopeSession, "control", worker.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ClaimChatControllerGeneration(ctx, worker.ID, "controlled-generation"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AppendUserMessage(ctx, conversation.ID, worker.ID, "controlled-generation", domain.ConversationMessage{ID: "bypass", Text: "unregistered rework", Origin: domain.MessageOriginHuman}, "bypass-turn", time.Now()); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("direct durable turn bypass: %v", err)
	}
	reserve := func(id, kind, testID, reviewID string) domain.DeliveryEvidence {
		t.Helper()
		rec, fresh, err := s.ReserveDeliveryEvidence(ctx, domain.DeliveryEvidence{ID: id, ProjectID: "control", WorkerID: worker.ID, WorkerPath: worker.Metadata.WorkspacePath, Request: domain.DeliveryRequest{ID: id, Control: claim, Kind: kind, BaseSHA: "base", HeadSHA: "head", TestID: testID, ReviewID: reviewID}, CreatedAt: time.Now()})
		if err != nil || !fresh {
			t.Fatalf("reserve %s: %+v %v %v", id, rec, fresh, err)
		}
		return rec
	}
	complete := func(id string, result domain.DeliveryResult) {
		t.Helper()
		if err := s.CompleteDeliveryEvidence(ctx, id, result, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	zero := 0
	for round := 1; round <= 3; round++ {
		testID, reviewID, verdictID := fmt.Sprintf("test-%d", round), fmt.Sprintf("review-%d", round), fmt.Sprintf("verdict-%d", round)
		reserve(testID, "test", "", "")
		complete(testID, domain.DeliveryResult{ExitCode: &zero})
		reserve(reviewID, "review", testID, "")
		complete(reviewID, domain.DeliveryResult{ReviewerID: "reviewer", ReviewerPath: "/readonly"})
		reserve(verdictID, "verdict", "", reviewID)
		complete(verdictID, domain.DeliveryResult{Verdict: "changes_requested"})
		in := domain.DeliveryEvidence{ID: fmt.Sprintf("rework-%d", round), ProjectID: "control", WorkerID: worker.ID, WorkerPath: worker.Metadata.WorkspacePath, Request: domain.DeliveryRequest{ID: fmt.Sprintf("rework-%d", round), Control: claim, Kind: "rework", BaseSHA: "base", HeadSHA: "head", ReviewID: verdictID}, CreatedAt: time.Now()}
		_, fresh, err := s.ReserveDeliveryEvidence(ctx, in)
		if round == 3 {
			if !errors.Is(err, domain.ErrProjectControlConflict) {
				t.Fatalf("third rework: %v", err)
			}
			break
		}
		if err != nil || !fresh {
			t.Fatalf("rework: %v %v", fresh, err)
		}
		if err := s.CheckProjectChatDelivery(ctx, worker.ID); !errors.Is(err, domain.ErrProjectControlConflict) {
			t.Fatalf("direct send bypass: %v", err)
		}
		if err := s.CheckProjectChatDelivery(domain.WithDeliveryReceipt(ctx, in.ID), worker.ID); err != nil {
			t.Fatalf("registered continuation: %v", err)
		}
		_, replayFresh, err := s.ReserveDeliveryEvidence(ctx, in)
		if err != nil || replayFresh {
			t.Fatalf("uncertain receipt reexecuted: %v %v", replayFresh, err)
		}
		complete(in.ID, domain.DeliveryResult{Delivered: true})
		in.ID, in.Request.ID = in.ID+"-duplicate", in.ID+"-duplicate"
		if _, _, err := s.ReserveDeliveryEvidence(ctx, in); !errors.Is(err, domain.ErrProjectControlConflict) {
			t.Fatalf("same verdict reworked twice: %v", err)
		}
	}
	if _, err := s.SetProjectControl(ctx, "control", "external", "next", 1, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.ReserveDeliveryEvidence(ctx, domain.DeliveryEvidence{ID: "stale", ProjectID: "control", WorkerID: worker.ID, WorkerPath: worker.Metadata.WorkspacePath, Request: domain.DeliveryRequest{Control: claim, Kind: "test"}}); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("old controller: %v", err)
	}
	// Accepted before takeover may finish; old receipts do not authorize new dispatch.
	if err := s.CheckProjectChatDelivery(domain.WithDeliveryReceipt(ctx, "rework-1"), worker.ID); err != nil {
		t.Fatalf("registered operation lost on takeover: %v", err)
	}
}

func TestDeliveryReviewerMessageRequiresSettledSuccessfulTurn(t *testing.T) {
	s, session, conversation := conversationFixture(t)
	ctx := context.Background()
	if err := s.AdoptProviderTurn(ctx, conversation, session, "gen-1", "review-turn", "provider-review", histClock); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendAssistantDelta(ctx, conversation, "verdict", "provider-review", "APPROVED\nObserved code", "verdict-message", histClock); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.CompletedReviewerMessage(ctx, session); err == nil {
		t.Fatal("streaming reviewer accepted")
	}
	if err := s.SettleTurn(ctx, conversation, "provider-review", domain.TurnStateCompleted, "", histClock.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	id, text, err := s.CompletedReviewerMessage(ctx, session)
	if err != nil || id != "verdict-message" || text != "APPROVED\nObserved code" {
		t.Fatalf("completed reviewer: %s %q %v", id, text, err)
	}
	if err := s.AdoptProviderTurn(ctx, conversation, session, "gen-1", "pending", "provider-pending", histClock.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.CompletedReviewerMessage(ctx, session); err == nil {
		t.Fatal("prior verdict accepted while reviewer has an unfinished turn")
	}
}

func TestControlledProjectRegistrationPreservesRolesAndRepository(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	seedProject(t, s, "control")
	if _, err := s.SetProjectControl(ctx, "control", "external", "desktop", 0, time.Now()); err != nil {
		t.Fatal(err)
	}
	p, _, err := s.GetProject(ctx, "control")
	if err != nil {
		t.Fatal(err)
	}
	p.Config = domain.ProjectConfig{}
	if err := s.UpsertProject(ctx, p); err != nil {
		t.Fatal(err)
	}
	got, _, _ := s.GetProject(ctx, "control")
	if got.Config.Worker.Harness != domain.HarnessDeepSeek || got.Config.ResolveReviewerHarness(domain.HarnessDeepSeek) != domain.ReviewerCodex {
		t.Fatalf("registration removed controller policy: %+v", got.Config)
	}
	p.Path += "-other"
	if err := s.UpsertProject(ctx, p); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("controlled repository replaced: %v", err)
	}
}
