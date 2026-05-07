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

// LanConfig reflects GET /api/v4/lan/config/
type LanConfig struct {
	IP          string `json:"ip"`
	Name        string `json:"name"`
	NameDNS     string `json:"name_dns"`
	NameMDNS    string `json:"name_mdns"`
	NameNetbios string `json:"name_netbios"`
	Type        string `json:"type"` // router | bridge
}

// LanHostUpdate is the body for PUT /api/v4/lan/browser/pub/{id}
type LanHostUpdate struct {
	PrimaryName string `json:"primary_name,omitempty"`
}

func registerLANConfig(s *server.MCPServer, c writer) {
	// ── Config LAN ────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_lan_config",
			mcp.WithDescription("Configuration du réseau LAN de la Freebox : IP, masque, nom DNS/mDNS/NetBIOS, mode (router/bridge). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg LanConfig
			if err := c.Get(ctx, "/lan/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Renommer un équipement ────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_lan_host_rename",
			mcp.WithDescription("Renomme un équipement du réseau local (modifie le nom affiché dans l'interface Freebox et retourné par freebox_lan_hosts)."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("ID de l'équipement (champ 'id' dans freebox_lan_hosts)")),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Nouveau nom de l'équipement (ex: CoreOS-11)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetString("id", "")
			body := LanHostUpdate{PrimaryName: req.GetString("name", "")}
			var updated LanHost
			if err := c.Put(ctx, fmt.Sprintf("/lan/browser/pub/%s", id), body, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)
}
