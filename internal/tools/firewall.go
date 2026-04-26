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

// FirewallIncomingRule reflects one entry from GET /api/v4/fw/incoming/
type FirewallIncomingRule struct {
	ID       int    `json:"id"`
	Enabled  bool   `json:"enabled"`
	Comment  string `json:"comment"`
	Action   string `json:"action"`    // accept, deny, drop
	IPProto  string `json:"ip_proto"`  // tcp, udp, icmp, all
	SrcIP    string `json:"src_ip"`    // source IP/CIDR, "" = any
	DstPort  int    `json:"dst_port"`  // destination port, 0 = any
	SrcPort  int    `json:"src_port"`  // source port, 0 = any
}

// DMZConfig reflects GET /api/v4/fw/dmz/
type DMZConfig struct {
	Enabled bool   `json:"enabled"`
	IP      string `json:"ip"`
}

func registerFirewall(s *server.MCPServer, c getter) {
	// ── Règles entrantes ─────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_firewall_incoming",
			mcp.WithDescription("Liste les règles de pare-feu personnalisées pour le trafic entrant (WAN → LAN). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var rules []FirewallIncomingRule
			if err := c.Get(ctx, "/fw/incoming/", &rules); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(rules)
		},
	)

	// ── DMZ ──────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_firewall_dmz",
			mcp.WithDescription("Configuration DMZ de la Freebox : état (activé/désactivé) et IP de l'hôte DMZ. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg DMZConfig
			if err := c.Get(ctx, "/fw/dmz/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)
}
