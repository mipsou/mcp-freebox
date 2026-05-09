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

// SambaConfig reflects GET /api/v4/netshare/samba/
type SambaConfig struct {
	FileShareEnabled  bool   `json:"file_share_enabled"`  // partage de fichiers activé
	PrintShareEnabled bool   `json:"print_share_enabled"` // partage d'imprimante activé
	LogonEnabled      bool   `json:"logon_enabled"`       // authentification requise
	LogonUser         string `json:"logon_user"`          // identifiant Samba
	Workgroup         string `json:"workgroup"`           // nom du workgroup
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

func registerNetshare(s *server.MCPServer, c writer) {
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

	// ── Samba — modifier config ───────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_samba_config_set",
			mcp.WithDescription("Modifie la configuration du serveur Samba (partage de fichiers, imprimante, authentification, workgroup). Patch partiel : seuls les champs fournis sont mis à jour. Nécessite la permission 'settings'."),
			mcp.WithBoolean("file_share_enabled",
				mcp.Description("Activer ou désactiver le partage de fichiers SMB")),
			mcp.WithBoolean("print_share_enabled",
				mcp.Description("Activer ou désactiver le partage d'imprimante SMB")),
			mcp.WithBoolean("logon_enabled",
				mcp.Description("Requérir une authentification pour accéder aux partages")),
			mcp.WithString("logon_user",
				mcp.Description("Identifiant Samba utilisé pour l'authentification")),
			mcp.WithString("workgroup",
				mcp.Description("Nom du workgroup SMB (ex: WORKGROUP)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			patch := map[string]any{}
			if v, ok := args["file_share_enabled"].(bool); ok {
				patch["file_share_enabled"] = v
			}
			if v, ok := args["print_share_enabled"].(bool); ok {
				patch["print_share_enabled"] = v
			}
			if v, ok := args["logon_enabled"].(bool); ok {
				patch["logon_enabled"] = v
			}
			if v, ok := args["logon_user"].(string); ok {
				patch["logon_user"] = v
			}
			if v, ok := args["workgroup"].(string); ok && v != "" {
				patch["workgroup"] = v
			}
			if len(patch) == 0 {
				return mcp.NewToolResultError("aucun champ à modifier (fournir au moins un paramètre)"), nil
			}
			var updated SambaConfig
			if err := c.Put(ctx, "/netshare/samba/", patch, &updated); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("freebox_samba_config_set : %v", err)), nil
			}
			return jsonResult(updated)
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
