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

func newWifiPlanningServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerWifiPlanning(s, mock)
	return s
}

func TestWifiPlanning_OK(t *testing.T) {
	mapping := make([]string, 168)
	for i := range mapping {
		mapping[i] = "on"
	}
	mapping[0] = "off"
	s := newWifiPlanningServer(t, mockGetter{
		"/wifi/planning/": WifiPlanning{
			UsePlanning: true,
			Resolution:  48,
			Mapping:     mapping,
		},
	})
	result := callTool(t, s, "freebox_wifi_planning")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	out := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(out, `"use_planning": true`) {
		t.Errorf("missing use_planning: %s", out)
	}
	if !strings.Contains(out, `"resolution": 48`) {
		t.Errorf("missing resolution: %s", out)
	}
}

func TestWifiPlanning_APIError(t *testing.T) {
	s := newWifiPlanningServer(t, mockGetter{})
	result := callTool(t, s, "freebox_wifi_planning")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
