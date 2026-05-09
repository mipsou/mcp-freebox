/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WifiMacFilterEntry reflects une entrée de la liste filtre MAC WiFi
// (GET /api/v15/wifi/mac_filter/). Type peut être "whitelist" ou "blacklist".
// Host est laissé en RawMessage car la structure imbriquée varie selon que
// l'hôte a déjà été vu (LanHost-like) ou non.
type WifiMacFilterEntry struct {
	ID       string          `json:"id"`
	Mac      string          `json:"mac"`
	Type     string          `json:"type"` // whitelist | blacklist
	Comment  string          `json:"comment"`
	Hostname string          `json:"hostname"`
	Host     json.RawMessage `json:"host,omitempty"`
}

func registerWifiMacFilter(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_wifi_mac_filter",
			mcp.WithDescription("Liste les entrées du filtre MAC WiFi (whitelist/blacklist). Retourne les MAC, type, commentaire et l'hôte associé s'il a été vu sur le LAN."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var entries []WifiMacFilterEntry
			if err := c.Get(ctx, "/wifi/mac_filter/", &entries); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(entries)
		},
	)
}
