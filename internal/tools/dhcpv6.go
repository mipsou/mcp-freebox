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

// DHCPv6Config reflects GET /api/v4/dhcpv6/config/
type DHCPv6Config struct {
	Enabled      bool     `json:"enabled"`
	UseCustomDNS bool     `json:"use_custom_dns"` // utiliser des serveurs DNS personnalisés
	DNS          []string `json:"dns"`            // liste des serveurs DNS IPv6
}

func registerDHCPv6(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_dhcpv6_config",
			mcp.WithDescription("Configuration du serveur DHCPv6 de la Freebox : activé, DNS personnalisés (IPv6). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg DHCPv6Config
			if err := c.Get(ctx, "/dhcpv6/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)
}
