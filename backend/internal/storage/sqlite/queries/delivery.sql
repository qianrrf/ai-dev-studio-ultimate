-- name: InsertDeliveryEvidence :exec
INSERT INTO project_delivery_evidence(id, project_id, worker_id, worker_path, owner_id, revision, kind, request_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetDeliveryEvidence :one
SELECT * FROM project_delivery_evidence WHERE id = ?;

-- name: ListDeliveryEvidence :many
SELECT * FROM project_delivery_evidence WHERE worker_id = ? ORDER BY created_at, id;

-- name: CompleteDeliveryEvidence :execrows
UPDATE project_delivery_evidence SET result_json = ?, completed_at = ? WHERE id = ? AND result_json IS NULL;

-- name: CountDeliveryRework :one
SELECT count(*) FROM project_delivery_evidence WHERE worker_id = ? AND kind = 'rework';

-- name: GetCompletedReviewerMessage :one
SELECT m.id, m.text FROM conversation_messages m
JOIN conversation_turns t ON t.id = m.turn_id
WHERE t.handled_by_session_id = ? AND t.state = 'completed' AND t.rolled_back_at IS NULL
AND m.role = 'assistant' AND m.streaming = 0 AND length(trim(m.text)) > 0
AND NOT EXISTS (SELECT 1 FROM conversation_turns pending WHERE pending.handled_by_session_id = t.handled_by_session_id AND pending.state IN ('queued', 'running'))
ORDER BY m.created_at DESC, m.sequence DESC LIMIT 1;

-- name: WorkerDeliveryBusy :one
SELECT EXISTS (SELECT 1 FROM conversation_turns WHERE handled_by_session_id = sqlc.arg(worker_id) AND state IN ('queued', 'running'))
OR EXISTS (SELECT 1 FROM project_delivery_evidence WHERE worker_id = sqlc.arg(worker_id) AND result_json IS NULL) AS busy;

-- name: ReworkForVerdictExists :one
SELECT EXISTS (SELECT 1 FROM project_delivery_evidence WHERE worker_id = sqlc.arg(worker_id) AND kind = 'rework' AND json_extract(request_json, '$.reviewId') = sqlc.arg(review_id));
