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

// AirMediaConfig reflects GET /api/v4/airmedia/config/
type AirMediaConfig struct {
	Enabled           bool   `json:"enabled"`
	Password          string `json:"password"`
}

// AirMediaReceiver reflects one entry from GET /api/v4/airmedia/receivers/
type AirMediaReceiver struct {
	Name              string   `json:"name"`
	PasswordProtected bool     `json:"password_protected"`
	Capabilities      []string `json:"capabilities"` // photo, video, audio, screen
}

func registerAirMedia(s *server.MCPServer, c getter) {
	// ── Configuration AirMedia ────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_airmedia_config",
			mcp.WithDescription("Lit la configuration AirMedia de la Freebox : activé/désactivé, présence d'un mot de passe. AirMedia permet de diffuser photo/vidéo/audio depuis un appareil Apple vers la Freebox."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg AirMediaConfig
			if err := c.Get(ctx, "/airmedia/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Liste des récepteurs AirMedia ─────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_airmedia_receivers",
			mcp.WithDescription("Liste les récepteurs AirMedia disponibles sur le réseau (Freebox Player, TV…) avec leurs capacités (photo/vidéo/audio/screen)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var receivers []AirMediaReceiver
			if err := c.Get(ctx, "/airmedia/receivers/", &receivers); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(receivers)
		},
	)
}
