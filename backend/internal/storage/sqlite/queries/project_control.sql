-- name: GetProjectControl :one
SELECT project_id, owner_kind, owner_id, revision, updated_at FROM project_control WHERE project_id = ?;

-- name: InsertProjectControl :exec
INSERT INTO project_control(project_id, owner_kind, owner_id, revision, updated_at) VALUES (?, ?, ?, 1, ?);

-- name: UpdateProjectControl :execrows
UPDATE project_control SET owner_kind = ?, owner_id = ?, revision = revision + 1, updated_at = ?
WHERE project_id = ? AND revision = ?;

-- name: InsertProjectRequest :exec
INSERT INTO project_requests(id, project_id, brief, attachments, created_at) VALUES (?, ?, ?, ?, ?);

-- name: GetProjectRequest :one
SELECT id, project_id, brief, attachments, created_at FROM project_requests WHERE id = ?;

-- name: ListProjectRequests :many
SELECT id, project_id, brief, attachments, created_at FROM project_requests WHERE project_id = ? ORDER BY created_at, id;

-- name: InsertProjectDispatch :exec
INSERT INTO project_dispatches(request_id, project_id, owner_id, revision, session_id, created_at) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetProjectDispatch :one
SELECT request_id, project_id, owner_id, revision, session_id, created_at FROM project_dispatches WHERE request_id = ?;

-- name: CountProjectExecutingWorkers :one
SELECT count(*) FROM sessions
WHERE project_id = ? AND kind = 'worker' AND read_only = 0 AND is_task_preparation = 0
AND is_terminated = 0 AND (provision_state = 'provisioning' OR activity_state IN ('active', 'waiting_input', 'blocked')
OR (client_request_id <> '' AND client_request_committed = 0)
OR EXISTS (SELECT 1 FROM conversation_turns t WHERE t.handled_by_session_id = sessions.id AND t.state IN ('queued', 'running')));

-- name: GetSessionNumFloor :one
SELECT max_num FROM session_num_floors WHERE id_prefix = ?;

-- name: CompleteAutomationIntake :execrows
UPDATE automation_runs SET status = 'completed', finished_at = ?, claimed_at = NULL,
lease_expires_at = NULL, error_message = NULL, updated_at = ?
WHERE automation_runs.id = ? AND status = 'spawning'
AND EXISTS (SELECT 1 FROM project_requests WHERE project_requests.id = 'automation:' || automation_runs.id);
