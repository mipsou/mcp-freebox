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

// WifiPlanning reflects GET /api/v15/wifi/planning/.
// Mapping = grille hebdomadaire d'allumage WiFi : 7 jours × 24 h × Resolution
// créneaux (typiquement Resolution=2 → demi-heures, soit 7×48 = 336 entrées
// observées en pratique 24×7 = 168 ; chaque entrée vaut "on" ou "off").
// UsePlanning = true active le planning, false laisse le WiFi toujours allumé.
type WifiPlanning struct {
	UsePlanning bool     `json:"use_planning"`
	Resolution  int      `json:"resolution"`
	Mapping     []string `json:"mapping"`
}

func registerWifiPlanning(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_wifi_planning",
			mcp.WithDescription("Planning d'allumage automatique du WiFi (grille hebdomadaire de créneaux on/off)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var p WifiPlanning
			if err := c.Get(ctx, "/wifi/planning/", &p); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(p)
		},
	)
}
