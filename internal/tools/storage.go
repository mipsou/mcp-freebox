/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// StorageDisk reflects one entry from GET /api/v4/storage/disk/
// connector is an integer enum (0=unknown, 1=USB, 2=eSATA, 3=PCIe, …).
// id is numeric per the Freebox OS API spec.
// partitions is an array of embedded StoragePartition objects (not IDs).
type StorageDisk struct {
	ID          int64              `json:"id"`
	Type        string             `json:"type"`
	Connector   int                `json:"connector"`
	State       string             `json:"state"`
	TotalBytes  int64              `json:"total_bytes"`
	Idle        bool               `json:"idle"`
	Spinning    bool               `json:"spinning"`
	TableType   string             `json:"table_type"`
	DisplayName string             `json:"display_name"`
	Partitions  []StoragePartition `json:"partitions"`
}

// StoragePartition reflects one entry from GET /api/v4/storage/partition/
// id and disk_id are numeric per the Freebox OS API spec.
type StoragePartition struct {
	ID         int64  `json:"id"`
	DiskID     int64  `json:"disk_id"`
	Fstype     string `json:"fstype"`
	Label      string `json:"label"`
	State      string `json:"state"`
	Path       string `json:"path"`
	TotalBytes int64  `json:"total_bytes"`
	FreeBytes  int64  `json:"free_bytes"`
	UsedBytes  int64  `json:"used_bytes"`
}

func registerStorage(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_storage_disks",
			mcp.WithDescription("Liste les disques connectés à la Freebox : type, état, taille, connecteur."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var disks []StorageDisk
			if err := c.Get(ctx, "/storage/disk/", &disks); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(disks)
		},
	)

	s.AddTool(
		mcp.NewTool("freebox_storage_partitions",
			mcp.WithDescription("Liste les partitions des disques Freebox : système de fichiers, état de montage, espace libre/utilisé."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var parts []StoragePartition
			if err := c.Get(ctx, "/storage/partition/", &parts); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(parts)
		},
	)
}
