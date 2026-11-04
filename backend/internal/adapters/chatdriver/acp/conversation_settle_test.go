package acp

import (
	"log/slog"
	"testing"
	"time"

	acpsdk "github.com/coder/acp-go-sdk"

	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestSettleOpenItemsSnapshotsConcurrentProviderUpdates(t *testing.T) {
	c := &conversation{
		log:            slog.New(slog.DiscardHandler),
		events:         make(chan ports.ChatEvent),
		messages:       map[string]string{"message": "original message"},
		thoughts:       map[string]string{"thought": "original thought"},
		nestedMessages: map[string]nestedMessageState{"nested": {text: "original response", parentID: "parent"}},
		tools:          map[string]*toolState{"tool": {id: "tool", title: "Original tool", status: acpsdk.ToolCallStatusInProgress}},
	}
	done := make(chan struct{})
	go func() {
		c.settleOpenItems("turn")
		close(done)
	}()
	// The first completion proves settlement has started. The unbuffered next
	// event keeps it from reaching the nested message and tool before this update.
	if event := nextEvent(t, c.events); event.ProviderItemID != "message" || event.Text != "original message" {
		t.Fatalf("message completion = %#v", event)
	}
	c.mu.Lock()
	c.nestedMessages["nested"] = nestedMessageState{text: "later response", parentID: "later-parent"}
	c.tools["tool"].title = "Later tool"
	c.tools["tool"].status = acpsdk.ToolCallStatusCompleted
	c.mu.Unlock()
	if event := nextEvent(t, c.events); event.ProviderItemID != "thought" || event.Text != "original thought" {
		t.Fatalf("thought completion = %#v", event)
	}
	if event := nextEvent(t, c.events); event.ProviderItemID != "nested" || event.Text != "original response" {
		t.Fatalf("nested completion = %#v", event)
	}
	if event := nextEvent(t, c.events); event.ProviderItemID != "tool" || event.Summary != "Original tool" {
		t.Fatalf("tool completion = %#v", event)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("settlement did not finish")
	}
}
