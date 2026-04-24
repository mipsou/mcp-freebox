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

func newVPNServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerVPN(s, mock)
	return s
}

func TestVPNServerStatus_OK(t *testing.T) {
	s := newVPNServer(t, mockGetter{
		"/vpn/": []VPNServer{
			{Type: "pptp", Name: "pptp", State: "stopped"},
			{Type: "wireguard", Name: "wireguard", State: "started", ConnectionCount: 0},
		},
	})
	result := callTool(t, s, "freebox_vpn_server_status")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"name": "wireguard"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVPNServerStatus_APIError(t *testing.T) {
	s := newVPNServer(t, mockGetter{})
	result := callTool(t, s, "freebox_vpn_server_status")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestVPNConnections_OK(t *testing.T) {
	s := newVPNServer(t, mockGetter{
		"/vpn/connection/": []VPNConnection{
			{
				ID:           "wg-001",
				VPN:          "wireguard",
				Login:        "mipsou",
				SrcIP:        "1.2.3.4",
				LocalIP:      "10.8.0.2",
				PushedRoutes: []string{"192.168.100.0/24"},
				ConnectedAt:  1714000000,
			},
		},
	})
	result := callTool(t, s, "freebox_vpn_connections")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"login": "mipsou"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVPNConnections_APIError(t *testing.T) {
	s := newVPNServer(t, mockGetter{})
	result := callTool(t, s, "freebox_vpn_connections")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestVPNClientConfigs_OK(t *testing.T) {
	s := newVPNServer(t, mockGetter{
		"/vpn_client/config/": []VPNClientConfig{
			{ID: "vpnc-1", Description: "Mullvad WG", Active: true, Type: "wireguard"},
		},
	})
	result := callTool(t, s, "freebox_vpn_client_configs")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"description": "Mullvad WG"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVPNClientConfigs_APIError(t *testing.T) {
	s := newVPNServer(t, mockGetter{})
	result := callTool(t, s, "freebox_vpn_client_configs")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
