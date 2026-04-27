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
	GatewayIP       string   `json:"gateway"`
	NetmaskIP       string   `json:"netmask"`
	IPRangeStart    string   `json:"ip_range_start"`
	IPRangeEnd      string   `json:"ip_range_end"`
	DNSServers      []string `json:"dns"`
	AlwaysBroadcast bool     `json:"always_broadcast"`
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
