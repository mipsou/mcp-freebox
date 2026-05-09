/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// L2Ident reflects the layer-2 identifier of a LAN host.
type L2Ident struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// L2Idents tolère le quirk firmware 4.9.18.1 où /lan/browser/pub/ peut renvoyer
// `l2ident` soit comme un tableau `[{id,type}]` soit comme un objet single
// `{id,type}` (un host n'a typiquement qu'un seul identifiant L2 → l'API
// déballe parfois). Custom unmarshaler à l'image de BindUSBPorts (#76).
type L2Idents []L2Ident

func (l *L2Idents) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*l = nil
		return nil
	}
	// Object single
	if trimmed[0] == '{' {
		var single L2Ident
		if err := json.Unmarshal(trimmed, &single); err != nil {
			return fmt.Errorf("l2ident object: %w", err)
		}
		*l = L2Idents{single}
		return nil
	}
	// Array
	var arr []L2Ident
	if err := json.Unmarshal(trimmed, &arr); err != nil {
		return fmt.Errorf("l2ident array: %w", err)
	}
	*l = arr
	return nil
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
	L2Ident           L2Idents         `json:"l2ident"`
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
