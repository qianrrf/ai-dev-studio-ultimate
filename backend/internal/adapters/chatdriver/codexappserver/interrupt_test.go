package codexappserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestInterruptCleansOnlyItsThreadBeforeNextDispatch(t *testing.T) {
	d, srv := newTestDriver(t)
	conv, err := d.Start(context.Background(), ports.ChatStartConfig{WorkspacePath: testWorkspacePath})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conv.Close() }()
	srv.mu.Lock()
	delete(srv.responses, "thread/backgroundTerminals/clean")
	srv.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	interrupted := make(chan error, 1)
	go func() { interrupted <- conv.Interrupt(ctx, "turn-1") }()
	cleanup := srv.awaitFrame(func(f frame) bool { return f.Method == "thread/backgroundTerminals/clean" })
	var params map[string]string
	if err := json.Unmarshal(cleanup.Params, &params); err != nil || len(params) != 1 || params["threadId"] != "thread-1" {
		t.Fatalf("cleanup target = %s, error=%v", cleanup.Params, err)
	}
	sent := make(chan error, 1)
	go func() {
		_, err := conv.SendTurn(ctx, ports.ChatUserMessage{Text: "after stop"})
		sent <- err
	}()
	select {
	case err := <-sent:
		t.Fatalf("next turn escaped pending cleanup: %v", err)
	case err := <-interrupted:
		t.Fatalf("interrupt returned before cleanup: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	if srv.sentMethod("turn/start") {
		t.Fatal("next turn started while cleanup was pending")
	}
	srv.push(`{"id":` + string(*cleanup.ID) + `,"result":{}}`)
	if err := <-interrupted; err != nil {
		t.Fatal(err)
	}
	if err := <-sent; err != nil {
		t.Fatal(err)
	}
}

func TestInterruptDoesNotReportSuccessWhenTerminalCleanupFails(t *testing.T) {
	for _, code := range []int{-32601, -32603} {
		d, srv := newTestDriver(t)
		conv, err := d.Start(context.Background(), ports.ChatStartConfig{WorkspacePath: testWorkspacePath})
		if err != nil {
			t.Fatal(err)
		}
		srv.replyError("thread/backgroundTerminals/clean", code, "cleanup unavailable")
		err = conv.Interrupt(context.Background(), "turn-1")
		_ = conv.Close()
		if err == nil || !strings.Contains(err.Error(), "cleanup unavailable") {
			t.Fatalf("cleanup failure was hidden: %v", err)
		}
	}
}

func TestInterruptRefusalDoesNotCleanACompletedOrDifferentTurn(t *testing.T) {
	d, srv := newTestDriver(t)
	conv, err := d.Start(context.Background(), ports.ChatStartConfig{WorkspacePath: testWorkspacePath})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conv.Close() }()
	srv.replyError("turn/interrupt", -32600, "no active turn")
	if err := conv.Interrupt(context.Background(), "finished-turn"); !errors.Is(err, ports.ErrChatNoActiveTurn) {
		t.Fatalf("interrupt = %v", err)
	}
	if srv.sentMethod("thread/backgroundTerminals/clean") {
		t.Fatal("stale interrupt cleaned terminals belonging to later work")
	}
}
