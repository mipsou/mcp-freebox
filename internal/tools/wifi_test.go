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

func newWifiServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerWifi(s, mock)
	return s
}

func TestWifiAps_OK(t *testing.T) {
	s := newWifiServer(t, mockGetter{
		"/wifi/ap/": []WifiAp{
			{
				ID:   0,
				Name: "Freebox WiFi",
				Status: WifiApStatus{State: "active", PrimaryChannel: 36, ChannelWidth: "80"},
				Config: WifiApConfig{Band: "5g", Enabled: true, Dfs: true},
			},
		},
	})
	result := callTool(t, s, "freebox_wifi_aps")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"band": "5g"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestWifiAps_APIError(t *testing.T) {
	s := newWifiServer(t, mockGetter{})
	result := callTool(t, s, "freebox_wifi_aps")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestWifiConfig_OK(t *testing.T) {
	s := newWifiServer(t, mockGetter{
		"/wifi/config/": WifiGlobalConfig{Enabled: true, MacFilterState: "disabled"},
	})
	result := callTool(t, s, "freebox_wifi_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"enabled": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestWifiConfig_APIError(t *testing.T) {
	s := newWifiServer(t, mockGetter{"/wifi/ap/": []WifiAp{}})
	result := callTool(t, s, "freebox_wifi_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
