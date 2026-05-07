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

// SambaConfig reflects GET /api/v4/netshare/samba/
type SambaConfig struct {
	FileShareEnabled bool   `json:"file_share_enabled"` // partage de fichiers activé
	PrintShareEnabled bool  `json:"print_share_enabled"` // partage d'imprimante activé
	LogonEnabled     bool   `json:"logon_enabled"`       // authentification requise
	LogonUser        string `json:"logon_user"`          // identifiant Samba
	Workgroup        string `json:"workgroup"`           // nom du workgroup
}

// AFPConfig reflects GET /api/v4/netshare/afp/
type AFPConfig struct {
	Enabled    bool   `json:"enabled"`
	GuestAllow bool   `json:"guest_allow"` // accès invité autorisé
	ServerType string `json:"server_type"` // type d'affichage macOS (imac, macbook, etc.)
	LoginName  string `json:"login_name"`  // identifiant AFP
}

// SambaShare reflects one entry from GET /api/v4/netshare/samba/share/
type SambaShare struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	ReadOnly bool   `json:"readonly"`
}

func registerNetshare(s *server.MCPServer, c getter) {
	// ── Samba — config globale ────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_samba_config",
			mcp.WithDescription("État et configuration du serveur Samba de la Freebox : activé, workgroup, logon, imprimante. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg SambaConfig
			if err := c.Get(ctx, "/netshare/samba/", &cfg); err != nil {
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
			if err := c.Get(ctx, "/netshare/samba/share/", &shares); err != nil {
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
			if err := c.Get(ctx, "/netshare/afp/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)
}
