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

func newFTPServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerFTP(s, mock)
	return s
}

func TestFTPConfig_OK(t *testing.T) {
	s := newFTPServer(t, mockGetter{
		"/ftp/config/": FtpConfig{
			Enabled: true, AllowRemoteAccess: true,
			PortCtrl: 21, PortData: 20, RemoteDomain: "ftp.example.com",
		},
	})
	result := callTool(t, s, "freebox_ftp_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"enabled": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestFTPConfig_APIError(t *testing.T) {
	s := newFTPServer(t, mockGetter{})
	result := callTool(t, s, "freebox_ftp_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
