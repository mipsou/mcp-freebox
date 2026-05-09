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

// ConnectionLogEntry reflects one entry from GET /api/v15/connection/logs/.
// Type "link" = transition d'etat ligne (up/down). BwDown/BwUp en bps.
type ConnectionLogEntry struct {
	ID     int    `json:"id"`
	Date   int64  `json:"date"`
	Type   string `json:"type"`  // "link"
	State  string `json:"state"` // "up" | "down"
	Link   string `json:"link"`  // "ftth" | "xdsl" | "ethernet" | "lte"
	BwDown int64  `json:"bw_down"`
	BwUp   int64  `json:"bw_up"`
}

func registerConnectionLogs(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_connection_logs",
			mcp.WithDescription("Historique des évènements de connexion WAN (montées/descentes de ligne, débits négociés). Utile pour PRA et diagnostic stabilité ligne."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var entries []ConnectionLogEntry
			if err := c.Get(ctx, "/connection/logs/", &entries); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(entries)
		},
	)
}
