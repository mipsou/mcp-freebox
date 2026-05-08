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

// FtpConfig reflects GET /api/v4/ftp/config/
type FtpConfig struct {
	Enabled             bool   `json:"enabled"`
	AllowAnonymous      bool   `json:"allow_anonymous"`       // connexion anonyme autorisée
	AllowAnonymousWrite bool   `json:"allow_anonymous_write"` // écriture anonyme autorisée
	AllowRemoteAccess   bool   `json:"allow_remote_access"`   // accès FTP distant activé
	WeakPassword        bool   `json:"weak_password"`         // mot de passe faible (lecture seule)
	PortCtrl            int    `json:"port_ctrl"`             // port de contrôle (accès distant)
	PortData            int    `json:"port_data"`             // port de données (accès distant)
	RemoteDomain        string `json:"remote_domain"`         // domaine pour l'accès distant
}

func registerFTP(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_ftp_config",
			mcp.WithDescription("Configuration du serveur FTP de la Freebox : activé, accès anonyme, accès distant (ports, domaine), solidité du mot de passe. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg FtpConfig
			if err := c.Get(ctx, "/ftp/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)
}
