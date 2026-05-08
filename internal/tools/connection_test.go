/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// mockGetter implements getter with a fixed response map.
type mockGetter map[string]any

func (m mockGetter) Get(_ context.Context, path string, dst any) error {
	v, ok := m[path]
	if !ok {
		return &notFoundErr{path}
	}
	b, _ := json.Marshal(v)
	return json.Unmarshal(b, dst)
}

func (m mockGetter) Post(_ context.Context, _ string, _, _ any) error                  { return nil }
func (m mockGetter) PostForm(_ context.Context, _ string, _ url.Values, _ any) error   { return nil }
func (m mockGetter) Put(_ context.Context, _ string, _, _ any) error                   { return nil }
func (m mockGetter) Delete(_ context.Context, _ string) error                          { return nil }

type notFoundErr struct{ path string }

func (e *notFoundErr) Error() string { return "not found: " + e.path }

// callTool invokes a registered tool by name with no arguments.
func callTool(t *testing.T, s *server.MCPServer, name string) *mcp.CallToolResult {
	t.Helper()
	return callToolWithArgs(t, s, name, nil)
}

func newServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerConnection(s, mock)
	return s
}

func TestConnectionStatus_OK(t *testing.T) {
	s := newServer(t, mockGetter{
		"/connection/": ConnectionStatus{State: "up", Type: "ethernet", IPv4: "1.2.3.4"},
	})
	result := callTool(t, s, "freebox_connection_status")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"ipv4": "1.2.3.4"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestConnectionStatus_APIError(t *testing.T) {
	s := newServer(t, mockGetter{})
	result := callTool(t, s, "freebox_connection_status")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestConnectionXdsl_OK(t *testing.T) {
	s := newServer(t, mockGetter{
		"/connection/xdsl/": XdslStatus{
			Status: struct {
				Status     string `json:"status"`
				Modulation string `json:"modulation"`
				Uptime     int    `json:"uptime"`
			}{Status: "showtime", Modulation: "VDSL2", Uptime: 3600},
			Down: XdslLine{SNR: 120, Attn: 100},
		},
	})
	result := callTool(t, s, "freebox_connection_xdsl")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"status": "showtime"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestConnectionFtth_OK(t *testing.T) {
	s := newServer(t, mockGetter{
		"/connection/ftth/": FtthStatus{SfpPresent: true, Link: true, SfpPwrRx: -1924},
	})
	result := callTool(t, s, "freebox_connection_ftth")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"sfp_present": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDynDNSList_OK(t *testing.T) {
	s := newServer(t, mockGetter{
		"/dynDns/": []DynDNSEntry{{Enabled: true, Hostname: "home.example.com", State: "ok"}},
	})
	result := callTool(t, s, "freebox_dyndns_list")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"hostname": "home.example.com"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}
