package chat_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
	chatsvc "github.com/aoagents/agent-orchestrator/backend/internal/service/chat"
)

func TestReadOnlyReviewerCannotOpenOrContinueIndependentSide(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	record, found, err := h.st.GetSession(ctx, testSession)
	if err != nil || !found {
		t.Fatalf("session: %v %v", found, err)
	}
	record.ReadOnly = true
	record.ID = "readonly-reviewer"
	record, err = h.st.CreateSession(ctx, record)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.svc.CreateIndependentSideChat(ctx, record.ID, chatsvc.SideCreateRequest{IdempotencyKey: "side"}); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("read-only side creation: %v", err)
	}
	if _, err := h.svc.SendSideQuestion(ctx, record.ID, "side", ports.ChatUserMessage{Text: "bypass", ClientMessageID: "side"}); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("read-only side send: %v", err)
	}
	if err := h.svc.RetrySideQuestion(ctx, record.ID, "side", "turn"); !errors.Is(err, domain.ErrProjectControlConflict) {
		t.Fatalf("read-only side retry: %v", err)
	}
}
