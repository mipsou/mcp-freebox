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

func newConnectionIPv6Server(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerConnectionIPv6(s, mock)
	return s
}

func TestConnectionIPv6Config_OK(t *testing.T) {
	s := newConnectionIPv6Server(t, mockGetter{
		"/connection/ipv6/config/": ConnectionIPv6Config{
			IPv6Enabled:        true,
			IPv6Firewall:       true,
			IPv6PrefixFirewall: true,
			IPv6LL:             "fe80::2266:cfff:fe75:8b2e",
			Delegations: []IPv6Delegation{
				{Prefix: "2a01:e0a:b:ef40::/64", NextHop: ""},
				{Prefix: "2a01:e0a:b:ef41::/64", NextHop: ""},
			},
		},
	})
	result := callTool(t, s, "freebox_connection_ipv6_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	out := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(out, `"ipv6_enabled": true`) {
		t.Errorf("missing ipv6_enabled: %s", out)
	}
	if !strings.Contains(out, `"prefix": "2a01:e0a:b:ef40::/64"`) {
		t.Errorf("missing delegation prefix: %s", out)
	}
}

func TestConnectionIPv6Config_APIError(t *testing.T) {
	s := newConnectionIPv6Server(t, mockGetter{})
	result := callTool(t, s, "freebox_connection_ipv6_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
