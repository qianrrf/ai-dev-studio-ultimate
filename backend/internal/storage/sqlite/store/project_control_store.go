package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/storage/sqlite/gen"
)

func projectControlRecord(row gen.ProjectControl) domain.ProjectControl {
	return domain.ProjectControl{ProjectID: domain.ProjectID(row.ProjectID), OwnerKind: row.OwnerKind, OwnerID: row.OwnerID, Revision: row.Revision, UpdatedAt: row.UpdatedAt}
}

// GetProjectControl distinguishes an unmanaged project from released ownership.
func (s *Store) GetProjectControl(ctx context.Context, id domain.ProjectID) (domain.ProjectControl, bool, error) {
	row, err := s.qr.GetProjectControl(ctx, string(id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ProjectControl{}, false, nil
	}
	return projectControlRecord(row), err == nil, err
}

// SetProjectControl serializes CAS and creation with the session seed writer.
func (s *Store) SetProjectControl(ctx context.Context, id domain.ProjectID, kind, owner string, expected int64, now time.Time) (domain.ProjectControl, error) {
	if expected < 0 || (kind != "internal" && kind != "external" && kind != "unowned") || len(owner) > 128 || strings.TrimSpace(owner) != owner || strings.IndexFunc(owner, unicode.IsControl) >= 0 || (kind == "unowned") != (owner == "") {
		return domain.ProjectControl{}, fmt.Errorf("invalid project controller")
	}
	if err := s.writeMu.LockContext(ctx); err != nil {
		return domain.ProjectControl{}, err
	}
	defer s.writeMu.Unlock()
	var out domain.ProjectControl
	err := s.inTx(ctx, "set project control", func(q *gen.Queries) error {
		project, err := q.GetProject(ctx, id)
		if err != nil {
			return err
		}
		if domain.ProjectKind(project.Kind).WithDefault() != domain.ProjectKindSingleRepo {
			return fmt.Errorf("Codex control requires a single-repository project")
		}
		config := unmarshalProjectConfig(project.Config)
		config.Worker.Harness = domain.HarnessDeepSeek
		config.Orchestrator.Harness = domain.HarnessCodex
		config.Reviewers = []domain.ReviewerConfig{{Harness: domain.ReviewerCodex}}
		config.AutoReview, config.WorkersRequestReview = false, false
		if config.AgentConfig.Permissions == "" {
			config.AgentConfig.Permissions = domain.PermissionModeDefault
		}
		encoded, err := marshalProjectConfig(config)
		if err != nil {
			return err
		}
		if _, err := q.UpdateProjectSettings(ctx, gen.UpdateProjectSettingsParams{ID: id, DisplayName: project.DisplayName, Config: encoded}); err != nil {
			return err
		}
		if kind == "internal" {
			session, err := q.GetSession(ctx, domain.SessionID(owner))
			if err != nil || session.ProjectID == nil || *session.ProjectID != id || session.Kind != domain.KindOrchestrator || session.Harness != domain.HarnessCodex || session.SessionMode != domain.SessionModeChat || session.IsTerminated {
				return domain.ErrProjectControlConflict
			}
		}
		current, err := q.GetProjectControl(ctx, string(id))
		if errors.Is(err, sql.ErrNoRows) {
			if expected != 0 {
				return domain.ErrProjectControlConflict
			}
			if err := q.InsertProjectControl(ctx, gen.InsertProjectControlParams{ProjectID: string(id), OwnerKind: kind, OwnerID: owner, UpdatedAt: now}); err != nil {
				return err
			}
		} else {
			if err != nil {
				return err
			}
			if current.Revision != expected {
				return domain.ErrProjectControlConflict
			}
			n, err := q.UpdateProjectControl(ctx, gen.UpdateProjectControlParams{ProjectID: string(id), OwnerKind: kind, OwnerID: owner, Revision: expected, UpdatedAt: now})
			if err != nil {
				return err
			}
			if n != 1 {
				return domain.ErrProjectControlConflict
			}
		}
		row, err := q.GetProjectControl(ctx, string(id))
		out = projectControlRecord(row)
		return err
	})
	return out, err
}

// AddProjectRequest makes user intake idempotent independently of worker launch.
func (s *Store) AddProjectRequest(ctx context.Context, in domain.ProjectRequest) (domain.ProjectRequest, error) {
	if in.ID == "" || len(in.ID) > 128 || strings.TrimSpace(in.ID) != in.ID || strings.IndexFunc(in.ID, unicode.IsControl) >= 0 || strings.TrimSpace(in.Brief) == "" || len(in.Brief) > 16384 {
		return domain.ProjectRequest{}, fmt.Errorf("request id and a brief of at most 16384 bytes are required")
	}
	if len(in.Attachments) > 8 {
		return domain.ProjectRequest{}, fmt.Errorf("too many attachments")
	}
	total := 0
	for _, a := range in.Attachments {
		total += len(a.Data)
		if len(a.Data) > 50<<20 || total > 100<<20 {
			return domain.ProjectRequest{}, fmt.Errorf("attachments too large")
		}
	}
	if in.Attachments == nil {
		in.Attachments = []domain.RequestAttachment{}
	}
	attachments, err := json.Marshal(in.Attachments)
	if err != nil {
		return domain.ProjectRequest{}, err
	}
	if err := s.writeMu.LockContext(ctx); err != nil {
		return domain.ProjectRequest{}, err
	}
	defer s.writeMu.Unlock()
	var out domain.ProjectRequest
	err = s.inTx(ctx, "add project request", func(q *gen.Queries) error {
		row, err := q.GetProjectRequest(ctx, in.ID)
		if err == nil {
			if row.ProjectID != string(in.ProjectID) || row.Brief != in.Brief || row.Attachments != string(attachments) {
				return domain.ErrProjectControlConflict
			}
			out = domain.ProjectRequest{ID: row.ID, ProjectID: domain.ProjectID(row.ProjectID), Brief: row.Brief, CreatedAt: row.CreatedAt}
			if err := json.Unmarshal([]byte(row.Attachments), &out.Attachments); err != nil {
				return err
			}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err := q.InsertProjectRequest(ctx, gen.InsertProjectRequestParams{ID: in.ID, ProjectID: string(in.ProjectID), Brief: in.Brief, Attachments: string(attachments), CreatedAt: in.CreatedAt}); err != nil {
			return err
		}
		out = in
		return nil
	})
	return out, err
}

func (s *Store) ListProjectRequests(ctx context.Context, id domain.ProjectID) ([]domain.ProjectRequest, error) {
	rows, err := s.qr.ListProjectRequests(ctx, string(id))
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProjectRequest, 0, len(rows))
	for _, row := range rows {
		request := domain.ProjectRequest{ID: row.ID, ProjectID: domain.ProjectID(row.ProjectID), Brief: row.Brief, CreatedAt: row.CreatedAt}
		if err := json.Unmarshal([]byte(row.Attachments), &request.Attachments); err != nil {
			return nil, err
		}
		out = append(out, request)
	}
	return out, nil
}

func (s *Store) CompleteAutomationIntake(ctx context.Context, id domain.AutomationRunID, now time.Time) (bool, error) {
	if err := s.writeMu.LockContext(ctx); err != nil {
		return false, err
	}
	defer s.writeMu.Unlock()
	n, err := s.qw.CompleteAutomationIntake(ctx, gen.CompleteAutomationIntakeParams{ID: id, FinishedAt: nullTime(now), UpdatedAt: now})
	return n == 1, err
}

func validateControlledSession(ctx context.Context, q *gen.Queries, rec domain.SessionRecord) (bool, error) {
	if rec.ProjectID == "" {
		return false, nil
	}
	control, err := q.GetProjectControl(ctx, string(rec.ProjectID))
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if rec.IsTaskPreparation {
		return true, domain.ErrProjectControlConflict
	}
	if rec.Kind == domain.KindOrchestrator {
		if rec.Harness != domain.HarnessCodex || rec.Mode != domain.SessionModeChat {
			return true, domain.ErrProjectControlConflict
		}
		return true, nil
	}
	if rec.Control == nil || control.OwnerID == "" || rec.Control.OwnerID != control.OwnerID || rec.Control.Revision != control.Revision || rec.ClientRequestID == "" || rec.ClientRequestHash == "" {
		return true, domain.ErrProjectControlConflict
	}
	if (!rec.ReadOnly && rec.Harness != domain.HarnessDeepSeek) || (rec.ReadOnly && rec.Harness != domain.HarnessCodex) || rec.Mode != domain.SessionModeChat {
		return true, domain.ErrProjectControlConflict
	}
	// A reservation remains after an uncertain launch, even if rollback removes
	// its seed. Such a retry must be reconciled, never silently dispatched again.
	if dispatch, err := q.GetProjectDispatch(ctx, rec.ClientRequestID); err == nil {
		if dispatch.ProjectID != string(rec.ProjectID) {
			return true, domain.ErrProjectControlConflict
		}
		existing, err := q.GetClientRequestSession(ctx, rec.ClientRequestID)
		if err != nil || existing.ClientRequestHash != rec.ClientRequestHash {
			return true, domain.ErrProjectControlConflict
		}
		if _, err := q.GetSession(ctx, domain.SessionID(dispatch.SessionID)); err != nil {
			return true, domain.ErrProjectControlConflict
		}
		return true, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return true, err
	}
	if !rec.ReadOnly {
		n, err := q.CountProjectExecutingWorkers(ctx, optionalProjectID(rec.ProjectID))
		if err != nil {
			return true, err
		}
		if n >= 2 {
			return true, domain.ErrProjectWorkerCapacity
		}
	}
	return true, nil
}
