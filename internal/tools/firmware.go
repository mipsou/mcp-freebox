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

// FirmwareUpdate reflects GET /api/v4/system/update/
type FirmwareUpdate struct {
	UpdateAvailable bool   `json:"update_available"`
	Version         string `json:"version"`
}

func registerFirmware(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_firmware_update_status",
			mcp.WithDescription("Vérifie si une mise à jour du firmware Freebox est disponible et la version proposée. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var fw FirmwareUpdate
			if err := c.Get(ctx, "/system/update/", &fw); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(fw)
		},
	)
}
