/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newWifiMacFilterServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerWifiMacFilter(s, mock)
	return s
}

func TestWifiMacFilter_OK(t *testing.T) {
	hostBlob := json.RawMessage(`{"vendor_name":"Withings","interface":"pub","primary_name":"Balance"}`)
	s := newWifiMacFilterServer(t, mockGetter{
		"/wifi/mac_filter/": []WifiMacFilterEntry{
			{
				ID:       "00:24:E4:1C:23:C8-whitelist",
				Mac:      "00:24:E4:1C:23:C8",
				Type:     "whitelist",
				Comment:  "Balance Withings",
				Hostname: "Balance",
				Host:     hostBlob,
			},
		},
	})
	result := callTool(t, s, "freebox_wifi_mac_filter")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	out := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(out, `"mac": "00:24:E4:1C:23:C8"`) {
		t.Errorf("missing mac: %s", out)
	}
	if !strings.Contains(out, `"type": "whitelist"`) {
		t.Errorf("missing type: %s", out)
	}
	if !strings.Contains(out, `"vendor_name": "Withings"`) {
		t.Errorf("host blob not preserved: %s", out)
	}
}

func TestWifiMacFilter_APIError(t *testing.T) {
	s := newWifiMacFilterServer(t, mockGetter{})
	result := callTool(t, s, "freebox_wifi_mac_filter")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
