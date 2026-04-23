/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// mockWriter extends mockGetter with configurable Post behaviour.
type mockWriter struct {
	mockGetter
	postErrs map[string]error
}

func (m mockWriter) Post(_ context.Context, path string, _, _ any) error {
	if err, ok := m.postErrs[path]; ok {
		return err
	}
	return nil
}

func newVMServer(t *testing.T, mock mockWriter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerVM(s, mock)
	return s
}

func callToolWithArgs(t *testing.T, s *server.MCPServer, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	st := s.GetTool(name)
	if st == nil {
		t.Fatalf("tool %q not registered", name)
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	result, err := st.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	return result
}

func TestVMList_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{
		"/vm/": []VM{
			{ID: 0, Name: "HomeAssistant", Status: "running", Memory: 2048, Vcpus: 2, OS: "debian"},
		},
	}})
	result := callTool(t, s, "freebox_vm_list")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"name": "HomeAssistant"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVMList_APIError(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callTool(t, s, "freebox_vm_list")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestVMStart_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_start", map[string]any{"id": float64(0)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"started": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVMStart_APIError(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{},
		postErrs:   map[string]error{"/vm/0/start": fmt.Errorf("permission denied")},
	})
	result := callToolWithArgs(t, s, "freebox_vm_start", map[string]any{"id": float64(0)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestVMKill_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_kill", map[string]any{"id": float64(0)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"killed": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}
