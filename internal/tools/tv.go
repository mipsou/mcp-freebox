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

// TVChannel reflects one entry from GET /api/v4/tv/channels/
type TVChannel struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Number  int    `json:"number"`
	Quality string `json:"quality"` // sd | hd | uhd
	Logo    string `json:"logo_url"`
}

// TVRecord reflects one entry from GET /api/v4/pvr/programmed/
// status : started | scheduled | finished | failed | disabled
type TVRecord struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	StartTime int64  `json:"start"`
	EndTime   int64  `json:"end"`
	ChannelID string `json:"channel_uuid"`
	Status    string `json:"status"`
	Error     string `json:"error"`
}

func registerTV(s *server.MCPServer, c writer) {
	// ── Chaînes TV ────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_tv_channels",
			mcp.WithDescription("Liste les chaînes TV disponibles sur la Freebox (nom, numéro, qualité SD/HD/UHD). Nécessite l'abonnement TV Freebox. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var channels []TVChannel
			if err := c.Get(ctx, "/tv/channels/", &channels); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(channels)
		},
	)

	// ── Enregistrements programmés ────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_tv_records",
			mcp.WithDescription("Liste les enregistrements TV programmés et leur état (scheduled/started/finished/failed). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var records []TVRecord
			if err := c.Get(ctx, "/pvr/programmed/", &records); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(records)
		},
	)
}
