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

// ConnectionConfig reflects GET /api/v4/connection/config/
type ConnectionConfig struct {
	Ping             bool   `json:"ping"`
	IsSecurePass     bool   `json:"is_secure_pass"`
	RemoteAccess     bool   `json:"remote_access"`
	RemoteAccessPort int    `json:"remote_access_port"`
	RemoteAccessIP   string `json:"remote_access_ip"`
	WakeOnLanPort    int    `json:"wol_port"`
	AdblockEnabled   bool   `json:"adblock_enabled"`
	AdblockMode      string `json:"adblock_mode"`
}

func registerConnectionConfig(s *server.MCPServer, c writer) {
	// ── Lire ─────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_connection_config",
			mcp.WithDescription("Lit la configuration de la connexion WAN de la Freebox : réponse au ping, accès distant (IP/port), Wake-on-LAN, blocage de publicités. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg ConnectionConfig
			if err := c.Get(ctx, "/connection/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Modifier ──────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_connection_config_set",
			mcp.WithDescription("Modifie la configuration de la connexion WAN de la Freebox. Seuls les champs fournis sont mis à jour (patch partiel)."),
			mcp.WithBoolean("ping",
				mcp.Description("Autoriser les réponses au ping WAN")),
			mcp.WithBoolean("remote_access",
				mcp.Description("Activer l'accès distant HTTPS à la Freebox")),
			mcp.WithNumber("remote_access_port",
				mcp.Description("Port HTTPS pour l'accès distant (1–65535, ex: 443)"),
				mcp.Min(1), mcp.Max(65535)),
			mcp.WithNumber("wol_port",
				mcp.Description("Port UDP Wake-on-LAN (1–65535, ex: 9)"),
				mcp.Min(1), mcp.Max(65535)),
			mcp.WithBoolean("adblock_enabled",
				mcp.Description("Activer le blocage de publicités DNS")),
			mcp.WithString("adblock_mode",
				mcp.Description("Mode de blocage publicitaire : all ou custom"),
				mcp.Enum("all", "custom")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()
			patch := map[string]any{}
			if v, ok := args["ping"].(bool); ok {
				patch["ping"] = v
			}
			if v, ok := args["remote_access"].(bool); ok {
				patch["remote_access"] = v
			}
			if v, ok := args["remote_access_port"]; ok && v != nil {
				patch["remote_access_port"] = int(toFloat(v))
			}
			if v, ok := args["wol_port"]; ok && v != nil {
				patch["wol_port"] = int(toFloat(v))
			}
			if v, ok := args["adblock_enabled"].(bool); ok {
				patch["adblock_enabled"] = v
			}
			if v, ok := args["adblock_mode"].(string); ok && v != "" {
				patch["adblock_mode"] = v
			}
			if len(patch) == 0 {
				return mcp.NewToolResultError("aucun champ à modifier (fournir au moins un paramètre)"), nil
			}
			var updated ConnectionConfig
			if err := c.Put(ctx, "/connection/config/", patch, &updated); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("freebox_connection_config_set : %v", err)), nil
			}
			return jsonResult(updated)
		},
	)
}
