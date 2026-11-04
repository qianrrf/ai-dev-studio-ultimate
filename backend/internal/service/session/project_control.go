package session

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/apierr"
)

type projectControlStore interface {
	GetProjectControl(context.Context, domain.ProjectID) (domain.ProjectControl, bool, error)
	SetProjectControl(context.Context, domain.ProjectID, string, string, int64, time.Time) (domain.ProjectControl, error)
	AddProjectRequest(context.Context, domain.ProjectRequest) (domain.ProjectRequest, error)
	ListProjectRequests(context.Context, domain.ProjectID) ([]domain.ProjectRequest, error)
}

func (s *Service) ProjectControl(ctx context.Context, id domain.ProjectID) (domain.ProjectControl, error) {
	if _, err := s.requireProject(ctx, id); err != nil {
		return domain.ProjectControl{}, err
	}
	store, ok := s.store.(projectControlStore)
	if !ok {
		return domain.ProjectControl{}, fmt.Errorf("project control storage unavailable")
	}
	c, found, err := store.GetProjectControl(ctx, id)
	if err != nil {
		return c, err
	}
	if !found {
		return domain.ProjectControl{ProjectID: id, OwnerKind: "unowned"}, nil
	}
	return c, nil
}

func (s *Service) SetProjectControl(ctx context.Context, id domain.ProjectID, kind, owner string, expected int64) (domain.ProjectControl, error) {
	if _, err := s.requireProject(ctx, id); err != nil {
		return domain.ProjectControl{}, err
	}
	store, ok := s.store.(projectControlStore)
	if !ok {
		return domain.ProjectControl{}, fmt.Errorf("project control storage unavailable")
	}
	c, err := store.SetProjectControl(ctx, id, kind, owner, expected, s.now())
	if err != nil {
		return c, toSpawnAPIError(err)
	}
	if c.OwnerKind == "internal" {
		s.notifyProjectController(ctx, c, "handover")
	}
	return c, nil
}

func (s *Service) SubmitProjectRequest(ctx context.Context, in domain.ProjectRequest) (domain.ProjectRequest, error) {
	project, err := s.requireProject(ctx, in.ProjectID)
	if err != nil {
		return domain.ProjectRequest{}, err
	}
	if project.Kind.WithDefault() != domain.ProjectKindSingleRepo {
		return domain.ProjectRequest{}, apierr.Invalid("SINGLE_REPOSITORY_REQUIRED", "Codex control supports one local repository per project", nil)
	}
	store, ok := s.store.(projectControlStore)
	if !ok {
		return domain.ProjectRequest{}, fmt.Errorf("project control storage unavailable")
	}
	in.CreatedAt = s.now()
	request, err := store.AddProjectRequest(ctx, in)
	if err != nil {
		return request, toSpawnAPIError(err)
	}
	// Intake is durable even while the external controller is offline. Controllers
	// reconcile the inbox on takeover; submission never launches a worker itself.
	if control, found, err := store.GetProjectControl(ctx, in.ProjectID); err == nil && found && control.OwnerKind == "internal" {
		s.notifyProjectController(ctx, control, request.ID)
	}
	return request, nil
}

func (s *Service) notifyProjectController(ctx context.Context, control domain.ProjectControl, requestID string) {
	sender, ok := s.manager.(interface {
		SendSemantic(context.Context, domain.SessionID, string, string) error
	})
	if !ok {
		return
	}
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s", control.ProjectID, control.Revision, requestID)))
	message := fmt.Sprintf("You are the current Codex controller for project %s (owner %s, revision %d). Reconcile the durable requirements with `ao project inbox %s`, sessions and reports. Decide task decomposition, dispatch, independent Codex review, rework (maximum two business rounds) and acceptance. Use stable request IDs with --controller %s --control-revision %d --request-id on every worker dispatch. Preserve manual approvals. Do not duplicate uncertain executions. User requirement notification: %s. Attachments remain in the project inbox API.", control.ProjectID, control.OwnerID, control.Revision, control.ProjectID, control.OwnerID, control.Revision, requestID)
	if err := sender.SendSemantic(ctx, domain.SessionID(control.OwnerID), message, fmt.Sprintf("project-intake-%x", digest)); err != nil && s.logger != nil {
		s.logger.Warn("controller notification deferred; requirement retained", "projectId", control.ProjectID, "err", err)
	}
}

func (s *Service) ProjectRequests(ctx context.Context, id domain.ProjectID) ([]domain.ProjectRequest, error) {
	if _, err := s.requireProject(ctx, id); err != nil {
		return nil, err
	}
	store, ok := s.store.(projectControlStore)
	if !ok {
		return nil, fmt.Errorf("project control storage unavailable")
	}
	return store.ListProjectRequests(ctx, id)
}

func (s *Service) checkProjectClaim(ctx context.Context, id domain.ProjectID, claim *domain.ControlClaim) (bool, error) {
	store, ok := s.store.(projectControlStore)
	if !ok || id == "" {
		return false, nil
	}
	c, managed, err := store.GetProjectControl(ctx, id)
	if err != nil || !managed {
		return managed, err
	}
	if claim == nil || c.OwnerID == "" || c.OwnerID != claim.OwnerID || c.Revision != claim.Revision {
		return true, toSpawnAPIError(domain.ErrProjectControlConflict)
	}
	return true, nil
}

func (s *Service) isControlledProject(ctx context.Context, id domain.ProjectID) (bool, error) {
	store, ok := s.store.(projectControlStore)
	if !ok || id == "" {
		return false, nil
	}
	_, managed, err := store.GetProjectControl(ctx, id)
	return managed, err
}
