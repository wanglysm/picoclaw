// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sipeed/picoclaw/pkg/config"
)

func boolPtr(b bool) *bool { return &b }

func TestServerIsDeferred(t *testing.T) {
	tests := []struct {
		name             string
		discoveryEnabled bool
		serverDeferred   *bool
		want             bool
	}{
		// --- global false always wins: per-server deferred is ignored ---
		{
			name:             "global false: per-server deferred=true is ignored",
			discoveryEnabled: false,
			serverDeferred:   boolPtr(true),
			want:             false,
		},
		{
			name:             "global false: per-server deferred=false stays false",
			discoveryEnabled: false,
			serverDeferred:   boolPtr(false),
			want:             false,
		},
		// --- global true: per-server override applies ---
		{
			name:             "global true: per-server deferred=false opts out",
			discoveryEnabled: true,
			serverDeferred:   boolPtr(false),
			want:             false,
		},
		{
			name:             "global true: per-server deferred=true stays true",
			discoveryEnabled: true,
			serverDeferred:   boolPtr(true),
			want:             true,
		},
		// --- no per-server override: fall back to global ---
		{
			name:             "no per-server field, global discovery enabled",
			discoveryEnabled: true,
			serverDeferred:   nil,
			want:             true,
		},
		{
			name:             "no per-server field, global discovery disabled",
			discoveryEnabled: false,
			serverDeferred:   nil,
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverCfg := config.MCPServerConfig{Deferred: tt.serverDeferred}
			got := serverIsDeferred(tt.discoveryEnabled, serverCfg)
			if got != tt.want {
				t.Errorf("serverIsDeferred(discoveryEnabled=%v, deferred=%v) = %v, want %v",
					tt.discoveryEnabled, tt.serverDeferred, got, tt.want)
			}
		})
	}
}

func TestResolveAgentMCPServerAllowlist(t *testing.T) {
	workspace := t.TempDir()
	agentPath := filepath.Join(workspace, "AGENT.md")
	content := `---
mcpServers: [GitHub, filesystem, github]
---
# Agent
`
	if err := os.WriteFile(agentPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(AGENT.md) error = %v", err)
	}

	allowlist := resolveAgentMCPServerAllowlist(loadAgentDefinition(workspace))
	if len(allowlist) != 2 {
		t.Fatalf("len(allowlist) = %d, want 2", len(allowlist))
	}
	if _, ok := allowlist["github"]; !ok {
		t.Fatal("expected github to be present in MCP allowlist")
	}
	if _, ok := allowlist["filesystem"]; !ok {
		t.Fatal("expected filesystem to be present in MCP allowlist")
	}
}

func TestAgentInstance_AllowsMCPServer(t *testing.T) {
	t.Run("nil allowlist allows all", func(t *testing.T) {
		agent := &AgentInstance{}
		if !agent.AllowsMCPServer("github") {
			t.Fatal("expected nil MCP allowlist to allow all servers")
		}
	})

	t.Run("explicit allowlist filters servers", func(t *testing.T) {
		agent := &AgentInstance{
			MCPServerAllowlist: map[string]struct{}{
				"github": {},
			},
		}
		if !agent.AllowsMCPServer("GitHub") {
			t.Fatal("expected MCP server matching to be case-insensitive")
		}
		if agent.AllowsMCPServer("filesystem") {
			t.Fatal("expected filesystem to be blocked by MCP allowlist")
		}
	})
}
