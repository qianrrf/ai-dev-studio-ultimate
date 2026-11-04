package session

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/process"
	"github.com/aoagents/agent-orchestrator/backend/internal/storage/sqlite/sqlitetest"
)

func deliveryGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := process.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %q: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestDeliveryExecutesTestsOnceAndInvalidatesChangedCommit(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	deliveryGit(t, repo, "init")
	deliveryGit(t, repo, "config", "user.name", "Synthetic Test")
	deliveryGit(t, repo, "config", "user.email", "synthetic@example.invalid")
	if err := os.WriteFile(filepath.Join(repo, "code.txt"), []byte("synthetic\n"), 0600); err != nil {
		t.Fatal(err)
	}
	deliveryGit(t, repo, "add", ".")
	deliveryGit(t, repo, "commit", "-m", "seed")
	head := deliveryGit(t, repo, "rev-parse", "HEAD")
	store := sqlitetest.MustOpen(t)
	if err := store.UpsertProject(ctx, domain.ProjectRecord{ID: "delivery", Path: repo, RegisteredAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	control, err := store.SetProjectControl(ctx, "delivery", "external", "desktop", 0, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	worker, _, err := store.CreateClientRequestSession(ctx, domain.SessionRecord{ProjectID: "delivery", Kind: domain.KindWorker, Harness: domain.HarnessDeepSeek, Mode: domain.SessionModeChat, Control: &domain.ControlClaim{OwnerID: "desktop", Revision: control.Revision}, ClientRequestID: "worker-seed", ClientRequestHash: "worker-hash", Activity: domain.Activity{State: domain.ActivityIdle}, Metadata: domain.SessionMetadata{WorkspacePath: repo}, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CommitClientRequestSession(ctx, worker.ID); err != nil {
		t.Fatal(err)
	}
	svc := NewWithDeps(Deps{Store: store})
	command := []string{"sh", "-c", "printf observed; exit 7"}
	if runtime.GOOS == "windows" {
		command = []string{"cmd.exe", "/d", "/c", "echo observed & exit /b 7"}
	}
	in := domain.DeliveryRequest{ID: "failed-test", Control: domain.ControlClaim{OwnerID: "desktop", Revision: control.Revision}, Kind: "test", BaseSHA: head, HeadSHA: head, Command: command}
	result, err := svc.RunDelivery(ctx, worker.ID, in)
	if err != nil || result.Result == nil || result.Result.ExitCode == nil || *result.Result.ExitCode != 7 || !strings.Contains(result.Result.Output, "observed") {
		t.Fatalf("actual failure: %+v %v", result, err)
	}
	again, err := svc.RunDelivery(ctx, worker.ID, in)
	if err != nil || !again.CreatedAt.Equal(result.CreatedAt) || again.Result == nil || *again.Result.ExitCode != 7 {
		t.Fatalf("replay: %+v %v", again, err)
	}
	review := in
	review.ID, review.Kind, review.TestID, review.Command = "bad-review", "review", in.ID, nil
	if _, err := svc.RunDelivery(ctx, worker.ID, review); err == nil {
		t.Fatal("failed actual test accepted as review evidence")
	}
	if err := os.WriteFile(filepath.Join(repo, "uncommitted.txt"), []byte("dirty"), 0600); err != nil {
		t.Fatal(err)
	}
	in.ID = "dirty-test"
	if _, err := svc.RunDelivery(ctx, worker.ID, in); err == nil {
		t.Fatal("uncommitted worker accepted")
	}
	deliveryGit(t, repo, "add", ".")
	deliveryGit(t, repo, "commit", "-m", "changed head")
	if _, err := svc.RunDelivery(ctx, worker.ID, in); err == nil {
		t.Fatal("old head accepted")
	}
	evidence, err := svc.DeliveryEvidence(ctx, worker.ID)
	if err != nil || len(evidence) != 1 || evidence[0].Current {
		t.Fatalf("stale evidence: %+v %v", evidence, err)
	}
}

func TestDeliveryOutputAndCancellationAreBounded(t *testing.T) {
	var out deliveryOutput
	data := []byte(strings.Repeat("x", 2<<20))
	if n, err := out.Write(data); err != nil || n != len(data) || !out.truncated || len(out.text) != 1<<20 {
		t.Fatalf("bounded output: %d %v", n, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := runDeliveryTest(ctx, t.TempDir(), []string{"git", "--version"})
	if result.Error == "" || (result.ExitCode != nil && *result.ExitCode == 0) {
		t.Fatalf("cancelled test passed: %+v", result)
	}
}

func TestDeliveryAcceptanceUsesIndependentCompletedVerdictAndInvalidatesReviewChanges(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	deliveryGit(t, repo, "init")
	deliveryGit(t, repo, "config", "user.name", "Synthetic")
	deliveryGit(t, repo, "config", "user.email", "synthetic@example.invalid")
	if err := os.WriteFile(filepath.Join(repo, "code.txt"), []byte("committed\n"), 0600); err != nil {
		t.Fatal(err)
	}
	deliveryGit(t, repo, "add", ".")
	deliveryGit(t, repo, "commit", "-m", "seed")
	head := deliveryGit(t, repo, "rev-parse", "HEAD")
	reviewPath := filepath.Join(t.TempDir(), "independent-review")
	deliveryGit(t, repo, "worktree", "add", "--detach", reviewPath, head)
	st := sqlitetest.MustOpen(t)
	if err := st.UpsertProject(ctx, domain.ProjectRecord{ID: "delivery", Path: repo, RegisteredAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.SetProjectControl(ctx, "delivery", "external", "desktop", 0, time.Now()); err != nil {
		t.Fatal(err)
	}
	claim := domain.ControlClaim{OwnerID: "desktop", Revision: 1}
	seed := func(request, path string, readOnly bool) domain.SessionRecord {
		t.Helper()
		harness := domain.HarnessDeepSeek
		if readOnly {
			harness = domain.HarnessCodex
		}
		row, _, err := st.CreateClientRequestSession(ctx, domain.SessionRecord{ProjectID: "delivery", Kind: domain.KindWorker, Harness: harness, ReadOnly: readOnly, Mode: domain.SessionModeChat, Control: &claim, ClientRequestID: request, ClientRequestHash: request, Activity: domain.Activity{State: domain.ActivityIdle}, Metadata: domain.SessionMetadata{WorkspacePath: path}, CreatedAt: time.Now(), UpdatedAt: time.Now()})
		if err != nil {
			t.Fatal(err)
		}
		if err := st.CommitClientRequestSession(ctx, row.ID); err != nil {
			t.Fatal(err)
		}
		return row
	}
	worker, reviewer := seed("worker", repo, false), seed("reviewer", reviewPath, true)
	svc := NewWithDeps(Deps{Store: st})
	in := domain.DeliveryRequest{ID: "test", Control: claim, Kind: "test", BaseSHA: head, HeadSHA: head, Command: []string{"git", "status", "--porcelain"}}
	if row, err := svc.RunDelivery(ctx, worker.ID, in); err != nil || row.Result == nil || row.Result.ExitCode == nil || *row.Result.ExitCode != 0 {
		t.Fatalf("test: %+v %v", row, err)
	}
	review := in
	review.ID, review.Kind, review.Command, review.TestID = "review", "review", nil, "test"
	if _, _, err := st.ReserveDeliveryEvidence(ctx, domain.DeliveryEvidence{ID: review.ID, ProjectID: "delivery", WorkerID: worker.ID, WorkerPath: repo, Request: review, CreatedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	if err := st.CompleteDeliveryEvidence(ctx, "review", domain.DeliveryResult{ReviewerID: reviewer.ID, ReviewerPath: reviewPath}, time.Now()); err != nil {
		t.Fatal(err)
	}
	conversation, err := st.CreateConversation(ctx, "review-conversation", domain.ConversationScopeSession, "delivery", reviewer.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if err := st.ClaimChatControllerGeneration(ctx, reviewer.ID, "review-generation"); err != nil {
		t.Fatal(err)
	}
	if err := st.AdoptProviderTurn(ctx, conversation.ID, reviewer.ID, "review-generation", "review-turn", "provider-review", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := st.AppendAssistantDelta(ctx, conversation.ID, "final", "provider-review", "APPROVED\nIndependent inspection", "review-message", time.Now()); err != nil {
		t.Fatal(err)
	}
	verdict := review
	verdict.ID, verdict.Kind, verdict.TestID, verdict.ReviewID = "verdict-running", "verdict", "", "review"
	row, err := svc.RunDelivery(ctx, worker.ID, verdict)
	if err == nil || row.Result == nil || row.Result.Error == "" {
		t.Fatalf("unfinished verdict accepted: %+v %v", row, err)
	}
	if err := st.SettleTurn(ctx, conversation.ID, "provider-review", domain.TurnStateCompleted, "", time.Now()); err != nil {
		t.Fatal(err)
	}
	verdict.ID = "verdict"
	row, err = svc.RunDelivery(ctx, worker.ID, verdict)
	if err != nil || row.Result.Verdict != "approved" || row.Result.MessageID != "review-message" {
		t.Fatalf("independent verdict: %+v %v", row, err)
	}
	accept := verdict
	accept.ID, accept.Kind, accept.ReviewID, accept.TestID, accept.Message = "accept-bad", "accept", "verdict", "wrong-test", "Controller accepts"
	if _, err := svc.RunDelivery(ctx, worker.ID, accept); err == nil {
		t.Fatal("acceptance with mismatched test succeeded")
	}
	accept.ID, accept.TestID = "accept", "test"
	row, err = svc.RunDelivery(ctx, worker.ID, accept)
	if err != nil || row.Result.Verdict != "accepted" {
		t.Fatalf("acceptance: %+v %v", row, err)
	}
	if err := os.WriteFile(filepath.Join(reviewPath, "uncommitted.txt"), []byte("changed review"), 0600); err != nil {
		t.Fatal(err)
	}
	rows, err := svc.DeliveryEvidence(ctx, worker.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, evidence := range rows {
		if evidence.Request.Kind != "test" && evidence.Current {
			t.Fatalf("changed reviewer retained current evidence: %+v", evidence)
		}
	}
	accept.ID = "accept-changed-review"
	if _, err := svc.RunDelivery(ctx, worker.ID, accept); err == nil {
		t.Fatal("changed reviewer accepted")
	}
}
