package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProjectControlCLIUsesDaemonAndPreservesConflicts(t *testing.T) {
	cfg := setConfigEnv(t)
	var requests []projectControlRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/telemetry/cli-invoked" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.URL.Path != "/api/v1/projects/demo/control" {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "PUT" {
			var in projectControlRequest
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				t.Error(err)
			}
			requests = append(requests, in)
			if in.ExpectedRevision == 9 {
				w.WriteHeader(409)
				_, _ = io.WriteString(w, `{"error":"conflict","code":"PROJECT_CONTROL_CONFLICT","message":"refresh ownership","requestId":"control-test"}`)
				return
			}
		}
		_, _ = io.WriteString(w, `{"projectId":"demo","ownerKind":"external","ownerId":"desktop","revision":1}`)
	}))
	t.Cleanup(srv.Close)
	writeRunFileFor(t, cfg, srv)
	for _, args := range [][]string{{"project", "control", "demo"}, {"project", "control", "demo", "--owner", "desktop", "--revision", "0"}, {"project", "control", "demo", "--release", "--revision", "1"}} {
		if _, stderr, err := executeCLI(t, Deps{ProcessAlive: func(int) bool { return true }}, args...); err != nil {
			t.Fatalf("%v: %v %s", args, err, stderr)
		}
	}
	if len(requests) != 2 || requests[0].OwnerID != "desktop" || requests[0].ExpectedRevision != 0 || requests[1].OwnerKind != "unowned" || requests[1].OwnerID != "" || requests[1].ExpectedRevision != 1 {
		t.Fatalf("requests: %+v", requests)
	}
	_, stderr, err := executeCLI(t, Deps{ProcessAlive: func(int) bool { return true }}, "project", "control", "demo", "--owner", "stale", "--revision", "9")
	if err == nil || !strings.Contains(err.Error(), "PROJECT_CONTROL_CONFLICT") || !strings.Contains(err.Error(), "control-test") {
		t.Fatalf("conflict envelope: %v %s", err, stderr)
	}
}

func TestControlledCLIRejectsIncompleteClaimsBeforeHTTP(t *testing.T) {
	for _, args := range [][]string{
		{"project", "control", "demo", "--owner", "desktop"},
		{"project", "control", "demo", "--release", "--owner", "desktop", "--revision", "1"},
		{"project", "control", "demo", "--revision", "0"},
		{"spawn", "--name", "task", "--controller", "desktop", "--control-revision", "1"},
		{"project", "request", "demo", "brief"},
	} {
		_, _, err := executeCLI(t, Deps{}, args...)
		if _, ok := err.(usageError); !ok {
			t.Fatalf("%v: want usage error, got %T %v", args, err, err)
		}
	}
}
