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

// SystemSensor reflects one sensor entry (CPU temp, switch temp, etc.)
type SystemSensor struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"` // °C
}

// SystemFan reflects one fan entry.
type SystemFan struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"` // RPM
}

// SystemInfo reflects GET /api/v4/system/
type SystemInfo struct {
	MAC             string         `json:"mac"`
	Serial          string         `json:"serial"`
	Uptime          string         `json:"uptime"`
	UptimeVal       int64          `json:"uptime_val"` // seconds
	BoardName       string         `json:"board_name"`
	FirmwareVersion string         `json:"firmware_version"`
	DiskStatus      string         `json:"disk_status"`
	BoxAuthenticated bool          `json:"box_authenticated"`
	Sensors         []SystemSensor `json:"sensors"`
	Fans            []SystemFan    `json:"fans"`
}

func registerSystem(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_system",
			mcp.WithDescription("Informations système Freebox : uptime, version firmware, températures (CPU, switch), vitesse ventilateurs."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var info SystemInfo
			if err := c.Get(ctx, "/system/", &info); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(info)
		},
	)
}
