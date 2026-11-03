package sessionmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestReadOnlySessionCannotSwitchToTerminal(t *testing.T) {
	m, st, _, _, operations := newTransitionManager(t, domain.SessionModeChat)
	rec := st.sessions["session-1"]
	rec.Harness, rec.ReadOnly = domain.HarnessCodex, true
	st.sessions[rec.ID] = rec
	_, err := m.StartInterfaceTransition(t.Context(), rec.ID, domain.SessionModeTUI, domain.SessionInterfaceTransitionDrain, domain.SessionInterfaceTransitionHistoryStrict)
	if !errors.Is(err, ErrInterfaceHandoffUnsupported) {
		t.Fatalf("terminal switch: %v", err)
	}
	if len(st.transitions) != 0 || len(*operations) != 0 {
		t.Fatal("unsupported transition mutated controller state")
	}
}

func TestReadOnlySpawnRejectsUnsupportedControllerWithoutCreatingSession(t *testing.T) {
	for _, tc := range []struct {
		harness domain.AgentHarness
		mode    domain.SessionMode
	}{
		{domain.HarnessCodex, domain.SessionModeTUI},
		{domain.HarnessClaudeCode, domain.SessionModeChat},
	} {
		m, st, rt := newChatManager(&recordingLauncher{})
		_, _, _, err := m.Spawn(context.Background(), ports.SpawnConfig{ProjectID: chatTestProject,
			Kind: domain.KindWorker, Harness: tc.harness, RequestedMode: tc.mode, ReadOnly: true})
		if !errors.Is(err, ports.ErrChatUnsupported) {
			t.Fatalf("%s/%s: %v", tc.harness, tc.mode, err)
		}
		if len(st.sessions) != 0 || rt.created != 0 {
			t.Fatal("unsupported read-only request created durable/runtime state")
		}
	}
}
