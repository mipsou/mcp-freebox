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

func newSwitchServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerSwitch(s, mock)
	return s
}

func TestSwitchPorts_OK(t *testing.T) {
	s := newSwitchServer(t, mockGetter{
		"/switch/status/": []SwitchPortStatus{
			{
				ID:     1,
				Link:   "up",
				Speed:  "1000",
				Duplex: "full",
				MacList: []SwitchMacEntry{
					{Mac: "aa:bb:cc:dd:ee:ff", Hostname: "coreos-11"},
				},
			},
			{
				ID:      2,
				Link:    "down",
				Speed:   "100",
				Duplex:  "full",
				MacList: []SwitchMacEntry{},
			},
		},
	})
	result := callTool(t, s, "freebox_switch_ports")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"speed": "1000"`) {
		t.Errorf("missing speed 1000 in: %s", text)
	}
	if !strings.Contains(text, `"hostname": "coreos-11"`) {
		t.Errorf("missing hostname coreos-11 in: %s", text)
	}
}

func TestSwitchPorts_APIError(t *testing.T) {
	s := newSwitchServer(t, mockGetter{})
	result := callTool(t, s, "freebox_switch_ports")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
