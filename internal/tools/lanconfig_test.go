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

func newLANConfigServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerLANConfig(s, mock)
	return s
}

func TestLanConfig_OK(t *testing.T) {
	s := newLANConfigServer(t, mockGetter{
		"/lan/config/": LanConfig{
			IP: "192.168.1.254", Mask: "255.255.255.0",
			NameDNS: "freebox", NameMDNS: "freebox", Mode: "router",
		},
	})
	result := callTool(t, s, "freebox_lan_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "router") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestLanConfig_APIError(t *testing.T) {
	s := newLANConfigServer(t, mockGetter{})
	result := callTool(t, s, "freebox_lan_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestLanHostRename_OK(t *testing.T) {
	s := newLANConfigServer(t, mockGetter{
		"/lan/browser/pub/ether-aa:bb:cc:dd:ee:ff": LanHostName{
			ID: "ether-aa:bb:cc:dd:ee:ff", Name: "CoreOS-11", MAC: "aa:bb:cc:dd:ee:ff",
		},
	})
	result := callToolWithArgs(t, s, "freebox_lan_host_rename", map[string]any{
		"id":   "ether-aa:bb:cc:dd:ee:ff",
		"name": "CoreOS-11",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestLanHostRename_NoError(t *testing.T) {
	// mockGetter.Put always returns nil — rename succeeds with empty struct result.
	s := newLANConfigServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_lan_host_rename", map[string]any{
		"id":   "ether-xx:xx:xx:xx:xx:xx",
		"name": "Test",
	})
	if result.IsError {
		t.Errorf("unexpected error: %v", result.Content)
	}
}
