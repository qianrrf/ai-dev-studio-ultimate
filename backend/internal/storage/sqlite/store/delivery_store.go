package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/storage/sqlite/gen"
)

func deliveryRecord(row gen.ProjectDeliveryEvidence) (domain.DeliveryEvidence, error) {
	out := domain.DeliveryEvidence{ID: row.ID, ProjectID: domain.ProjectID(row.ProjectID), WorkerID: domain.SessionID(row.WorkerID), WorkerPath: row.WorkerPath, CreatedAt: row.CreatedAt}
	if err := json.Unmarshal([]byte(row.RequestJson), &out.Request); err != nil {
		return out, err
	}
	if row.ResultJson.Valid {
		out.Result = &domain.DeliveryResult{}
		if err := json.Unmarshal([]byte(row.ResultJson.String), out.Result); err != nil {
			return out, err
		}
	}
	if row.CompletedAt.Valid {
		out.CompletedAt = &row.CompletedAt.Time
	}
	return out, nil
}

func (s *Store) ListDeliveryEvidence(ctx context.Context, worker domain.SessionID) ([]domain.DeliveryEvidence, error) {
	rows, err := s.qr.ListDeliveryEvidence(ctx, string(worker))
	if err != nil {
		return nil, err
	}
	out := make([]domain.DeliveryEvidence, 0, len(rows))
	for _, row := range rows {
		rec, err := deliveryRecord(row)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

func (s *Store) GetDeliveryEvidence(ctx context.Context, id string) (domain.DeliveryEvidence, error) {
	row, err := s.qr.GetDeliveryEvidence(ctx, id)
	if err != nil {
		return domain.DeliveryEvidence{}, err
	}
	return deliveryRecord(row)
}

// ReserveDeliveryEvidence fences ownership and registers side effects in the
// same writer transaction. An incomplete receipt is reconciled, never rerun.
func (s *Store) ReserveDeliveryEvidence(ctx context.Context, in domain.DeliveryEvidence) (domain.DeliveryEvidence, bool, error) {
	encoded, err := json.Marshal(in.Request)
	if err != nil {
		return in, false, err
	}
	if err := s.writeMu.LockContext(ctx); err != nil {
		return in, false, err
	}
	defer s.writeMu.Unlock()
	var out domain.DeliveryEvidence
	fresh := false
	err = s.inTx(ctx, "reserve delivery evidence", func(q *gen.Queries) error {
		control, err := q.GetProjectControl(ctx, string(in.ProjectID))
		if err != nil {
			return err
		}
		if control.OwnerID == "" || control.OwnerID != in.Request.Control.OwnerID || control.Revision != in.Request.Control.Revision {
			return domain.ErrProjectControlConflict
		}
		prior, err := q.GetDeliveryEvidence(ctx, in.ID)
		if err == nil {
			if prior.ProjectID != string(in.ProjectID) || prior.WorkerID != string(in.WorkerID) || prior.WorkerPath != in.WorkerPath || prior.RequestJson != string(encoded) {
				return domain.ErrProjectControlConflict
			}
			out, err = deliveryRecord(prior)
			return err
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		worker, err := q.GetSession(ctx, in.WorkerID)
		if err != nil {
			return err
		}
		if worker.ProjectID == nil || *worker.ProjectID != in.ProjectID || worker.Kind != domain.KindWorker || worker.ReadOnly || worker.Harness != domain.HarnessDeepSeek || worker.WorkspacePath != in.WorkerPath || worker.IsTerminated {
			return domain.ErrProjectControlConflict
		}
		busy, err := q.WorkerDeliveryBusy(ctx, in.WorkerID)
		if err != nil {
			return err
		}
		if !busy.Valid || busy.Bool || worker.ActivityState == domain.ActivityActive || worker.ActivityState == domain.ActivityWaitingInput || worker.ActivityState == domain.ActivityBlocked || worker.ProvisionState.IsProvisioning() {
			return domain.ErrProjectControlConflict
		}
		load := func(id, kind string) (domain.DeliveryEvidence, error) {
			row, err := q.GetDeliveryEvidence(ctx, id)
			if err != nil {
				return domain.DeliveryEvidence{}, err
			}
			rec, err := deliveryRecord(row)
			if err != nil {
				return rec, err
			}
			if rec.WorkerID != in.WorkerID || rec.WorkerPath != in.WorkerPath || rec.Request.Kind != kind || rec.Request.BaseSHA != in.Request.BaseSHA || rec.Request.HeadSHA != in.Request.HeadSHA || rec.Result == nil || rec.Result.Error != "" {
				return rec, domain.ErrProjectControlConflict
			}
			return rec, nil
		}
		if in.Request.Kind == "review" || in.Request.Kind == "accept" {
			test, err := load(in.Request.TestID, "test")
			if err != nil {
				return err
			}
			if test.Result.ExitCode == nil || *test.Result.ExitCode != 0 {
				return domain.ErrProjectControlConflict
			}
		}
		if in.Request.Kind == "verdict" {
			if _, err := load(in.Request.ReviewID, "review"); err != nil {
				return err
			}
		}
		if in.Request.Kind == "accept" || in.Request.Kind == "rework" {
			verdict, err := load(in.Request.ReviewID, "verdict")
			if err != nil {
				return err
			}
			want := "approved"
			if in.Request.Kind == "rework" {
				want = "changes_requested"
			}
			if verdict.Result.Verdict != want {
				return domain.ErrProjectControlConflict
			}
			if in.Request.Kind == "accept" {
				review, err := load(verdict.Request.ReviewID, "review")
				if err != nil {
					return err
				}
				if review.Request.TestID != in.Request.TestID {
					return domain.ErrProjectControlConflict
				}
			}
		}
		if in.Request.Kind == "rework" {
			duplicate, err := q.ReworkForVerdictExists(ctx, gen.ReworkForVerdictExistsParams{WorkerID: string(in.WorkerID), ReviewID: in.Request.ReviewID})
			if err != nil {
				return err
			}
			if duplicate {
				return domain.ErrProjectControlConflict
			}
			n, err := q.CountDeliveryRework(ctx, string(in.WorkerID))
			if err != nil {
				return err
			}
			if n >= 2 {
				return fmt.Errorf("two business rework rounds exhausted; refer to user: %w", domain.ErrProjectControlConflict)
			}
			n, err = q.CountProjectExecutingWorkers(ctx, optionalProjectID(in.ProjectID))
			if err != nil {
				return err
			}
			if n >= 2 {
				return domain.ErrProjectWorkerCapacity
			}
		}
		if err := q.InsertDeliveryEvidence(ctx, gen.InsertDeliveryEvidenceParams{ID: in.ID, ProjectID: string(in.ProjectID), WorkerID: string(in.WorkerID), WorkerPath: in.WorkerPath, OwnerID: in.Request.Control.OwnerID, Revision: in.Request.Control.Revision, Kind: in.Request.Kind, RequestJson: string(encoded), CreatedAt: in.CreatedAt}); err != nil {
			return err
		}
		out, fresh = in, true
		return nil
	})
	return out, fresh, err
}

func (s *Store) CompleteDeliveryEvidence(ctx context.Context, id string, result domain.DeliveryResult, now time.Time) error {
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if err := s.writeMu.LockContext(ctx); err != nil {
		return err
	}
	defer s.writeMu.Unlock()
	n, err := s.qw.CompleteDeliveryEvidence(ctx, gen.CompleteDeliveryEvidenceParams{ID: id, ResultJson: sql.NullString{String: string(encoded), Valid: true}, CompletedAt: nullTime(now)})
	if err == nil && n != 1 {
		return domain.ErrProjectControlConflict
	}
	return err
}

func (s *Store) CompletedReviewerMessage(ctx context.Context, reviewer domain.SessionID) (string, string, error) {
	row, err := s.qr.GetCompletedReviewerMessage(ctx, reviewer)
	return row.ID, row.Text, err
}

// CheckProjectChatDelivery also covers CLI relays, direct Chat HTTP requests,
// editing and steering, so a hidden UI button is never the ownership boundary.
func (s *Store) CheckProjectChatDelivery(ctx context.Context, id domain.SessionID) error {
	return checkProjectChatDelivery(ctx, s.qr, id)
}

func checkProjectChatDelivery(ctx context.Context, q *gen.Queries, id domain.SessionID) error {
	session, err := q.GetSession(ctx, id)
	if err != nil {
		return err
	}
	if session.ProjectID == nil {
		return nil
	}
	control, err := q.GetProjectControl(ctx, string(*session.ProjectID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if session.Kind == domain.KindOrchestrator {
		if control.OwnerKind == "internal" && control.OwnerID == string(id) {
			return nil
		}
		return domain.ErrProjectControlConflict
	}
	if receipt := domain.DeliveryReceipt(ctx); receipt != "" {
		row, err := q.GetDeliveryEvidence(ctx, receipt)
		if err == nil && row.WorkerID == string(id) && row.Kind == "rework" {
			return nil
		}
	}
	return domain.ErrProjectControlConflict
}
