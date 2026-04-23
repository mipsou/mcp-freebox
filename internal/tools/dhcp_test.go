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

func newDHCPServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerDHCP(s, mock)
	return s
}

func TestDhcpStatic_OK(t *testing.T) {
	s := newDHCPServer(t, mockGetter{
		"/dhcp/static_lease/": []DhcpStaticLease{
			{ID: "aa:bb:cc:dd:ee:ff", Mac: "aa:bb:cc:dd:ee:ff", Hostname: "mypc", IP: "192.168.1.100"},
		},
	})
	result := callTool(t, s, "freebox_dhcp_static")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"ip": "192.168.1.100"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDhcpStatic_APIError(t *testing.T) {
	s := newDHCPServer(t, mockGetter{})
	result := callTool(t, s, "freebox_dhcp_static")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestDhcpLeases_OK(t *testing.T) {
	s := newDHCPServer(t, mockGetter{
		"/dhcp/dynamic_lease/": []DhcpDynamicLease{
			{Mac: "11:22:33:44:55:66", Hostname: "tablet", IP: "192.168.1.42", AssignTime: 1700000000, ExpireTime: 1700086400},
		},
	})
	result := callTool(t, s, "freebox_dhcp_leases")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"hostname": "tablet"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDhcpLeases_APIError(t *testing.T) {
	s := newDHCPServer(t, mockGetter{"/dhcp/static_lease/": []DhcpStaticLease{}})
	result := callTool(t, s, "freebox_dhcp_leases")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
