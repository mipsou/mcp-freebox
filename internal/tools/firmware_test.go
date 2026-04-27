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

func newFirmwareServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerFirmware(s, mock)
	return s
}

func TestFirmwareUpdateStatus_OK(t *testing.T) {
	s := newFirmwareServer(t, mockGetter{
		"/system/update/": FirmwareUpdate{UpdateAvailable: true, Version: "4.8.0"},
	})
	result := callTool(t, s, "freebox_firmware_update_status")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "4.8.0") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestFirmwareUpdateStatus_APIError(t *testing.T) {
	s := newFirmwareServer(t, mockGetter{})
	result := callTool(t, s, "freebox_firmware_update_status")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
