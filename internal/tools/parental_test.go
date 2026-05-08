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

func newParentalServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerParental(s, mock)
	return s
}

func TestParentalConfig_OK(t *testing.T) {
	s := newParentalServer(t, mockGetter{
		"/parental/config/": ParentalConfig{DefaultFilterMode: "allow"},
	})
	result := callTool(t, s, "freebox_parental_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "allow") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestParentalConfig_APIError(t *testing.T) {
	s := newParentalServer(t, mockGetter{})
	result := callTool(t, s, "freebox_parental_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestParentalPlanning_OK(t *testing.T) {
	s := newParentalServer(t, mockGetter{
		"/parental/filter/1/planning": ParentalFilterPlanning{
			Resolution: 48,
			Mapping:    []string{"allowed", "denied"},
		},
	})
	result := callToolWithArgs(t, s, "freebox_parental_planning", map[string]any{"filter_id": float64(1)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "denied") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestParentalPlanning_APIError(t *testing.T) {
	s := newParentalServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_parental_planning", map[string]any{"filter_id": float64(1)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestParentalFilters_OK(t *testing.T) {
	s := newParentalServer(t, mockGetter{
		"/parental/filter/": []ParentalFilter{
			{ID: 1, Macs: []string{"aa:bb:cc:dd:ee:ff"}, Desc: "Tablette enfant"},
		},
	})
	result := callTool(t, s, "freebox_parental_filters")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "Tablette") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestParentalFilters_APIError(t *testing.T) {
	s := newParentalServer(t, mockGetter{})
	result := callTool(t, s, "freebox_parental_filters")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
