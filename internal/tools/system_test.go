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

func newSystemServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerSystem(s, mock)
	return s
}

func TestSystem_OK(t *testing.T) {
	s := newSystemServer(t, mockGetter{
		"/system/": SystemInfo{
			MAC:              "68:a3:78:00:00:01",
			Serial:           "6312345678",
			Uptime:           "2 days 3 hours",
			UptimeVal:        183600,
			BoardName:        "delta",
			FirmwareVersion:  "4.8.8",
			DiskStatus:       "active",
			BoxAuthenticated: true,
			Sensors: []SystemSensor{
				{ID: "temp_cpu", Name: "Température CPU", Value: 58},
				{ID: "temp_sw", Name: "Température Switch", Value: 47},
			},
			Fans: []SystemFan{
				{ID: "fan0", Name: "Fan 0", Value: 1200},
			},
		},
	})
	result := callTool(t, s, "freebox_system")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"firmware_version": "4.8.8"`) {
		t.Errorf("missing firmware_version in: %s", text)
	}
	if !strings.Contains(text, `"uptime_val": 183600`) {
		t.Errorf("missing uptime_val in: %s", text)
	}
	if !strings.Contains(text, `"temp_cpu"`) {
		t.Errorf("missing sensor temp_cpu in: %s", text)
	}
}

func TestSystem_APIError(t *testing.T) {
	s := newSystemServer(t, mockGetter{})
	result := callTool(t, s, "freebox_system")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
