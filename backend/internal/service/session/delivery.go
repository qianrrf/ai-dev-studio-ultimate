package session

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/apierr"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
	"github.com/aoagents/agent-orchestrator/backend/internal/process"
)

type deliveryStore interface {
	ReserveDeliveryEvidence(context.Context, domain.DeliveryEvidence) (domain.DeliveryEvidence, bool, error)
	CompleteDeliveryEvidence(context.Context, string, domain.DeliveryResult, time.Time) error
	ListDeliveryEvidence(context.Context, domain.SessionID) ([]domain.DeliveryEvidence, error)
	GetDeliveryEvidence(context.Context, string) (domain.DeliveryEvidence, error)
	CompletedReviewerMessage(context.Context, domain.SessionID) (string, string, error)
}

var deliverySHA = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)
var deliveryID = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.:-]{0,79}$`)

func deliveryConflict(message string) error {
	return apierr.Conflict("DELIVERY_EVIDENCE_CONFLICT", message, nil)
}

// verifyDeliveryTree refuses all dirty files, including untracked files. A test
// output directory belongs in the synthetic project's .gitignore if intentional.
func verifyDeliveryTree(ctx context.Context, path, head string) error {
	if path == "" {
		return deliveryConflict("Worker workspace is missing")
	}
	cmd := process.CommandContext(ctx, "git", "-C", path, "rev-parse", "HEAD")
	actual, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(actual)) != head {
		return deliveryConflict("Workspace commit changed or cannot be read")
	}
	status := process.CommandContext(ctx, "git", "-C", path, "status", "--porcelain=v1", "--untracked-files=all")
	output, err := status.Output()
	if err != nil || len(output) != 0 {
		return deliveryConflict("Commit all worker changes before testing or review")
	}
	return nil
}

func (s *Service) DeliveryEvidence(ctx context.Context, workerID domain.SessionID) ([]domain.DeliveryEvidence, error) {
	worker, found, err := s.store.GetSession(ctx, workerID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, apierr.NotFound("SESSION_NOT_FOUND", "Unknown worker")
	}
	store, ok := s.store.(deliveryStore)
	if !ok {
		return nil, fmt.Errorf("delivery evidence storage unavailable")
	}
	rows, err := store.ListDeliveryEvidence(ctx, workerID)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Current = worker.Metadata.WorkspacePath == rows[i].WorkerPath && verifyDeliveryTree(ctx, rows[i].WorkerPath, rows[i].Request.HeadSHA) == nil
		if rows[i].Current {
			switch rows[i].Request.Kind {
			case "review":
				_, err = s.deliveryReview(ctx, store, rows[i].ID, "review")
			case "verdict", "accept":
				kind := "review"
				if rows[i].Request.Kind == "accept" {
					kind = "verdict"
				}
				_, err = s.deliveryReview(ctx, store, rows[i].Request.ReviewID, kind)
			default:
				err = nil
			}
			rows[i].Current = err == nil
		}
	}
	return rows, nil
}

// RunDelivery registers first. Cancellation/restart after registration leaves an
// observable incomplete receipt instead of risking duplicate side effects.
func (s *Service) RunDelivery(ctx context.Context, workerID domain.SessionID, in domain.DeliveryRequest) (domain.DeliveryEvidence, error) {
	if !deliveryID.MatchString(in.ID) || !deliverySHA.MatchString(in.BaseSHA) || !deliverySHA.MatchString(in.HeadSHA) || len(in.Message) > 16384 || len(in.Command) > 64 {
		return domain.DeliveryEvidence{}, apierr.Invalid("INVALID_DELIVERY_REQUEST", "Provide a stable ID, full base/head SHA and bounded evidence", nil)
	}
	switch in.Kind {
	case "test":
		if len(in.Command) == 0 || strings.TrimSpace(in.Command[0]) == "" {
			return domain.DeliveryEvidence{}, apierr.Invalid("TEST_COMMAND_REQUIRED", "Provide executable and arguments", nil)
		}
	case "review", "verdict":
	case "rework", "accept":
		if strings.TrimSpace(in.Message) == "" {
			return domain.DeliveryEvidence{}, apierr.Invalid("CONCLUSION_REQUIRED", "Provide rework instructions or an acceptance conclusion", nil)
		}
	default:
		return domain.DeliveryEvidence{}, apierr.Invalid("INVALID_DELIVERY_KIND", "Unknown delivery operation", nil)
	}
	for _, arg := range in.Command {
		if len(arg) > 16384 || strings.ContainsRune(arg, 0) {
			return domain.DeliveryEvidence{}, apierr.Invalid("INVALID_TEST_COMMAND", "Invalid command argument", nil)
		}
	}
	worker, found, err := s.store.GetSession(ctx, workerID)
	if err != nil {
		return domain.DeliveryEvidence{}, err
	}
	if !found {
		return domain.DeliveryEvidence{}, apierr.NotFound("SESSION_NOT_FOUND", "Unknown worker")
	}
	managed, err := s.checkProjectClaim(ctx, worker.ProjectID, &in.Control)
	if err != nil {
		return domain.DeliveryEvidence{}, err
	}
	if !managed || worker.Kind != domain.KindWorker || worker.ReadOnly || worker.Harness != domain.HarnessDeepSeek {
		return domain.DeliveryEvidence{}, deliveryConflict("Delivery requires a controlled DSH worker")
	}
	store, ok := s.store.(deliveryStore)
	if !ok {
		return domain.DeliveryEvidence{}, fmt.Errorf("delivery evidence storage unavailable")
	}
	if err := verifyDeliveryTree(ctx, worker.Metadata.WorkspacePath, in.HeadSHA); err != nil {
		return domain.DeliveryEvidence{}, err
	}
	base := process.CommandContext(ctx, "git", "-C", worker.Metadata.WorkspacePath, "merge-base", "--is-ancestor", in.BaseSHA, in.HeadSHA)
	if err := base.Run(); err != nil {
		return domain.DeliveryEvidence{}, deliveryConflict("Base commit must be an ancestor of the worker head")
	}
	if worker.Activity.State == domain.ActivityActive || worker.Activity.State == domain.ActivityWaitingInput || worker.Activity.State == domain.ActivityBlocked || worker.ProvisionState.IsProvisioning() {
		return domain.DeliveryEvidence{}, deliveryConflict("Wait for the worker to finish or resolve its approval")
	}
	if in.Kind == "rework" || in.Kind == "accept" {
		if _, err := s.deliveryReview(ctx, store, in.ReviewID, "verdict"); err != nil {
			return domain.DeliveryEvidence{}, err
		}
	}
	if in.Kind == "verdict" {
		if _, err := s.deliveryReview(ctx, store, in.ReviewID, "review"); err != nil {
			return domain.DeliveryEvidence{}, err
		}
	}
	evidence, fresh, err := store.ReserveDeliveryEvidence(ctx, domain.DeliveryEvidence{ID: in.ID, ProjectID: worker.ProjectID, WorkerID: workerID, WorkerPath: worker.Metadata.WorkspacePath, Request: in, CreatedAt: s.now()})
	if err != nil {
		return evidence, toSpawnAPIError(err)
	}
	if !fresh {
		return evidence, nil
	}
	result := domain.DeliveryResult{}
	switch in.Kind {
	case "test":
		result = runDeliveryTest(ctx, evidence.WorkerPath, in.Command)
		if err := verifyDeliveryTree(ctx, evidence.WorkerPath, in.HeadSHA); err != nil {
			result.Error = err.Error()
		}
	case "review":
		result, err = s.startDeliveryReview(ctx, store, worker, in)
	case "verdict":
		review, readErr := store.GetDeliveryEvidence(ctx, in.ReviewID)
		if readErr != nil {
			err = readErr
			break
		}
		result.MessageID, result.Output, err = store.CompletedReviewerMessage(ctx, review.Result.ReviewerID)
		if err == nil {
			first, _, _ := strings.Cut(strings.TrimSpace(result.Output), "\n")
			switch strings.TrimSpace(first) {
			case "APPROVED":
				result.Verdict = "approved"
			case "CHANGES_REQUESTED":
				result.Verdict = "changes_requested"
			default:
				err = deliveryConflict("Reviewer did not produce an explicit verdict")
			}
		}
	case "rework":
		sender, ok := s.manager.(interface {
			SendSemantic(context.Context, domain.SessionID, string, string) error
		})
		if !ok {
			err = fmt.Errorf("semantic delivery unavailable")
			break
		}
		err = sender.SendSemantic(domain.WithDeliveryReceipt(ctx, in.ID), workerID, in.Message, "delivery:"+in.ID)
		result.Delivered = err == nil
	case "accept":
		result.Verdict, result.Output = "accepted", in.Message
	}
	if err != nil {
		result.Error = err.Error()
	}
	// Preserve the result even if the caller disconnects after execution.
	persistCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if persistErr := store.CompleteDeliveryEvidence(persistCtx, in.ID, result, s.now()); persistErr != nil {
		return evidence, persistErr
	}
	evidence, readErr := store.GetDeliveryEvidence(persistCtx, in.ID)
	if readErr != nil {
		return evidence, readErr
	}
	return evidence, err
}

func (s *Service) deliveryReview(ctx context.Context, store deliveryStore, id, kind string) (domain.DeliveryEvidence, error) {
	evidence, err := store.GetDeliveryEvidence(ctx, id)
	if err != nil || evidence.Request.Kind != kind || evidence.Result == nil || evidence.Result.Error != "" {
		return evidence, deliveryConflict("Completed independent review evidence is required")
	}
	if kind == "verdict" {
		return s.deliveryReview(ctx, store, evidence.Request.ReviewID, "review")
	}
	if evidence.Result.ReviewerID == "" || evidence.Result.ReviewerPath == "" {
		return evidence, deliveryConflict("Reviewer workspace is missing")
	}
	reviewer, found, err := s.store.GetSession(ctx, evidence.Result.ReviewerID)
	if err != nil || !found || !reviewer.ReadOnly || reviewer.Harness != domain.HarnessCodex || reviewer.Mode != domain.SessionModeChat || reviewer.Metadata.WorkspacePath != evidence.Result.ReviewerPath || reviewer.Activity.State == domain.ActivityActive || reviewer.Activity.State == domain.ActivityWaitingInput || reviewer.Activity.State == domain.ActivityBlocked || reviewer.ProvisionState.IsProvisioning() {
		return evidence, deliveryConflict("Independent Codex reviewer is incomplete or changed")
	}
	if err := verifyDeliveryTree(ctx, evidence.Result.ReviewerPath, evidence.Request.HeadSHA); err != nil {
		return evidence, err
	}
	return evidence, nil
}

func (s *Service) startDeliveryReview(ctx context.Context, store deliveryStore, worker domain.SessionRecord, in domain.DeliveryRequest) (domain.DeliveryResult, error) {
	project, err := s.requireProject(ctx, worker.ProjectID)
	if err != nil {
		return domain.DeliveryResult{}, err
	}
	if len(project.Config.Reviewers) != 1 || project.Config.Reviewers[0].Harness != domain.ReviewerCodex {
		return domain.DeliveryResult{}, deliveryConflict("Configure exactly one Codex reviewer")
	}
	tuning := project.Config.Reviewers[0].AgentConfig
	test, err := store.GetDeliveryEvidence(ctx, in.TestID)
	if err != nil {
		return domain.DeliveryResult{}, err
	}
	output := test.Result.Output
	if len(output) > 16384 {
		output = output[:16384] + "\n[Preview truncated; full output is retained in delivery evidence.]"
	}
	prompt := fmt.Sprintf("Independently review the committed diff %s..%s for worker %s. This is a fresh read-only context. Inspect actual code and tests; do not edit files, use external tools, spawn agents, or request elevation. The first line of your final answer must be exactly APPROVED or CHANGES_REQUESTED, followed by concrete findings. Test evidence: command %q; exit code %d; result %s. Controller brief: %s", in.BaseSHA, in.HeadSHA, worker.ID, test.Request.Command, *test.Result.ExitCode, output, in.Message)
	hash := sha256.Sum256([]byte(prompt))
	reviewer, _, _, err := s.Spawn(ctx, ports.SpawnConfig{ProjectID: worker.ProjectID, Kind: domain.KindWorker, Harness: domain.HarnessCodex, ReadOnly: true, ReviewCommit: in.HeadSHA, RequestedMode: domain.SessionModeChat, RequireNewBranch: true, Control: &in.Control, ClientRequestID: "review:" + in.ID, ClientRequestHash: fmt.Sprintf("%x", hash), Prompt: prompt, AgentConfig: ports.AgentConfig{Model: tuning.Model, Effort: tuning.Effort, Permissions: domain.PermissionModeDefault}})
	if err != nil {
		return domain.DeliveryResult{}, err
	}
	rec, found, err := s.store.GetSession(ctx, reviewer.ID)
	if err != nil || !found {
		return domain.DeliveryResult{}, deliveryConflict("Reviewer session cannot be read")
	}
	result := domain.DeliveryResult{ReviewerID: reviewer.ID, ReviewerPath: rec.Metadata.WorkspacePath}
	if err := verifyDeliveryTree(ctx, rec.Metadata.WorkspacePath, in.HeadSHA); err != nil {
		return result, err
	}
	return result, nil
}

type deliveryOutput struct {
	mu        sync.Mutex
	text      []byte
	truncated bool
}

func (b *deliveryOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(p)
	remaining := (1 << 20) - len(b.text)
	if len(p) > remaining {
		p = p[:remaining]
		b.truncated = true
	}
	b.text = append(b.text, p...)
	return n, nil
}

func runDeliveryTest(ctx context.Context, path string, command []string) domain.DeliveryResult {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd := process.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = path
	process.ConfigureTreeCancellation(cmd)
	var output deliveryOutput
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	result := domain.DeliveryResult{Output: string(output.text), Truncated: output.truncated}
	if cmd.ProcessState != nil {
		code := cmd.ProcessState.ExitCode()
		result.ExitCode = &code
	}
	if err != nil {
		if _, exited := err.(*exec.ExitError); !exited || ctx.Err() != nil {
			result.Error = err.Error()
		}
	}
	return result
}
