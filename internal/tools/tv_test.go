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

func newTVServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerTV(s, mock)
	return s
}

func TestTVChannels_OK(t *testing.T) {
	s := newTVServer(t, mockGetter{
		"/tv/channels/": []TVChannel{
			{UUID: "ch1", Name: "TF1", Number: 1, Quality: "hd"},
		},
	})
	result := callTool(t, s, "freebox_tv_channels")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "TF1") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestTVChannels_APIError(t *testing.T) {
	s := newTVServer(t, mockGetter{})
	result := callTool(t, s, "freebox_tv_channels")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestTVRecords_OK(t *testing.T) {
	s := newTVServer(t, mockGetter{
		"/pvr/programmed/": []TVRecord{
			{ID: 1, Name: "Journal 20h", Status: "scheduled", ChannelID: "ch1"},
		},
	})
	result := callTool(t, s, "freebox_tv_records")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "scheduled") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestTVRecords_APIError(t *testing.T) {
	s := newTVServer(t, mockGetter{})
	result := callTool(t, s, "freebox_tv_records")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
