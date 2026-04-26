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

func newSwitchConfigServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerSwitchConfig(s, mock)
	return s
}

func TestSwitchPortConfig_OK(t *testing.T) {
	s := newSwitchConfigServer(t, mockGetter{
		"/switch/port/1/config/": SwitchPortConfig{ID: 1, Enabled: true, DuplexMode: "auto", SpeedMode: "auto"},
	})
	result := callToolWithArgs(t, s, "freebox_switch_port_config", map[string]any{"id": float64(1)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "auto") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSwitchPortConfig_APIError(t *testing.T) {
	s := newSwitchConfigServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_switch_port_config", map[string]any{"id": float64(1)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestSwitchPortStats_OK(t *testing.T) {
	s := newSwitchConfigServer(t, mockGetter{
		"/switch/port/2/stats/": SwitchStats{PortID: 2, RxBytesRate: 1000000, TxBytesRate: 500000},
	})
	result := callToolWithArgs(t, s, "freebox_switch_port_stats", map[string]any{"id": float64(2)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "1000000") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSwitchPortStats_APIError(t *testing.T) {
	s := newSwitchConfigServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_switch_port_stats", map[string]any{"id": float64(1)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
