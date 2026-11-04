package controllers_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/config"
	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd"
	sessionsvc "github.com/aoagents/agent-orchestrator/backend/internal/service/session"
	"github.com/aoagents/agent-orchestrator/backend/internal/storage/sqlite/sqlitetest"
)

func TestProjectControlAPIRevisionFencingAndDurableIntake(t *testing.T) {
	store := sqlitetest.MustOpen(t)
	if err := store.UpsertProject(context.Background(), domain.ProjectRecord{ID: "demo", Path: t.TempDir(), RegisteredAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	svc := sessionsvc.NewWithDeps(sessionsvc.Deps{Store: store})
	server := httptest.NewServer(httpd.NewRouterWithControl(config.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)), nil, httpd.APIDeps{Sessions: svc}, httpd.ControlDeps{}))
	t.Cleanup(server.Close)
	for _, tc := range []struct {
		method, path, body string
		status             int
		contains           string
	}{
		{"GET", "/projects/demo/control", "", 200, `"revision":0`},
		{"PUT", "/projects/demo/control", `{"ownerKind":"external","ownerId":"desktop","expectedRevision":0}`, 200, `"revision":1`},
		{"PUT", "/projects/demo/control", `{"ownerKind":"external","ownerId":"stale","expectedRevision":0}`, 409, "PROJECT_CONTROL_CONFLICT"},
		{"PUT", "/projects/demo/control", `{"ownerKind":"internal","ownerId":"missing","expectedRevision":1}`, 409, "PROJECT_CONTROL_CONFLICT"},
		{"PUT", "/projects/demo/control", `{"ownerKind":"unowned","ownerId":"","expectedRevision":1}`, 200, `"revision":2`},
		{"PUT", "/projects/demo/control", `{"ownerKind":"external","ownerId":"next","expectedRevision":2}`, 200, `"revision":3`},
		{"POST", "/sessions", `{"projectId":"demo","kind":"worker","harness":"deepseek-harness","control":{"ownerId":"desktop","revision":1},"clientRequestId":"old"}`, 409, "PROJECT_CONTROL_CONFLICT"},
		{"POST", "/orchestrators/delegate", `{"projectId":"demo","brief":"bypass"}`, 400, "REQUEST_ID_REQUIRED"},
		{"POST", "/orchestrators/delegate", `{"projectId":"demo","brief":"user intake","clientRequestId":"ui-1"}`, 202, `"requestId":"ui-1"`},
		{"POST", "/projects/demo/requests", `{"id":"user-1","brief":"fix calculator","attachments":[{"mimeType":"text/plain","data":"aGVsbG8="}]}`, 202, `"ext":".txt"`},
		{"POST", "/projects/demo/requests", `{"id":"user-1","brief":"changed"}`, 409, "PROJECT_CONTROL_CONFLICT"},
		{"GET", "/projects/demo/requests", "", 200, `"data":"aGVsbG8="`},
		{"PUT", "/projects/demo/control", `{"ownerKind":"external","ownerId":"bad\nowner","expectedRevision":3}`, 400, "INVALID_CONTROLLER"},
		{"POST", "/projects/demo/requests", `{"id":"","brief":"empty key"}`, 400, "INVALID_REQUIREMENT"},
		{"GET", "/projects/missing/control", "", 404, "PROJECT_NOT_FOUND"},
	} {
		t.Run(tc.method+tc.path+tc.contains, func(t *testing.T) {
			body, status, _ := doRequest(t, server, tc.method, "/api/v1"+tc.path, tc.body)
			if status != tc.status || !strings.Contains(string(body), tc.contains) {
				t.Fatalf("status=%d body=%s; want %d containing %s", status, body, tc.status, tc.contains)
			}
			if !json.Valid(body) {
				t.Fatalf("invalid JSON: %s", body)
			}
		})
	}
}
