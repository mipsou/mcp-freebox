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

// SystemLog reflects one entry from GET /api/v4/system/log/
// Freebox OS system event log.
type SystemLog struct {
	Timestamp int64  `json:"timestamp"`
	Level     string `json:"level"` // info | warning | error | critical
	Message   string `json:"msg"`
	Category  string `json:"category"`
}

func registerSysLog(s *server.MCPServer, c writer) {
	s.AddTool(
		mcp.NewTool("freebox_system_log",
			mcp.WithDescription("Journal système de la Freebox : événements (démarrage, erreurs, mises à jour, connexions) avec horodatage, niveau (info/warning/error/critical) et catégorie. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var logs []SystemLog
			if err := c.Get(ctx, "/system/log/", &logs); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(logs)
		},
	)
}
