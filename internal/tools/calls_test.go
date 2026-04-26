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

func newCallsServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerCalls(s, mock)
	return s
}

func TestCallLog_OK(t *testing.T) {
	s := newCallsServer(t, mockGetter{
		"/call/log/": []CallEntry{
			{ID: 1, Type: "missed", Number: "0612345678", Name: "Maman", Duration: 0, New: true},
		},
	})
	result := callTool(t, s, "freebox_call_log")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "missed") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestCallLog_APIError(t *testing.T) {
	s := newCallsServer(t, mockGetter{})
	result := callTool(t, s, "freebox_call_log")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
