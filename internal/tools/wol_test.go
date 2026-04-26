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

func newWOLServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerWOL(s, mock)
	return s
}

func TestWOL_OK(t *testing.T) {
	s := newWOLServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_wol", map[string]any{
		"mac": "aa:bb:cc:dd:ee:ff",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "aa:bb:cc:dd:ee:ff") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestWOL_WithIface(t *testing.T) {
	s := newWOLServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_wol", map[string]any{
		"mac":   "11:22:33:44:55:66",
		"iface": "pub",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "pub") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}
