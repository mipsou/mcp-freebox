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

func newDHCPConfigServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerDHCPConfig(s, mock)
	return s
}

func TestDHCPConfig_OK(t *testing.T) {
	s := newDHCPConfigServer(t, mockGetter{
		"/dhcp/config/": DHCPConfig{
			Enabled: true, GatewayIP: "192.168.1.254",
			IPRangeStart: "192.168.1.10", IPRangeEnd: "192.168.1.199",
		},
	})
	result := callTool(t, s, "freebox_dhcp_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "192.168.1.254") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDHCPConfig_APIError(t *testing.T) {
	s := newDHCPConfigServer(t, mockGetter{})
	result := callTool(t, s, "freebox_dhcp_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
