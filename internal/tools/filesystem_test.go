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

func newFSServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerFilesystem(s, mock)
	return s
}

func TestEncodeFSPath(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"/Freebox/VMs", encodeFSPath("/Freebox/VMs")},
		{"Freebox/VMs", encodeFSPath("/Freebox/VMs")}, // leading slash added
		{"/Freebox/VMs/", encodeFSPath("/Freebox/VMs")}, // trailing slash trimmed
	}
	for _, tc := range cases {
		got := encodeFSPath(tc.input)
		if got != tc.want {
			t.Errorf("encodeFSPath(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFSList_OK(t *testing.T) {
	path := "/Freebox/VMs"
	encoded := encodeFSPath(path)
	s := newFSServer(t, mockGetter{
		"/fs/ls/" + encoded: []FSEntry{
			{Name: "Fedora.qcow2", Type: "file", Size: 5368709120, Path: encodeFSPath("/Freebox/VMs/Fedora.qcow2")},
			{Name: "Fedora-Server-KVM-40-1.14.aarch64.qcow2", Type: "file", Size: 1073741824, Path: encodeFSPath("/Freebox/VMs/Fedora-Server-KVM-40-1.14.aarch64.qcow2")},
		},
	})
	req := callToolWithArgs(t, s, "freebox_fs_list", map[string]any{"path": path})
	if req.IsError {
		t.Fatalf("tool returned error: %v", req.Content)
	}
	if !strings.Contains(req.Content[0].(mcp.TextContent).Text, `"name": "Fedora.qcow2"`) {
		t.Errorf("unexpected result: %s", req.Content[0].(mcp.TextContent).Text)
	}
}

func TestFSList_APIError(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_list", map[string]any{"path": "/Freebox/VMs"})
	if !req.IsError {
		t.Error("expected tool error result")
	}
}

// mockGetter.Post always returns nil → mkdir/delete never error
func TestFSMkdir_NoError(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_mkdir", map[string]any{
		"parent": "/Freebox/Downloads",
		"name":   "test-dir",
	})
	if req.IsError {
		t.Errorf("unexpected error: %v", req.Content)
	}
}

func TestFSDelete_NoError(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_delete", map[string]any{
		"path": "/Freebox/Downloads/old-file.iso",
	})
	if req.IsError {
		t.Errorf("unexpected error: %v", req.Content)
	}
}
