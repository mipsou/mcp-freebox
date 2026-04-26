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

func newUPnPServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerUPnP(s, mock)
	return s
}

func TestUPnPConfig_OK(t *testing.T) {
	s := newUPnPServer(t, mockGetter{
		"/upnp/config/": UPnPConfig{Enabled: true},
	})
	result := callTool(t, s, "freebox_upnp_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"enabled"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestUPnPConfig_APIError(t *testing.T) {
	s := newUPnPServer(t, mockGetter{})
	result := callTool(t, s, "freebox_upnp_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestUPnPRules_OK(t *testing.T) {
	s := newUPnPServer(t, mockGetter{
		"/upnp/igd/rules/": []UPnPIGDMapping{
			{ID: 1, InternalIP: "192.168.1.20", ExternalPort: 32400, InternalPort: 32400, Protocol: "tcp", Description: "Plex"},
		},
	})
	result := callTool(t, s, "freebox_upnp_rules")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "Plex") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestUPnPRules_APIError(t *testing.T) {
	s := newUPnPServer(t, mockGetter{})
	result := callTool(t, s, "freebox_upnp_rules")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
