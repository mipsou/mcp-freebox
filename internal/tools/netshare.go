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

// SambaConfig reflects GET /api/v4/share/samba/
type SambaConfig struct {
	Enabled       bool   `json:"enabled"`
	LogonEnabled  bool   `json:"logon_enabled"`
	WorkGroup     string `json:"work_group"`
	PrinterEnabled bool  `json:"printer_enabled"`
}

// AFPConfig reflects GET /api/v4/share/afp/
type AFPConfig struct {
	Enabled bool `json:"enabled"`
}

// SambaShare reflects one entry from GET /api/v4/share/samba/share/
type SambaShare struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	ReadOnly bool   `json:"readonly"`
}

func registerNetshare(s *server.MCPServer, c writer) {
	// ── Samba — config globale ────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_samba_config",
			mcp.WithDescription("État et configuration du serveur Samba de la Freebox : activé, workgroup, logon, imprimante. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg SambaConfig
			if err := c.Get(ctx, "/share/samba/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Samba — partages ─────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_samba_shares",
			mcp.WithDescription("Liste les partages Samba configurés sur la Freebox (nom, chemin, lecture seule). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var shares []SambaShare
			if err := c.Get(ctx, "/share/samba/share/", &shares); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(shares)
		},
	)

	// ── AFP — config ─────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_afp_config",
			mcp.WithDescription("État du serveur AFP (Apple Filing Protocol) de la Freebox : activé ou non. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg AFPConfig
			if err := c.Get(ctx, "/share/afp/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)
}
