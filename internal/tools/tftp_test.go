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

func newTFTPServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerTFTP(s, mock)
	return s
}

func TestTFTPConfig_OK(t *testing.T) {
	s := newTFTPServer(t, mockGetter{
		"/tftp/config/": TftpConfig{
			Enabled: true,
			Root:    "L0ZyZWVib3gvdGZ0cA==", // /Freebox/tftp
		},
	})
	result := callTool(t, s, "freebox_tftp_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"enabled": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestTFTPConfig_APIError(t *testing.T) {
	s := newTFTPServer(t, mockGetter{})
	result := callTool(t, s, "freebox_tftp_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
