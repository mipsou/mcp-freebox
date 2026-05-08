/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newNetshareServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerNetshare(s, mock)
	return s
}

func TestSambaConfig_OK(t *testing.T) {
	s := newNetshareServer(t, mockGetter{
		"/netshare/samba/": SambaConfig{FileShareEnabled: true, Workgroup: "WORKGROUP"},
	})
	result := callTool(t, s, "freebox_samba_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"file_share_enabled": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSambaConfig_APIError(t *testing.T) {
	s := newNetshareServer(t, mockGetter{})
	result := callTool(t, s, "freebox_samba_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestSambaShares_OK(t *testing.T) {
	s := newNetshareServer(t, mockGetter{
		"/netshare/samba/share/": []SambaShare{
			{ID: "share1", Name: "Freebox", Path: "/", ReadOnly: false},
		},
	})
	result := callTool(t, s, "freebox_samba_shares")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "Freebox") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSambaShares_APIError(t *testing.T) {
	s := newNetshareServer(t, mockGetter{})
	result := callTool(t, s, "freebox_samba_shares")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestAFPConfig_OK(t *testing.T) {
	s := newNetshareServer(t, mockGetter{
		"/netshare/afp/": AFPConfig{Enabled: false},
	})
	result := callTool(t, s, "freebox_afp_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"enabled"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestAFPConfig_APIError(t *testing.T) {
	s := newNetshareServer(t, mockGetter{})
	result := callTool(t, s, "freebox_afp_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
