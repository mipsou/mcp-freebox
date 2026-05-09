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

func newLANServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerLAN(s, mock)
	return s
}

func TestLanHosts_OK(t *testing.T) {
	s := newLANServer(t, mockGetter{
		"/lan/browser/pub/": []LanHost{
			{
				ID:          "ether-aa:bb:cc:dd:ee:ff",
				PrimaryName: "MyPC",
				Reachable:   true,
				L3Connectivities: []L3Connectivity{
					{Addr: "192.168.1.100", AF: "ipv4", Active: true},
				},
			},
		},
	})
	result := callTool(t, s, "freebox_lan_hosts")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"primary_name": "MyPC"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestLanHosts_APIError(t *testing.T) {
	s := newLANServer(t, mockGetter{})
	result := callTool(t, s, "freebox_lan_hosts")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestLanInterfaces_OK(t *testing.T) {
	s := newLANServer(t, mockGetter{
		"/lan/browser/interfaces/": []LanInterface{
			{Name: "pub", HostCount: 12},
			{Name: "guest", HostCount: 3},
		},
	})
	result := callTool(t, s, "freebox_lan_interfaces")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"name": "pub"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"host_count": 12`) {
		t.Errorf("missing host_count: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestLanInterfaces_APIError(t *testing.T) {
	s := newLANServer(t, mockGetter{})
	result := callTool(t, s, "freebox_lan_interfaces")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
