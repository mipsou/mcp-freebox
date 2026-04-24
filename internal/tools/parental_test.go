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
		"/parental/config/": ParentalConfig{Enabled: true, DefaultPolicy: "allow"},
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
		"/parental/planning/": []ParentalPlanning{
			{ID: 1, Day: 0, Start: 1320, End: 1440, Policy: "deny"}, // lundi 22h-minuit
		},
	})
	result := callTool(t, s, "freebox_parental_planning")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "deny") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestParentalPlanning_APIError(t *testing.T) {
	s := newParentalServer(t, mockGetter{})
	result := callTool(t, s, "freebox_parental_planning")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestParentalFilters_OK(t *testing.T) {
	s := newParentalServer(t, mockGetter{
		"/parental/filter/": []ParentalFilter{
			{ID: 1, MACAddr: "aa:bb:cc:dd:ee:ff", Comment: "Tablette enfant", Enabled: true},
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
