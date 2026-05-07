/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ParentalConfig reflects GET /api/v4/parental/config/
type ParentalConfig struct {
	DefaultFilterMode string `json:"default_filter_mode"` // allowed | denied | webonly
}

// ParentalFilterPlanning reflects GET /api/v4/parental/filter/{id}/planning
type ParentalFilterPlanning struct {
	Resolution  int      `json:"resolution"`   // nombre de slots par jour (ex: 48 = tranches 30 min)
	CDayRanges  []string `json:"cdayranges"`   // plages personnalisées (ex: ":fr_bank_holidays")
	Mapping     []string `json:"mapping"`      // état par slot : "allowed" | "denied" | "webonly"
}

// ParentalFilter reflects one entry from GET /api/v4/parental/filter/
type ParentalFilter struct {
	ID              int      `json:"id"`
	Macs            []string `json:"macs"`              // adresses MAC concernées
	Hosts           []string `json:"hosts"`             // noms d'hôtes associés (lecture seule)
	Desc            string   `json:"desc"`              // description du filtre
	Forced          bool     `json:"forced"`            // ignorer le planning
	ForcedMode      string   `json:"forced_mode"`       // état si forced=true
	TmpMode         string   `json:"tmp_mode"`          // état temporaire en cours
	TmpModeExpire   int      `json:"tmp_mode_expire"`   // secondes avant fin du mode temporaire
	SchedulingMode  string   `json:"scheduling_mode"`   // forced | temporary | planning (lecture seule)
	FilterState     string   `json:"filter_state"`      // allowed | denied | webonly (lecture seule)
}

func registerParental(s *server.MCPServer, c getter) {
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

	// ── Planning horaire d'un filtre ─────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_parental_planning",
			mcp.WithDescription("Planning horaire d'un filtre de contrôle parental : résolution, plages personnalisées, état par slot (allowed/denied/webonly). Lecture seule."),
			mcp.WithNumber("filter_id",
				mcp.Required(),
				mcp.Description("ID du filtre parental (voir freebox_parental_filters)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("filter_id", 0)
			var planning ParentalFilterPlanning
			if err := c.Get(ctx, fmt.Sprintf("/parental/filter/%d/planning", id), &planning); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(planning)
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
