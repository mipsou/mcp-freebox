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

func newStorageServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerStorage(s, mock)
	return s
}

func TestStorageDisks_OK(t *testing.T) {
	s := newStorageServer(t, mockGetter{
		"/storage/disk/": []StorageDisk{
			{ID: "sda", DisplayName: "WD Elements 1 TB", State: "enabled", TotalBytes: 1000204886016},
		},
	})
	result := callTool(t, s, "freebox_storage_disks")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"display_name": "WD Elements 1 TB"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestStorageDisks_APIError(t *testing.T) {
	s := newStorageServer(t, mockGetter{})
	result := callTool(t, s, "freebox_storage_disks")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestStoragePartitions_OK(t *testing.T) {
	s := newStorageServer(t, mockGetter{
		"/storage/partition/": []StoragePartition{
			{ID: 1, DiskID: "sda", Fstype: "ext4", State: "mounted", Path: "/mnt/data", FreeBytes: 500000000000},
		},
	})
	result := callTool(t, s, "freebox_storage_partitions")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"path": "/mnt/data"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestStoragePartitions_APIError(t *testing.T) {
	s := newStorageServer(t, mockGetter{"/storage/disk/": []StorageDisk{}})
	result := callTool(t, s, "freebox_storage_partitions")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
