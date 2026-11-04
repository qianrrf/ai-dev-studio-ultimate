package codexappserver

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"slices"
	"testing"

	"github.com/aoagents/agent-orchestrator/backend/internal/adapters/chatdriver/persistenthost"
	"github.com/aoagents/agent-orchestrator/backend/internal/ports"
)

func TestReadOnlySandboxOverrideUsesNativeProcessArguments(t *testing.T) {
	for _, value := range []string{"", "elevated", "unelevated", "mxc", "invalid"} {
		args, err := appServerArgs([]string{"PATH=ignored", "AO_CODEX_WINDOWS_SANDBOX=" + value})
		if value == "invalid" {
			if err == nil {
				t.Fatal("accepted unknown native sandbox")
			}
			continue
		}
		want := []string{"app-server"}
		if value == "" && runtime.GOOS == "windows" {
			want = []string{"-c", "windows.sandbox=unelevated", "app-server"}
		}
		if value != "" {
			want = []string{"-c", "windows.sandbox=" + value, "app-server"}
		}
		if err != nil || !slices.Equal(args, want) {
			t.Fatalf("%s: args=%v err=%v", value, args, err)
		}
	}
}

func TestReadOnlyReconnectVerifiesSurvivingHostSandbox(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows sandbox policy")
	}
	d, srv := newTestDriver(t)
	srv.replyError("initialize", -32600, "Already initialized")
	srv.reply("config/read", `{"config":{"windows":{"sandbox":"disabled"}}}`)
	proc, err := d.spawn(t.Context(), "codex", t.TempDir(), nil)
	if err != nil {
		t.Fatal(err)
	}
	d.persistent = true
	d.connectHost = func(context.Context, persistenthost.Config) (*persistenthost.Transport, error) {
		return &persistenthost.Transport{Stdin: proc.stdin, Stdout: proc.stdout, Reconnected: true, NextRequestID: 41}, nil
	}
	// A new daemon's preference does not change the surviving native process.
	_, err = d.Reconnect(t.Context(), ports.ChatResumeConfig{
		SessionID: "readonly-reconnect", WorkspacePath: t.TempDir(), DataDir: t.TempDir(),
		ProviderConversationID: "thread-1", ReadOnly: true, DisableExternalTools: true, DisableNetwork: true,
		Env: map[string]string{"AO_CODEX_WINDOWS_SANDBOX": "elevated"},
	})
	if !errors.Is(err, ports.ErrChatUnsupported) || srv.sentMethod("turn/start") || srv.sentMethod("thread/resume") {
		t.Fatalf("adopted unsafe surviving process: %v", err)
	}
}

func TestReadOnlyExcludesExternalToolsOnEveryOpenPath(t *testing.T) {
	for _, method := range []string{"start", "resume", "fork"} {
		t.Run(method, func(t *testing.T) {
			d, srv := newTestDriver(t)
			srv.reply("config/read", `{"config":{"windows":{"sandbox":"elevated"},"mcp_servers":{"unsafe":{"command":"must-not-launch"}},"plugins":{"managed@local":{"enabled":true}},"apps":{"connector":{"enabled":true}}}}`)
			srv.reply("thread/fork", `{"thread":{"id":"child"}}`)
			cfg := ports.ChatStartConfig{WorkspacePath: t.TempDir(), ReadOnly: true, DisableExternalTools: true, DisableNetwork: true, MCPServers: []ports.ChatMCPServerConfig{{Name: "ao", Command: "ao"}}}
			var conv ports.ChatConversation
			var err error
			switch method {
			case "start":
				conv, err = d.Start(t.Context(), cfg)
			case "resume":
				conv, err = d.Resume(t.Context(), ports.ChatResumeConfig{WorkspacePath: cfg.WorkspacePath, ProviderConversationID: "thread-1", ReadOnly: true, DisableExternalTools: true, DisableNetwork: true, MCPServers: cfg.MCPServers})
			case "fork":
				conv, err = d.ForkIntoHost(t.Context(), "parent", "turn", cfg)
			}
			if err != nil {
				t.Fatal(err)
			}
			defer conv.Close()
			if conv.(ports.ChatSandboxNetwork).SandboxAllowsNetwork(ports.PermissionModeBypassPermissions) {
				t.Fatal("read-only public review permits tool network access")
			}
			var params struct {
				Sandbox string                     `json:"sandbox"`
				Config  map[string]json.RawMessage `json:"config"`
			}
			f := srv.awaitFrame(func(f frame) bool { return f.Method == "thread/"+method })
			if err := json.Unmarshal(f.Params, &params); err != nil {
				t.Fatal(err)
			}
			if params.Sandbox != "read-only" {
				t.Fatal("sandbox not read-only")
			}
			var mcp map[string]map[string]any
			if err := json.Unmarshal(params.Config["mcp_servers"], &mcp); err != nil || mcp["ao"]["command"] != "ao" {
				t.Fatal("disabled supplied server lost its required transport")
			}
			for section, names := range map[string][]string{"mcp_servers": {"unsafe", "ao"}, "plugins": {"_default", "managed@local"}, "apps": {"_default", "connector"}, "agents": {"enabled"}} {
				var settings map[string]json.RawMessage
				if err := json.Unmarshal(params.Config[section], &settings); err != nil {
					t.Fatal(err)
				}
				for _, name := range names {
					if section == "agents" {
						if string(settings[name]) != "false" {
							t.Fatal("agents enabled")
						}
						continue
					}
					var entry struct {
						Enabled *bool `json:"enabled"`
					}
					if err := json.Unmarshal(settings[name], &entry); err != nil || entry.Enabled == nil || *entry.Enabled {
						t.Fatalf("%s.%s is not disabled: %s", section, name, settings[name])
					}
				}
			}
			if _, err := conv.(ports.ChatMCPReloader).ReloadMCPServers(t.Context()); !errors.Is(err, ports.ErrChatUnsupported) || srv.sentMethod("config/mcpServer/reload") {
				t.Fatalf("reload escaped read-only configuration: %v", err)
			}
		})
	}
}

