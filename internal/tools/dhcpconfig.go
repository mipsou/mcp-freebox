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

// DHCPOption reflects one entry in the options[] array of DhcpConfig.
// Each option maps a DHCP option identifier (RFC2132) to its value.
// Known identifiers: "tftp_server_name" (opt 66), "bootfile_name" (opt 67).
type DHCPOption struct {
	ID  string `json:"id"`
	Val string `json:"val"`
}

// DHCPConfig reflects GET /api/v4/dhcp/config/
type DHCPConfig struct {
	Enabled                bool         `json:"enabled"`
	StickyAssign           bool         `json:"sticky_assign"`              // toujours attribuer la même IP à un hôte donné
	BootServer             string       `json:"boot_server"`                // serveur TFTP / next-server BOOTP (siaddr)
	BootFile               string       `json:"boot_file"`                  // fichier de boot (option 67)
	GatewayIP              string       `json:"gateway"`                    // lecture seule
	NetmaskIP              string       `json:"netmask"`                    // lecture seule
	IPRangeStart           string       `json:"ip_range_start"`
	IPRangeEnd             string       `json:"ip_range_end"`
	AlwaysBroadcast        bool         `json:"always_broadcast"`
	IgnoreOutOfRangeHint   bool         `json:"ignore_out_of_range_hint"`   // ignorer hint hors plage
	DNSServers             []string     `json:"dns"`
	Options                []DHCPOption `json:"options"`                    // options DHCP personnalisées RFC2132
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
