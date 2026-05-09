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
			{
				ID: 1, DisplayName: "WD Elements 1 TB", State: "enabled", TotalBytes: 1000204886016,
				Partitions: []StoragePartition{
					{ID: 3, DiskID: 1, Label: "Disque dur", Fstype: "ext4", State: "mounted", TotalBytes: 245091500032},
				},
			},
		},
	})
	result := callTool(t, s, "freebox_storage_disks")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"display_name": "WD Elements 1 TB"`) {
		t.Errorf("unexpected result: %s", text)
	}
	if !strings.Contains(text, `"fstype": "ext4"`) {
		t.Errorf("embedded partitions missing: %s", text)
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
			{ID: 1, DiskID: 1, Fstype: "ext4", State: "mounted", Path: "/mnt/data", FreeBytes: 500000000000},
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

func TestStorageRAID_OK(t *testing.T) {
	s := newStorageServer(t, mockGetter{
		"/storage/raid/": []StorageRAID{
			{ID: 1, Name: "raid0", State: "ok", Level: "raid1"},
		},
	})
	result := callTool(t, s, "freebox_storage_raid")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"level": "raid1"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestStorageRAID_NullEmptyArray(t *testing.T) {
	// Quand la Freebox n'a aucun RAID configure, /storage/raid/ retourne null
	// dans le wrapper (success=true, result=null). Le mock simule ca via
	// une slice vide ; le tool doit retourner "[]" pas "null".
	s := newStorageServer(t, mockGetter{"/storage/raid/": []StorageRAID{}})
	result := callTool(t, s, "freebox_storage_raid")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if result.Content[0].(mcp.TextContent).Text != "[]" {
		t.Errorf("expected empty array, got: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestStorageRAID_APIError(t *testing.T) {
	s := newStorageServer(t, mockGetter{"/storage/disk/": []StorageDisk{}})
	result := callTool(t, s, "freebox_storage_raid")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
