-- +goose Up
CREATE TABLE project_delivery_evidence (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    worker_id TEXT NOT NULL,
    worker_path TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    revision INTEGER NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('test', 'review', 'verdict', 'rework', 'accept')),
    request_json TEXT NOT NULL CHECK (json_valid(request_json)),
    result_json TEXT CHECK (result_json IS NULL OR json_valid(result_json)),
    created_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP
);
CREATE INDEX idx_delivery_worker ON project_delivery_evidence(worker_id, created_at, id);
-- +goose StatementBegin
CREATE TRIGGER project_delivery_insert AFTER INSERT ON project_delivery_evidence BEGIN
    INSERT INTO change_log(project_id, session_id, event_type, payload, created_at)
    VALUES (NEW.project_id, NEW.worker_id, 'session_updated', json_object('deliveryEvidenceId', NEW.id), NEW.created_at);
END;
CREATE TRIGGER project_delivery_complete AFTER UPDATE OF result_json ON project_delivery_evidence BEGIN
    INSERT INTO change_log(project_id, session_id, event_type, payload, created_at)
    VALUES (NEW.project_id, NEW.worker_id, 'session_updated', json_object('deliveryEvidenceId', NEW.id), NEW.completed_at);
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS project_delivery_complete;
DROP TRIGGER IF EXISTS project_delivery_insert;
DROP TABLE project_delivery_evidence;
