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

func newDownloadConfigServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerDownloadConfig(s, mock)
	return s
}

func TestDownloadConfig_OK(t *testing.T) {
	s := newDownloadConfigServer(t, mockGetter{
		"/downloads/config/": DownloadConfig{MaxDownloadingTasks: 5, MaxDownloadSpeed: 10485760},
	})
	result := callTool(t, s, "freebox_download_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "10485760") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDownloadConfig_APIError(t *testing.T) {
	s := newDownloadConfigServer(t, mockGetter{})
	result := callTool(t, s, "freebox_download_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
