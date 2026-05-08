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

// TftpConfig reflects GET /api/v4/tftp/config/
// Ajouté en firmware 4.9.15 (janvier 2026) — non documenté dans le SDK officiel.
type TftpConfig struct {
	Enabled bool   `json:"enabled"`
	Root    string `json:"root"` // chemin racine encodé en base64 standard (ex : L0ZyZWVib3gvdGZ0cA== → /Freebox/tftp)
}

func registerTFTP(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_tftp_config",
			mcp.WithDescription("Configuration du serveur TFTP de la Freebox : activé ou non, répertoire racine (base64). Fonctionnalité ajoutée en firmware 4.9.15 (janvier 2026). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg TftpConfig
			if err := c.Get(ctx, "/tftp/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)
}
