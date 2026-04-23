/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mockDiscoverer struct{ info any }

func (m mockDiscoverer) DiscoverAPI(_ context.Context, dst any) error {
	b, _ := json.Marshal(m.info)
	return json.Unmarshal(b, dst)
}

func newDiscoveryServer(t *testing.T, d discoverer) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerDiscovery(s, d)
	return s
}

func TestDiscover_OK(t *testing.T) {
	s := newDiscoveryServer(t, mockDiscoverer{info: ApiVersionInfo{
		DeviceName:   "Freebox Server",
		APIVersion:   "4.0",
		BoxModelName: "Freebox Delta",
		HTTPSPort:    443,
	}})
	result := callTool(t, s, "freebox_discover")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"box_model_name": "Freebox Delta"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDiscover_NetworkError(t *testing.T) {
	s := newDiscoveryServer(t, mockDiscoverer{info: nil})
	result := callTool(t, s, "freebox_discover")
	if result.IsError {
		t.Fatalf("nil info should produce empty struct, not error")
	}
}
