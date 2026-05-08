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
