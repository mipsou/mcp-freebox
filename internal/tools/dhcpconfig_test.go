/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newDHCPConfigServer(t *testing.T, mock mockWriter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerDHCPConfig(s, mock)
	return s
}

func TestDHCPConfig_OK(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{mockGetter: mockGetter{
		"/dhcp/config/": DHCPConfig{
			Enabled: true, GatewayIP: "192.168.1.254",
			IPRangeStart: "192.168.1.10", IPRangeEnd: "192.168.1.199",
			BootServer: "192.168.1.254", BootFile: "/boot/ipxe",
			Options: []DHCPOption{
				{ID: "tftp_server_name", Val: "192.168.1.254"},
			},
		},
	}})
	result := callTool(t, s, "freebox_dhcp_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "192.168.1.254") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDHCPConfig_APIError(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{})
	result := callTool(t, s, "freebox_dhcp_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// Regression #60 : la Freebox retourne options:{} au lieu de [] — ne doit pas crasher.
func TestDHCPConfig_EmptyOptions(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{mockGetter: mockGetter{
		"/dhcp/config/": json.RawMessage(`{"enabled":true,"options":{}}`),
	}})
	result := callTool(t, s, "freebox_dhcp_config")
	if result.IsError {
		t.Fatalf("options:{} should not cause error: %v", result.Content)
	}
}

func TestDHCPOptions_OK(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{mockGetter: mockGetter{
		"/dhcp/config/": DHCPConfig{
			BootServer: "192.168.1.254",
			BootFile:   "/boot/default.ipxe",
			Options: []DHCPOption{
				{ID: "tftp_server_name", Val: "192.168.1.254"},
				{ID: "bootfile_name", Val: "/boot/default.ipxe"},
			},
		},
	}})
	result := callTool(t, s, "freebox_dhcp_options")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "tftp_server_name") {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestDHCPOptions_APIError(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{})
	result := callTool(t, s, "freebox_dhcp_options")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// ── DHCPOptions.UnmarshalJSON ─────────────────────────────────────────────────

// Regression #60 : décodeur JSON robuste pour le type DHCPOptions.
func TestDHCPOptions_UnmarshalEmptyObject(t *testing.T) {
	// La Freebox retourne {} quand la liste est vide — doit produire un slice vide.
	var opts DHCPOptions
	if err := json.Unmarshal([]byte(`{}`), &opts); err != nil {
		t.Fatalf("UnmarshalJSON({}): %v", err)
	}
	if opts == nil {
		t.Error("opts should be non-nil empty slice, got nil")
	}
	if len(opts) != 0 {
		t.Errorf("opts should be empty, got %v", opts)
	}
}

func TestDHCPOptions_UnmarshalArray(t *testing.T) {
	raw := `[{"id":"tftp_server_name","val":"192.168.1.254"}]`
	var opts DHCPOptions
	if err := json.Unmarshal([]byte(raw), &opts); err != nil {
		t.Fatalf("UnmarshalJSON(array): %v", err)
	}
	if len(opts) != 1 || opts[0].ID != "tftp_server_name" {
		t.Errorf("unexpected opts: %v", opts)
	}
}

func TestDHCPOptions_UnmarshalNull(t *testing.T) {
	var opts DHCPOptions
	if err := json.Unmarshal([]byte(`null`), &opts); err != nil {
		t.Fatalf("UnmarshalJSON(null): %v", err)
	}
	if len(opts) != 0 {
		t.Errorf("opts should be empty, got %v", opts)
	}
}

// ── freebox_dhcp_config_set ───────────────────────────────────────────────────

func TestDHCPConfigSet_OK(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{mockGetter: mockGetter{
		"/dhcp/config/": DHCPConfig{Enabled: true, IPRangeStart: "192.168.1.10", IPRangeEnd: "192.168.1.200"},
	}})
	result := callToolWithArgs(t, s, "freebox_dhcp_config_set", map[string]any{
		"enabled": false,
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestDHCPConfigSet_NoArgs(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{mockGetter: mockGetter{
		"/dhcp/config/": DHCPConfig{Enabled: true},
	}})
	result := callToolWithArgs(t, s, "freebox_dhcp_config_set", map[string]any{})
	if !result.IsError {
		t.Error("expected error when no fields provided")
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "aucun champ") {
		t.Errorf("expected 'aucun champ' error, got: %s", text)
	}
}

func TestDHCPConfigSet_GetError(t *testing.T) {
	// GET échoue (pas de /dhcp/config/ dans le mock) → doit retourner une erreur.
	s := newDHCPConfigServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_dhcp_config_set", map[string]any{
		"enabled": false,
	})
	if !result.IsError {
		t.Error("expected tool error result when GET fails")
	}
}

func TestDHCPConfigSet_PutError(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{
		mockGetter: mockGetter{
			"/dhcp/config/": DHCPConfig{Enabled: true},
		},
		putErrs: map[string]error{"/dhcp/config/": fmt.Errorf("permission denied")},
	})
	result := callToolWithArgs(t, s, "freebox_dhcp_config_set", map[string]any{
		"enabled": false,
	})
	if !result.IsError {
		t.Error("expected tool error result when PUT fails")
	}
}

func TestDHCPConfigSet_DNS(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{mockGetter: mockGetter{
		"/dhcp/config/": DHCPConfig{Enabled: true},
	}})
	result := callToolWithArgs(t, s, "freebox_dhcp_config_set", map[string]any{
		"dns": "192.168.1.1, 192.168.1.2",
	})
	if result.IsError {
		t.Fatalf("dns field should be accepted: %v", result.Content)
	}
}

func TestDHCPConfigSet_MultipleFields(t *testing.T) {
	s := newDHCPConfigServer(t, mockWriter{mockGetter: mockGetter{
		"/dhcp/config/": DHCPConfig{Enabled: true, BootServer: "192.168.1.254"},
	}})
	result := callToolWithArgs(t, s, "freebox_dhcp_config_set", map[string]any{
		"boot_server": "",
		"boot_file":   "",
		"sticky_assign": true,
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}
