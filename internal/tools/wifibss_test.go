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

func newWifiBSSServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerWifiBSS(s, mock)
	return s
}

func TestWifiSSIDs_OK(t *testing.T) {
	s := newWifiBSSServer(t, mockGetter{
		"/wifi/bss/": []WifiBss{
			{ID: "bss0", BSSID: "aa:bb:cc:dd:ee:ff", SSID: "Freebox-5G", Band: "5g", Enabled: true, Encryption: "wpa2_psk"},
		},
	})
	result := callTool(t, s, "freebox_wifi_ssids")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "Freebox-5G") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestWifiSSIDs_APIError(t *testing.T) {
	s := newWifiBSSServer(t, mockGetter{})
	result := callTool(t, s, "freebox_wifi_ssids")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestWifiStations_OK(t *testing.T) {
	s := newWifiBSSServer(t, mockGetter{
		"/wifi/stations/": []WifiStation{
			{MAC: "11:22:33:44:55:66", Band: "5g", Signal: -55, RxRate: 300000, TxRate: 300000, Active: true},
		},
	})
	result := callTool(t, s, "freebox_wifi_stations")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "11:22:33:44:55:66") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestWifiStations_APIError(t *testing.T) {
	s := newWifiBSSServer(t, mockGetter{})
	result := callTool(t, s, "freebox_wifi_stations")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestWifiSSIDToggle_OK(t *testing.T) {
	// mockGetter.Put returns nil — toggle succeeds
	s := newWifiBSSServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_wifi_ssid_toggle", map[string]any{
		"id":      "bss0",
		"enabled": false,
	})
	if result.IsError {
		t.Errorf("unexpected error: %v", result.Content)
	}
}
