/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newNetshareServer(t *testing.T, mock mockWriter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerNetshare(s, mock)
	return s
}

func TestSambaConfig_OK(t *testing.T) {
	s := newNetshareServer(t, mockWriter{mockGetter: mockGetter{
		"/netshare/samba/": SambaConfig{FileShareEnabled: true, Workgroup: "WORKGROUP"},
	}})
	result := callTool(t, s, "freebox_samba_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"file_share_enabled": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSambaConfig_APIError(t *testing.T) {
	s := newNetshareServer(t, mockWriter{})
	result := callTool(t, s, "freebox_samba_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// ── freebox_samba_config_set ──────────────────────────────────────────────────

func TestSambaConfigSet_OK(t *testing.T) {
	s := newNetshareServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_samba_config_set", map[string]any{
		"file_share_enabled": true,
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestSambaConfigSet_MultipleFields(t *testing.T) {
	s := newNetshareServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_samba_config_set", map[string]any{
		"file_share_enabled":  true,
		"print_share_enabled": false,
		"workgroup":           "HOMELAB",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestSambaConfigSet_NoArgs(t *testing.T) {
	s := newNetshareServer(t, mockWriter{})
	result := callToolWithArgs(t, s, "freebox_samba_config_set", map[string]any{})
	if !result.IsError {
		t.Error("expected error when no fields provided")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "aucun champ") {
		t.Errorf("expected 'aucun champ' error, got: %s", text)
	}
}

func TestSambaConfigSet_PutError(t *testing.T) {
	s := newNetshareServer(t, mockWriter{
		mockGetter: mockGetter{},
		putErrs:    map[string]error{"/netshare/samba/": fmt.Errorf("permission denied")},
	})
	result := callToolWithArgs(t, s, "freebox_samba_config_set", map[string]any{
		"file_share_enabled": false,
	})
	if !result.IsError {
		t.Error("expected tool error result when PUT fails")
	}
}

// ── freebox_samba_shares ──────────────────────────────────────────────────────

func TestSambaShares_OK(t *testing.T) {
	s := newNetshareServer(t, mockWriter{mockGetter: mockGetter{
		"/netshare/samba/share/": []SambaShare{
			{ID: "share1", Name: "Freebox", Path: "/", ReadOnly: false},
		},
	}})
	result := callTool(t, s, "freebox_samba_shares")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "Freebox") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSambaShares_APIError(t *testing.T) {
	s := newNetshareServer(t, mockWriter{})
	result := callTool(t, s, "freebox_samba_shares")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// ── freebox_afp_config ────────────────────────────────────────────────────────

func TestAFPConfig_OK(t *testing.T) {
	s := newNetshareServer(t, mockWriter{mockGetter: mockGetter{
		"/netshare/afp/": AFPConfig{Enabled: false},
	}})
	result := callTool(t, s, "freebox_afp_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"enabled"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestAFPConfig_APIError(t *testing.T) {
	s := newNetshareServer(t, mockWriter{})
	result := callTool(t, s, "freebox_afp_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
