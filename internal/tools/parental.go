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

// ParentalConfig reflects GET /api/v4/parental/config/
type ParentalConfig struct {
	Enabled        bool   `json:"enabled"`
	DefaultPolicy  string `json:"default_policy"`  // allow | deny
}

// ParentalPlanning reflects one entry from GET /api/v4/parental/planning/
// Each entry is a time slot with a policy override.
type ParentalPlanning struct {
	ID     int    `json:"id"`
	Day    int    `json:"day"`    // 0=lundi … 6=dimanche
	Start  int    `json:"start"`  // minutes depuis minuit
	End    int    `json:"end"`    // minutes depuis minuit
	Policy string `json:"policy"` // allow | deny
}

// ParentalFilter reflects one entry from GET /api/v4/parental/filter/
type ParentalFilter struct {
	ID      int    `json:"id"`
	MACAddr string `json:"mac_addr"`
	Comment string `json:"comment"`
	Enabled bool   `json:"enabled"`
}

func registerParental(s *server.MCPServer, c writer) {
	// ── Config globale ────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_parental_config",
			mcp.WithDescription("Configuration du contrôle parental de la Freebox : activé, politique par défaut (allow/deny). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg ParentalConfig
			if err := c.Get(ctx, "/parental/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Planning horaire ─────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_parental_planning",
			mcp.WithDescription("Plages horaires du contrôle parental : jour, heure début/fin (en minutes depuis minuit), politique (allow/deny). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var slots []ParentalPlanning
			if err := c.Get(ctx, "/parental/planning/", &slots); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(slots)
		},
	)

	// ── Filtres MAC ───────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_parental_filters",
			mcp.WithDescription("Liste des appareils soumis au contrôle parental (filtre par adresse MAC). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var filters []ParentalFilter
			if err := c.Get(ctx, "/parental/filter/", &filters); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(filters)
		},
	)
}
