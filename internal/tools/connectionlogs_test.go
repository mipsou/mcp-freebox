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

func newConnectionLogsServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerConnectionLogs(s, mock)
	return s
}

func TestConnectionLogs_OK(t *testing.T) {
	s := newConnectionLogsServer(t, mockGetter{
		"/connection/logs/": []ConnectionLogEntry{
			{ID: 1, Date: 1775928581, Type: "link", State: "up", Link: "ftth", BwDown: 10000000000, BwUp: 900000000},
			{ID: 2, Date: 1775928000, Type: "link", State: "down", Link: "ftth"},
		},
	})
	result := callTool(t, s, "freebox_connection_logs")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"link": "ftth"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestConnectionLogs_APIError(t *testing.T) {
	s := newConnectionLogsServer(t, mockGetter{})
	result := callTool(t, s, "freebox_connection_logs")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