func TestReadOnlyToolInventoryFailureRefusesLaunch(t *testing.T) {
	d, srv := newTestDriver(t)
	srv.replyError("config/read", -32603, "unavailable")
	if _, err := d.Start(t.Context(), ports.ChatStartConfig{WorkspacePath: t.TempDir(), ReadOnly: true, DisableExternalTools: true}); err == nil || srv.sentMethod("thread/start") {
		t.Fatal("launched with unknown tool inventory")
	}
}

func TestReadOnlyRequiresExplicitWindowsSandbox(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows sandbox policy")
	}
	for _, sandbox := range []string{"", "disabled", "unelevated", "elevated", "mxc"} {
		d, srv := newTestDriver(t)
		cfg := ports.ChatStartConfig{WorkspacePath: t.TempDir(), ReadOnly: true, DisableExternalTools: true, DisableNetwork: true}
		inventory, _ := json.Marshal(map[string]any{"config": map[string]any{"windows": map[string]any{"sandbox": sandbox}}})
		srv.reply("config/read", string(inventory))
		conv, err := d.Start(t.Context(), cfg)
		if sandbox == "" || sandbox == "disabled" {
			if !errors.Is(err, ports.ErrChatUnsupported) || srv.sentMethod("thread/start") {
				t.Fatalf("unconfigured sandbox accepted: %v", err)
			}
		} else {
			if err != nil {
				t.Fatalf("native sandbox %s: %v", sandbox, err)
			}
			_ = conv.Close()
		}
	}
}

func TestReadOnlyVerifiesManagedToolsBeforeDispatch(t *testing.T) {
	for _, inventory := range []string{`{"data":[{"name":"managed","tools":{"write":{}},"resources":[],"resourceTemplates":[]}]}`, `{"data":[],"nextCursor":"same"}`, `{}`} {
		d, srv := newTestDriver(t)
		srv.reply("config/read", `{"config":{"windows":{"sandbox":"elevated"}}}`)
		srv.reply("mcpServerStatus/list", inventory)
		conv, err := d.Start(t.Context(), ports.ChatStartConfig{WorkspacePath: t.TempDir(), ReadOnly: true, DisableExternalTools: true, DisableNetwork: true})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := conv.SendTurn(t.Context(), ports.ChatUserMessage{Text: "review"}); !errors.Is(err, ports.ErrChatUnsupported) || srv.sentMethod("turn/start") {
			t.Fatalf("unverified tools received a prompt: %v", err)
		}
		_ = conv.Close()
	}
}

func TestReadOnlyAllowsVerifiedNativeTools(t *testing.T) {
	d, srv := newTestDriver(t)
	srv.reply("config/read", `{"config":{"windows":{"sandbox":"elevated"}}}`)
	srv.reply("mcpServerStatus/list", `{"data":[{"name":"disabled","tools":{},"resources":[],"resourceTemplates":[]}]}`)
	conv, err := d.Start(t.Context(), ports.ChatStartConfig{WorkspacePath: t.TempDir(), ReadOnly: true, DisableExternalTools: true, DisableNetwork: true})
	if err != nil {
		t.Fatal(err)
	}
	defer conv.Close()
	if _, err := conv.SendTurn(t.Context(), ports.ChatUserMessage{Text: "review", Settings: ports.ChatTurnSettings{Approval: ports.PermissionModeBypassPermissions}}); err != nil {
		t.Fatal(err)
	}
	var params struct {
		Approval string `json:"approvalPolicy"`
		Sandbox  struct {
			Type    string `json:"type"`
			Network bool   `json:"networkAccess"`
		} `json:"sandboxPolicy"`
	}
	f := srv.awaitFrame(func(f frame) bool { return f.Method == "turn/start" })
	if err := json.Unmarshal(f.Params, &params); err != nil {
		t.Fatal(err)
	}
	if params.Approval != "never" || params.Sandbox.Type != "readOnly" || params.Sandbox.Network {
		t.Fatalf("turn escaped read-only policy: %s", f.Params)
	}
}
