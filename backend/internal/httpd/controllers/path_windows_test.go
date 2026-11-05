//go:build windows

package controllers_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestProjectsAPIWindowsRootRefusal(t *testing.T) {
	srv := newTestServer(t)
	parent := t.TempDir()
	long := filepath.Join(parent, strings.Repeat("x", 201-len(utf16.Encode([]rune(parent)))-1))
	for _, route := range []string{"/projects", "/projects/initialize", "/projects/clone/check", "/projects/clone/prepare", "/projects/clone"} {
		in := map[string]string{"path": long}
		if strings.Contains(route, "/clone") {
			in = map[string]string{"remoteUrl": "https://github.com/example/repo.git", "destinationParent": long}
		}
		payload, err := json.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		body, status, headers := doRequest(t, srv, "POST", "/api/v1"+route, string(payload))
		assertJSON(t, headers)
		assertErrorCode(t, body, status, http.StatusBadRequest, "WINDOWS_GIT_PATH_TOO_LONG")
		if !strings.Contains(string(body), "maxUtf16Units") || !strings.Contains(string(body), "Projects") {
			t.Fatalf("missing path details or shorter-location guidance: %s", body)
		}
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 0 {
		t.Fatalf("validation changed destination parent: %v, %v", entries, err)
	}
}
