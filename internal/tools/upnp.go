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

// UPnPConfig reflects GET /api/v4/upnp/config/
type UPnPConfig struct {
	Enabled bool `json:"enabled"`
}

// UPnPIGDMapping reflects one entry from GET /api/v4/upnp/igd/rules/
// An IGD rule is a port forwarding created by a UPnP client.
type UPnPIGDMapping struct {
	ID           int    `json:"id"`
	Enabled      bool   `json:"enabled"`
	ExternalIP   string `json:"ext_ip"`
	ExternalPort int    `json:"ext_port"`
	InternalIP   string `json:"int_ip"`
	InternalPort int    `json:"int_port"`
	Protocol     string `json:"proto"` // tcp | udp
	Description  string `json:"desc"`
	Duration     int    `json:"duration"` // secondes, 0 = permanent
}

func registerUPnP(s *server.MCPServer, c getter) {
	// ── Config UPnP ───────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_upnp_config",
			mcp.WithDescription("Configuration UPnP/IGD de la Freebox : activé ou désactivé. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg UPnPConfig
			if err := c.Get(ctx, "/upnp/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Mappings IGD actifs ────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_upnp_rules",
			mcp.WithDescription("Liste les règles de port forwarding créées automatiquement par UPnP/IGD (jeux, Plex, etc.) : port externe, IP/port interne, protocole, description. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var rules []UPnPIGDMapping
			if err := c.Get(ctx, "/upnp/igd/rules/", &rules); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(rules)
		},
	)
}
