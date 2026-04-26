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

// ── Sécurité : validation RFC1918 et ports ────────────────────────────────────

func TestValidateRFC1918_Valid(t *testing.T) {
	cases := []string{
		"192.168.100.11", "10.0.0.1", "172.16.0.1",
		"172.31.255.254", "10.255.255.255",
	}
	for _, ip := range cases {
		if err := validateRFC1918(ip); err != nil {
			t.Errorf("validateRFC1918(%q) unexpected error: %v", ip, err)
		}
	}
}

func TestValidateRFC1918_Invalid(t *testing.T) {
	cases := []string{
		"8.8.8.8",        // public IP
		"1.2.3.4",        // public IP
		"172.15.0.1",     // just outside range
		"172.32.0.1",     // just outside range
		"not-an-ip",
		"",
	}
	for _, ip := range cases {
		if err := validateRFC1918(ip); err == nil {
			t.Errorf("validateRFC1918(%q) should have returned error", ip)
		}
	}
}

func TestValidatePort_Valid(t *testing.T) {
	cases := []int{1, 80, 443, 2222, 65535}
	for _, p := range cases {
		if err := validatePort(p, "port"); err != nil {
			t.Errorf("validatePort(%d) unexpected error: %v", p, err)
		}
	}
}

func TestValidatePort_Invalid(t *testing.T) {
	cases := []int{0, -1, 65536, 99999}
	for _, p := range cases {
		if err := validatePort(p, "port"); err == nil {
			t.Errorf("validatePort(%d) should have returned error", p)
		}
	}
}

func TestNATCreate_InvalidIP(t *testing.T) {
	s := newNATServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_nat_create", map[string]any{
		"lan_ip": "8.8.8.8", "lan_port": float64(22),
		"wan_port_start": float64(2222), "ip_proto": "tcp",
	})
	if !result.IsError {
		t.Error("public IP should be rejected for NAT lan_ip")
	}
}

func TestNATCreate_InvalidPort(t *testing.T) {
	s := newNATServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_nat_create", map[string]any{
		"lan_ip": "192.168.100.11", "lan_port": float64(0),
		"wan_port_start": float64(2222), "ip_proto": "tcp",
	})
	if !result.IsError {
		t.Error("port 0 should be rejected")
	}
}

func TestNATCreate_OK(t *testing.T) {
	s := newNATServer(t, mockGetter{
		"/fw/redir/": PortForwarding{ID: 5, Enabled: true, LanIP: "192.168.100.11", LanPort: 22, WanPortStart: 2222, WanPortEnd: 2222, IPProto: "tcp"},
	})
	result := callToolWithArgs(t, s, "freebox_nat_create", map[string]any{
		"lan_ip": "192.168.100.11", "lan_port": float64(22),
		"wan_port_start": float64(2222), "ip_proto": "tcp",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}
