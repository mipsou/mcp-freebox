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

// DHCPConfig reflects GET /api/v4/dhcp/config/
type DHCPConfig struct {
	Enabled         bool     `json:"enabled"`
	StickyAssign    bool     `json:"sticky_assign"`    // toujours attribuer la même IP à un hôte donné
	GatewayIP       string   `json:"gateway"`          // lecture seule
	NetmaskIP       string   `json:"netmask"`          // lecture seule
	IPRangeStart    string   `json:"ip_range_start"`
	IPRangeEnd      string   `json:"ip_range_end"`
	AlwaysBroadcast bool     `json:"always_broadcast"`
	DNSServers      []string `json:"dns"`
}

func registerDHCPConfig(s *server.MCPServer, c getter) {
	// ── Config serveur DHCP ───────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_config",
			mcp.WithDescription("Configuration du serveur DHCP de la Freebox : activé, plage d'adresses (start/end), passerelle, masque, serveurs DNS. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg DHCPConfig
			if err := c.Get(ctx, "/dhcp/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)
}
