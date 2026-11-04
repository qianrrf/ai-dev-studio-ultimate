-- +goose Up
CREATE TABLE project_control (
    project_id TEXT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    owner_kind TEXT NOT NULL CHECK (owner_kind IN ('internal', 'external', 'unowned')),
    owner_id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK (revision > 0),
    updated_at TIMESTAMP NOT NULL,
    CHECK ((owner_kind = 'unowned' AND owner_id = '') OR (owner_kind != 'unowned' AND length(owner_id) > 0))
);
CREATE TABLE project_requests (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    brief TEXT NOT NULL CHECK (length(trim(brief)) > 0 AND length(brief) <= 16384),
    attachments TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(attachments)),
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_project_requests_project ON project_requests(project_id, created_at, id);
CREATE TABLE project_dispatches (
    request_id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    owner_id TEXT NOT NULL,
    revision INTEGER NOT NULL,
    session_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_project_dispatches_project ON project_dispatches(project_id, session_id);

-- +goose StatementBegin
CREATE TRIGGER project_control_insert AFTER INSERT ON project_control BEGIN
    INSERT INTO change_log (project_id, session_id, event_type, payload, created_at) VALUES (NEW.project_id, NULL, 'session_updated', json_object('projectControl', 1), NEW.updated_at);
END;
CREATE TRIGGER project_control_update AFTER UPDATE ON project_control BEGIN
    INSERT INTO change_log (project_id, session_id, event_type, payload, created_at) VALUES (NEW.project_id, NULL, 'session_updated', json_object('projectControl', 1), NEW.updated_at);
END;
CREATE TRIGGER project_request_insert AFTER INSERT ON project_requests BEGIN
    INSERT INTO change_log (project_id, session_id, event_type, payload, created_at) VALUES (NEW.project_id, NULL, 'session_updated', json_object('projectRequestId', NEW.id), NEW.created_at);
END;
-- +goose StatementEnd
-- +goose Down
DROP TRIGGER IF EXISTS project_request_insert;
DROP TRIGGER IF EXISTS project_control_update;
DROP TRIGGER IF EXISTS project_control_insert;
DROP TABLE project_dispatches;
DROP TABLE project_requests;
DROP TABLE project_control;
