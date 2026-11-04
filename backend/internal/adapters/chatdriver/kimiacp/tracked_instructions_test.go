package kimiacp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/chatdriver/persistenthost"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "acp" {
		os.Exit(runTrackedInstructionsACP())
	}
	if len(os.Args) >= 7 && os.Args[1] == "chat-host" {
		protocol := persistenthost.ProtocolRaw
		fingerprint := ""
		separator := 5
		if os.Args[5] == string(persistenthost.ProtocolACP) {
			protocol = persistenthost.ProtocolACP
			if len(os.Args) > 6 {
				fingerprint = os.Args[6]
			}
			separator = 7
		}
		if len(os.Args) <= separator || os.Args[separator] != "--" {
			os.Exit(2)
		}
		err := persistenthost.Run(context.Background(), persistenthost.Config{
			SessionID: os.Args[2], DataDir: os.Args[3], Workdir: os.Args[4],
			Env: os.Environ(), Argv: os.Args[separator+1:], Protocol: protocol,
			OwnershipFingerprint: fingerprint,
		})
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestKimiHandoffReachesKimiCodeWhenTheProjectCommitsItsInstructions(t *testing.T) {
	for _, tracked := range []bool{false, true} {
		t.Run(fmt.Sprintf("tracked=%v", tracked), func(t *testing.T) {
			workspace := t.TempDir()
			t.Setenv("KIMI_CODE_HOME", t.TempDir())
			path := filepath.Join(workspace, ".kimi-code", "AGENTS.md")
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			const original = "# Project rules\nUse the project's conventions.\n"
			if err := os.WriteFile(path, []byte(original), 0600); err != nil {
				t.Fatal(err)
			}
			if tracked {
				for _, args := range [][]string{{"init", "-q"}, {"add", ".kimi-code/AGENTS.md"}} {
					cmd := exec.Command("git", args...)
					cmd.Dir = workspace
					if out, err := cmd.CombinedOutput(); err != nil {
						t.Fatalf("git %v: %v: %s", args, err, out)
					}
				}
			}
			fake, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			capture := filepath.Join(t.TempDir(), "requests.jsonl")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			driver := New(fakePlugin{binary: fake, status: ports.AgentAuthStatusAuthorized}, nil)
			const note = "<ao-handoff-request>Earlier user: remember PINEAPPLE-42</ao-handoff-request>"
			conv, err := driver.Start(ctx, ports.ChatStartConfig{SessionID: "kimi-repro", DataDir: t.TempDir(), WorkspacePath: workspace, Permissions: ports.PermissionModeDefault, SystemPrompt: note, Env: map[string]string{"AO_KIMI_REPRO_CAPTURE": capture}})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = conv.(ports.ChatProviderTerminator).Terminate() }()
			ref, err := conv.SendTurn(ctx, ports.ChatUserMessage{Text: "What is the codeword?"})
			if err != nil {
				t.Fatal(err)
			}
			if err := conv.(ports.ChatDeferredTurnStarter).StartDeferredTurn(ref.ProviderTurnID); err != nil {
				t.Fatal(err)
			}
			for {
				select {
				case ev := <-conv.Events():
					if ev.Kind == ports.ChatEventTurnCompleted {
						goto settled
					}
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
			}
		settled:
			loaded, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if tracked && string(loaded) != original {
				t.Fatal("tracked project rules were modified")
			}
			raw, err := os.ReadFile(capture)
			if err != nil {
				t.Fatal(err)
			}
			var channels []string
			for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
				var frame struct {
					Method string          `json:"method"`
					Params json.RawMessage `json:"params"`
				}
				if err := json.Unmarshal([]byte(line), &frame); err != nil {
					t.Fatal(err)
				}
				if frame.Method == "session/new" || frame.Method == "session/prompt" {
					channels = append(channels, string(frame.Params))
				}
			}
			if !strings.Contains(string(loaded), note) && !strings.Contains(strings.Join(channels, "\n"), "PINEAPPLE-42") {
				t.Fatal("handoff missing from every Kimi Code delivery channel: project instructions, session/new metadata, and first prompt")
			}
		})
	}
}

// The test executable is the protocol peer on every OS; no shebang or Python runtime is needed.
func runTrackedInstructionsACP() int {
	file, err := os.OpenFile(os.Getenv("AO_KIMI_REPRO_CAPTURE"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return 1
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	for scanner.Scan() {
		var request struct {
			ID     *json.RawMessage
			Method string
		}
		if json.Unmarshal(scanner.Bytes(), &request) != nil {
			return 1
		}
		if _, err := file.Write(append(append([]byte(nil), scanner.Bytes()...), '\n')); err != nil {
			return 1
		}
		if request.ID == nil {
			continue
		}
		result := map[string]any{}
		switch request.Method {
		case "initialize":
			result = map[string]any{"protocolVersion": 1, "agentCapabilities": map[string]any{}, "authMethods": []any{}}
		case "session/new":
			result["sessionId"] = "kimi-repro-session"
		case "session/prompt":
			result["stopReason"] = "end_turn"
		}
		if err := encoder.Encode(map[string]any{"jsonrpc": "2.0", "id": *request.ID, "result": result}); err != nil {
			return 1
		}
	}
	if scanner.Err() != nil {
		return 1
	}
	return 0
}
