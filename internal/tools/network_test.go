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

func newNetworkServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerNetwork(s, mock)
	return s
}

func TestRoutesIPv4_OK(t *testing.T) {
	s := newNetworkServer(t, mockGetter{
		"/network/route/ipv4/": []StaticRoute{
			{ID: "r1", IP: "10.0.0.0", Mask: "255.255.255.0", Gateway: "192.168.1.11", Active: true, Comment: "CoreOS VMs"},
		},
	})
	result := callTool(t, s, "freebox_routes_ipv4")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "10.0.0.0") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestRoutesIPv4_APIError(t *testing.T) {
	s := newNetworkServer(t, mockGetter{})
	result := callTool(t, s, "freebox_routes_ipv4")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestRoutesIPv6_OK(t *testing.T) {
	s := newNetworkServer(t, mockGetter{
		"/network/route/ipv6/": []StaticRoute6{
			{ID: "r6-1", Dest: "fd15:b9b9:650e:200::", Prefix: 64, Gateway: "fd15:b9b9:650e:100::11", Active: true},
		},
	})
	result := callTool(t, s, "freebox_routes_ipv6")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "fd15") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestRoutesIPv6_APIError(t *testing.T) {
	s := newNetworkServer(t, mockGetter{})
	result := callTool(t, s, "freebox_routes_ipv6")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestRouteAdd_OK(t *testing.T) {
	s := newNetworkServer(t, mockGetter{
		"/network/route/ipv4/": StaticRoute{ID: "r2", IP: "172.16.0.0", Mask: "255.255.0.0", Gateway: "192.168.1.11", Active: true},
	})
	result := callToolWithArgs(t, s, "freebox_route_add", map[string]any{
		"ip":   "172.16.0.0",
		"mask": "255.255.0.0",
		"gw":   "192.168.1.11",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestRouteDelete_OK(t *testing.T) {
	s := newNetworkServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_route_delete", map[string]any{"id": "r1"})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "supprimée") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}
