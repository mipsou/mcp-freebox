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

// IPv6Delegation reflects un préfixe IPv6 délégué (DHCPv6-PD).
type IPv6Delegation struct {
	Prefix  string `json:"prefix"`
	NextHop string `json:"next_hop"`
}

// ConnectionIPv6Config reflects GET /api/v15/connection/ipv6/config/.
// IPv6Firewall = pare-feu IPv6 actif sur la box. IPv6PrefixFirewall = pare-feu
// par préfixe délégué (filtrage entre LAN et délégations).
type ConnectionIPv6Config struct {
	IPv6Enabled        bool             `json:"ipv6_enabled"`
	IPv6Firewall       bool             `json:"ipv6_firewall"`
	IPv6PrefixFirewall bool             `json:"ipv6_prefix_firewall"`
	IPv6LL             string           `json:"ipv6ll"`
	Delegations        []IPv6Delegation `json:"delegations"`
}

func registerConnectionIPv6(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_connection_ipv6_config",
			mcp.WithDescription("Configuration IPv6 de la connexion : activation, pare-feu, link-local, préfixes délégués (DHCPv6-PD)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg ConnectionIPv6Config
			if err := c.Get(ctx, "/connection/ipv6/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)
}
