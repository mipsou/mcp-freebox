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

func (m mockGetter) Post(_ context.Context, _ string, _, _ any) error { return nil }
func (m mockGetter) Put(_ context.Context, _ string, _, _ any) error  { return nil }
func (m mockGetter) Delete(_ context.Context, _ string) error         { return nil }

type notFoundErr struct{ path string }

func (e *notFoundErr) Error() string { return "not found: " + e.path }

// callTool invokes a registered tool by name using its Handler directly.
func callTool(t *testing.T, s *server.MCPServer, name string) *mcp.CallToolResult {
	t.Helper()
	st := s.GetTool(name)
	if st == nil {
		t.Fatalf("tool %q not registered", name)
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	result, err := st.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	return result
}

func TestConnectionStatus_OK(t *testing.T) {
	mock := mockGetter{
		"/connection/": ConnectionStatus{
			State: "up",
			Type:  "ethernet",
			IPv4:  "1.2.3.4",
		},
	}

	s := server.NewMCPServer("test", "0.0.0")
	registerConnection(s, mock)

	result := callTool(t, s, "freebox_connection_status")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"ipv4": "1.2.3.4"`) {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestConnectionStatus_APIError(t *testing.T) {
	mock := mockGetter{} // empty — will return notFoundErr

	s := server.NewMCPServer("test", "0.0.0")
	registerConnection(s, mock)

	result := callTool(t, s, "freebox_connection_status")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
