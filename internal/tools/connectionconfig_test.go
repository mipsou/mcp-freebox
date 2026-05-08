/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newConnectionConfigServer(t *testing.T, mock mockWriter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerConnectionConfig(s, mock)
	return s
}

func TestConnectionConfig_OK(t *testing.T) {
	s := newConnectionConfigServer(t, mockWriter{mockGetter: mockGetter{
		"/connection/config/": ConnectionConfig{Ping: true, RemoteAccess: false, RemoteAccessPort: 8765},
	}})
	result := callTool(t, s, "freebox_connection_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "8765") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestConnectionConfig_APIError(t *testing.T) {
	s := newConnectionConfigServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callTool(t, s, "freebox_connection_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// ── freebox_connection_config_set ─────────────────────────────────────────────

func TestConnectionConfigSet_OK(t *testing.T) {
	s := newConnectionConfigServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_connection_config_set", map[string]any{
		"ping": true, "remote_access_port": float64(443),
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestConnectionConfigSet_NoArgs(t *testing.T) {
	s := newConnectionConfigServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_connection_config_set", map[string]any{})
	if !result.IsError {
		t.Error("expected error when no fields provided")
	}
}

func TestConnectionConfigSet_APIError(t *testing.T) {
	s := newConnectionConfigServer(t, mockWriter{
		mockGetter: mockGetter{},
		putErrs:    map[string]error{"/connection/config/": fmt.Errorf("permission denied")},
	})
	result := callToolWithArgs(t, s, "freebox_connection_config_set", map[string]any{
		"ping": false,
	})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
