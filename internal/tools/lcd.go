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

// LCDConfig reflects GET /api/v4/lcd/config/
type LCDConfig struct {
	Brightness        int  `json:"brightness"` // 0-100
	OrientationForced bool `json:"orientation_forced"`
	Orientation       int  `json:"orientation"` // degrés
}

func registerLCD(s *server.MCPServer, c writer) {
	// ── État LCD ──────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_lcd_config",
			mcp.WithDescription("Configuration de l'écran LCD de la Freebox Delta : luminosité (0-100), orientation. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg LCDConfig
			if err := c.Get(ctx, "/lcd/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Régler luminosité ─────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_lcd_brightness",
			mcp.WithDescription("Règle la luminosité de l'écran LCD de la Freebox (0 = éteint, 100 = maximum)."),
			mcp.WithNumber("brightness",
				mcp.Required(),
				mcp.Description("Luminosité de 0 à 100")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			brightness := req.GetInt("brightness", 50)
			body := map[string]any{"brightness": brightness}
			var updated LCDConfig
			if err := c.Put(ctx, "/lcd/config/", body, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)
}
