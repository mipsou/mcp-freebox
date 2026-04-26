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

func newSysLogServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerSysLog(s, mock)
	return s
}

func TestSysLog_OK(t *testing.T) {
	s := newSysLogServer(t, mockGetter{
		"/system/log/": []SystemLog{
			{Timestamp: 1776866810, Level: "info", Message: "Freebox started", Category: "system"},
		},
	})
	result := callTool(t, s, "freebox_system_log")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "Freebox started") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSysLog_APIError(t *testing.T) {
	s := newSysLogServer(t, mockGetter{})
	result := callTool(t, s, "freebox_system_log")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
