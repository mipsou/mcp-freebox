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

func newNATServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerNAT(s, mock)
	return s
}

func TestNatRules_OK(t *testing.T) {
	s := newNATServer(t, mockGetter{
		"/fw/redir/": []PortForwarding{
			{ID: 1, Enabled: true, Comment: "SSH", LanPort: 22, WanPortStart: 2222, WanPortEnd: 2222, LanIP: "192.168.1.100", IPProto: "tcp", SrcIP: "0.0.0.0"},
		},
	})
	result := callTool(t, s, "freebox_nat_rules")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"comment": "SSH"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestNatRules_APIError(t *testing.T) {
	s := newNATServer(t, mockGetter{})
	result := callTool(t, s, "freebox_nat_rules")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
