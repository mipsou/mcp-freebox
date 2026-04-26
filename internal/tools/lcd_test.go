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

func newLCDServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerLCD(s, mock)
	return s
}

func TestLCDConfig_OK(t *testing.T) {
	s := newLCDServer(t, mockGetter{
		"/lcd/config/": LCDConfig{Brightness: 50, Orientation: 0},
	})
	result := callTool(t, s, "freebox_lcd_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"brightness"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestLCDConfig_APIError(t *testing.T) {
	s := newLCDServer(t, mockGetter{})
	result := callTool(t, s, "freebox_lcd_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestLCDBrightness_OK(t *testing.T) {
	// mockGetter.Put returns nil — no error expected
	s := newLCDServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_lcd_brightness", map[string]any{"brightness": float64(75)})
	if result.IsError {
		t.Errorf("unexpected error: %v", result.Content)
	}
}
