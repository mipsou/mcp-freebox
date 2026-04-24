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

func newFirewallServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerFirewall(s, mock)
	return s
}

func TestFirewallIncoming_OK(t *testing.T) {
	s := newFirewallServer(t, mockGetter{
		"/fw/incoming/": []FirewallIncomingRule{
			{ID: 1, Enabled: true, Comment: "Block scan", Action: "drop", IPProto: "tcp", SrcIP: "0.0.0.0/0"},
		},
	})
	result := callTool(t, s, "freebox_firewall_incoming")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"comment": "Block scan"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestFirewallIncoming_APIError(t *testing.T) {
	s := newFirewallServer(t, mockGetter{})
	result := callTool(t, s, "freebox_firewall_incoming")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestFirewallDMZ_OK(t *testing.T) {
	s := newFirewallServer(t, mockGetter{
		"/fw/dmz/": DMZConfig{Enabled: true, IP: "192.168.100.50"},
	})
	result := callTool(t, s, "freebox_firewall_dmz")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"ip": "192.168.100.50"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestFirewallDMZ_APIError(t *testing.T) {
	s := newFirewallServer(t, mockGetter{"/fw/incoming/": []FirewallIncomingRule{}})
	result := callTool(t, s, "freebox_firewall_dmz")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
