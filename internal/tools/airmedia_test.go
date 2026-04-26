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

func newAirMediaServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerAirMedia(s, mock)
	return s
}

func TestAirMediaConfig_OK(t *testing.T) {
	s := newAirMediaServer(t, mockGetter{
		"/airmedia/config/": AirMediaConfig{Enabled: true, Password: "secret"},
	})
	result := callTool(t, s, "freebox_airmedia_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "true") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestAirMediaConfig_APIError(t *testing.T) {
	s := newAirMediaServer(t, mockGetter{})
	result := callTool(t, s, "freebox_airmedia_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestAirMediaReceivers_OK(t *testing.T) {
	s := newAirMediaServer(t, mockGetter{
		"/airmedia/receivers/": []AirMediaReceiver{
			{Name: "Freebox Player", PasswordProtected: false, Capabilities: []string{"video", "audio", "photo"}},
		},
	})
	result := callTool(t, s, "freebox_airmedia_receivers")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "Freebox Player") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestAirMediaReceivers_APIError(t *testing.T) {
	s := newAirMediaServer(t, mockGetter{})
	result := callTool(t, s, "freebox_airmedia_receivers")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
