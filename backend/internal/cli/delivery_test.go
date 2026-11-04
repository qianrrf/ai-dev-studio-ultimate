package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestDeliveryCLIUsesDaemonAndPreservesArgv(t *testing.T) {
	cfg := setConfigEnv(t)
	var received deliveryRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/telemetry/cli-invoked" {
			w.WriteHeader(204)
			return
		}
		if r.URL.Path != "/api/v1/sessions/demo-1/delivery" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method == "POST" {
			if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
				t.Error(err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"evidence":[]}`)
	}))
	t.Cleanup(server.Close)
	writeRunFileFor(t, cfg, server)
	args := []string{"session", "delivery", "demo-1", "--kind", "test", "--request-id", "test-1", "--controller", "desktop", "--control-revision", "4", "--base", strings.Repeat("a", 40), "--head", strings.Repeat("b", 40), "--", "npm.cmd", "test", "path with spaces"}
	if _, stderr, err := executeCLI(t, Deps{ProcessAlive: func(int) bool { return true }}, args...); err != nil {
		t.Fatalf("delivery: %v %s", err, stderr)
	}
	if received.Control.OwnerID != "desktop" || received.Control.Revision != 4 || !reflect.DeepEqual(received.Command, []string{"npm.cmd", "test", "path with spaces"}) {
		t.Fatalf("request: %+v", received)
	}
	if _, stderr, err := executeCLI(t, Deps{ProcessAlive: func(int) bool { return true }}, "session", "delivery", "demo-1"); err != nil {
		t.Fatalf("read evidence: %v %s", err, stderr)
	}
}

func TestDeliveryCLIValidatesBeforeHTTP(t *testing.T) {
	for _, args := range [][]string{
		{"session", "delivery", "demo-1", "--kind", "test"},
		{"session", "delivery", "demo-1", "--request-id", "stray"},
		{"session", "delivery", "demo-1", "--kind", "invalid", "--controller", "desktop", "--control-revision", "1", "--request-id", "x", "--base", "a", "--head", "b"},
	} {
		if _, _, err := executeCLI(t, Deps{}, args...); err == nil {
			t.Fatalf("invalid delivery accepted: %q", args)
		} else if _, ok := err.(usageError); !ok {
			t.Fatalf("want usage error, got %T %v", err, err)
		}
	}
}
