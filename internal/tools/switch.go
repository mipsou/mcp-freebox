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

// SwitchMacEntry reflects a MAC address seen on a switch port.
type SwitchMacEntry struct {
	MacAddr  string `json:"mac_addr"`
	Hostname string `json:"hostname"`
}

// SwitchPortStatus reflects one entry from GET /api/v4/switch/status/
type SwitchPortStatus struct {
	ID      int              `json:"id"`
	Link    bool             `json:"link"`
	Speed   string           `json:"speed"`   // "10", "100", "1000" (Mbps)
	Duplex  string           `json:"duplex"`  // "full", "half"
	MacList []SwitchMacEntry `json:"mac_list"`
}

func registerSwitch(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_switch_ports",
			mcp.WithDescription("État des ports du switch Freebox : lien, vitesse (Mbps), duplex, équipements connectés (MAC + hostname)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var ports []SwitchPortStatus
			if err := c.Get(ctx, "/switch/status/", &ports); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(ports)
		},
	)
}
