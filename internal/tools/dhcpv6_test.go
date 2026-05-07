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

func newDHCPv6Server(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerDHCPv6(s, mock)
	return s
}

func TestDHCPv6Config_OK(t *testing.T) {
	s := newDHCPv6Server(t, mockGetter{
		"/dhcpv6/config/": DHCPv6Config{
			Enabled:      false,
			UseCustomDNS: true,
			DNS:          []string{"fe80::1", "2a01:e0a:b:ef40::1"},
		},
	})
	result := callTool(t, s, "freebox_dhcpv6_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"use_custom_dns": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDHCPv6Config_APIError(t *testing.T) {
	s := newDHCPv6Server(t, mockGetter{})
	result := callTool(t, s, "freebox_dhcpv6_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
