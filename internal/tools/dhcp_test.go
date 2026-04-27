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

// ── Sécurité : validation MAC et IP DHCP ──────────────────────────────────────

func TestValidateDHCPIP_Valid(t *testing.T) {
	cases := []string{"192.168.1.50", "192.168.1.100", "10.0.0.10", "172.16.0.5"}
	for _, ip := range cases {
		if err := validateDHCPIP(ip); err != nil {
			t.Errorf("validateDHCPIP(%q) unexpected error: %v", ip, err)
		}
	}
}

func TestValidateDHCPIP_ReservedRejected(t *testing.T) {
	cases := []string{
		"192.168.1.0",   // réseau
		"192.168.1.1",   // gateway
		"192.168.1.254", // Freebox
		"192.168.1.255", // broadcast
	}
	for _, ip := range cases {
		if err := validateDHCPIP(ip); err == nil {
			t.Errorf("validateDHCPIP(%q) should have returned error", ip)
		}
	}
}

func TestDHCPStaticCreate_InvalidMAC(t *testing.T) {
	s := newDHCPServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_dhcp_static_create", map[string]any{
		"mac": "not-a-mac", "ip": "192.168.1.50",
	})
	if !result.IsError {
		t.Error("invalid MAC should return error")
	}
}

func TestDHCPStaticCreate_ReservedIP(t *testing.T) {
	s := newDHCPServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_dhcp_static_create", map[string]any{
		"mac": "aa:bb:cc:dd:ee:ff", "ip": "192.168.1.254",
	})
	if !result.IsError {
		t.Error("reserved IP (.254) should return error")
	}
}

func TestDHCPStaticCreate_OK(t *testing.T) {
	s := newDHCPServer(t, mockGetter{
		"/dhcp/static_lease/": DhcpStaticLease{Mac: "aa:bb:cc:dd:ee:ff", IP: "192.168.1.50", Hostname: "mypc"},
	})
	result := callToolWithArgs(t, s, "freebox_dhcp_static_create", map[string]any{
		"mac": "aa:bb:cc:dd:ee:ff", "ip": "192.168.1.50", "hostname": "mypc",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}
