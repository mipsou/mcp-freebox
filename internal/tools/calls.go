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

// CallEntry reflects one entry from GET /api/v4/call/log/
// type     : accepted | missed | outgoing
// contact  : number or name if in phonebook
type CallEntry struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Number    string `json:"number"`
	Name      string `json:"name"`
	Duration  int    `json:"duration"`  // secondes
	Timestamp int64  `json:"datetime"`  // Unix timestamp
	New       bool   `json:"new"`       // non lu
}

func registerCalls(s *server.MCPServer, c writer) {
	// ── Journal d'appels ─────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_call_log",
			mcp.WithDescription("Journal des appels téléphoniques de la Freebox : appels entrants (acceptés et manqués) et sortants, avec numéro, nom, durée et horodatage. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var entries []CallEntry
			if err := c.Get(ctx, "/call/log/", &entries); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(entries)
		},
	)
}
