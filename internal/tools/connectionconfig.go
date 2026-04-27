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

func registerConnectionConfig(s *server.MCPServer, c getter) {
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
}
