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

// L2Ident reflects the layer-2 identifier of a LAN host.
type L2Ident struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// L3Connectivity reflects one IP address assigned to a LAN host.
type L3Connectivity struct {
	Addr              string `json:"addr"`
	AF                string `json:"af"`
	Active            bool   `json:"active"`
	Reachable         bool   `json:"reachable"`
	LastTimeReachable int64  `json:"last_time_reachable"`
}

// LanHost reflects one entry from GET /api/v4/lan/browser/pub/
type LanHost struct {
	ID                string           `json:"id"`
	PrimaryName       string           `json:"primary_name"`
	HostType          string           `json:"host_type"`
	PrimaryNameManual bool             `json:"primary_name_manual"`
	L2Ident           []L2Ident        `json:"l2ident"`
	VendorName        string           `json:"vendor_name"`
	Persistent        bool             `json:"persistent"`
	Reachable         bool             `json:"reachable"`
	LastTimeReachable int64            `json:"last_time_reachable"`
	Active            bool             `json:"active"`
	L3Connectivities  []L3Connectivity `json:"l3connectivities"`
}

// LanInterface reflects une interface LAN listée par /lan/browser/interfaces/.
// La Freebox expose typiquement deux interfaces : "pub" (LAN principal) et
// "guest" (réseau invité), avec le nombre d'hôtes vus sur chacune.
type LanInterface struct {
	Name      string `json:"name"`
	HostCount int    `json:"host_count"`
}

func registerLAN(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_lan_hosts",
			mcp.WithDescription("Liste les équipements présents sur le réseau local (LAN) : nom, MAC, IP, type d'hôte, accessibilité."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var hosts []LanHost
			if err := c.Get(ctx, "/lan/browser/pub/", &hosts); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(hosts)
		},
	)

	s.AddTool(
		mcp.NewTool("freebox_lan_interfaces",
			mcp.WithDescription("Liste les interfaces LAN exposées par la Freebox (par exemple 'pub' = LAN principal, 'guest' = réseau invité) avec le nombre d'hôtes vus sur chacune."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var ifaces []LanInterface
			if err := c.Get(ctx, "/lan/browser/interfaces/", &ifaces); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(ifaces)
		},
	)
}
