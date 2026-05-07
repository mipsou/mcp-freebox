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

// DHCPOption reflects one entry in the options[] array of DhcpConfig.
// Each option maps a DHCP option identifier (RFC2132) to its value.
// Known identifiers: "tftp_server_name" (opt 66), "bootfile_name" (opt 67).
type DHCPOption struct {
	ID  string `json:"id"`
	Val string `json:"val"`
}

// DHCPConfig reflects GET /api/v4/dhcp/config/
type DHCPConfig struct {
	Enabled              bool         `json:"enabled"`
	StickyAssign         bool         `json:"sticky_assign"`            // toujours attribuer la même IP à un hôte donné
	BootServer           string       `json:"boot_server"`              // serveur TFTP / next-server BOOTP (siaddr)
	BootFile             string       `json:"boot_file"`                // fichier de boot (option 67)
	GatewayIP            string       `json:"gateway"`                  // lecture seule
	NetmaskIP            string       `json:"netmask"`                  // lecture seule
	IPRangeStart         string       `json:"ip_range_start"`
	IPRangeEnd           string       `json:"ip_range_end"`
	AlwaysBroadcast      bool         `json:"always_broadcast"`
	IgnoreOutOfRangeHint bool         `json:"ignore_out_of_range_hint"` // ignorer hint hors plage
	DNSServers           []string     `json:"dns"`
	Options              []DHCPOption `json:"options"`                  // options DHCP personnalisées RFC2132
}

// dhcpOptionsUpdate is the minimal payload accepted by PUT /api/v4/dhcp/config/
// for modifying only the custom options (partial update).
type dhcpOptionsUpdate struct {
	BootServer string       `json:"boot_server"`
	BootFile   string       `json:"boot_file"`
	Options    []DHCPOption `json:"options"`
}

func registerDHCPConfig(s *server.MCPServer, c writer) {
	// ── Config serveur DHCP ───────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_config",
			mcp.WithDescription("Configuration complète du serveur DHCP : activé, sticky_assign, plage IP, passerelle, DNS, boot_server (TFTP/PXE), boot_file, options custom RFC2132. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg DHCPConfig
			if err := c.Get(ctx, "/dhcp/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Options DHCP custom (lecture) ─────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_options",
			mcp.WithDescription("Options DHCP personnalisées du serveur Freebox : boot_server (TFTP/PXE, option 66), boot_file (option 67), tableau options[] RFC2132. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg DHCPConfig
			if err := c.Get(ctx, "/dhcp/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			result := struct {
				BootServer string       `json:"boot_server"`
				BootFile   string       `json:"boot_file"`
				Options    []DHCPOption `json:"options"`
			}{
				BootServer: cfg.BootServer,
				BootFile:   cfg.BootFile,
				Options:    cfg.Options,
			}
			return jsonResult(result)
		},
	)

	// ── Ajouter / modifier une option DHCP custom ─────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_option_set",
			mcp.WithDescription("Ajoute ou modifie une option DHCP personnalisée (ex: tftp_server_name, bootfile_name). Nécessite la permission 'settings'. Fait un read-modify-write sur /dhcp/config/."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Identifiant de l'option DHCP RFC2132 (ex: tftp_server_name, bootfile_name)")),
			mcp.WithString("val",
				mcp.Required(),
				mcp.Description("Valeur de l'option (ex: 192.168.100.254, /boot/default.ipxe)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetString("id", "")
			val := req.GetString("val", "")
			if id == "" {
				return mcp.NewToolResultError("id requis"), nil
			}

			// Lire config actuelle
			var cfg DHCPConfig
			if err := c.Get(ctx, "/dhcp/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Mettre à jour ou ajouter l'option
			found := false
			for i, opt := range cfg.Options {
				if opt.ID == id {
					cfg.Options[i].Val = val
					found = true
					break
				}
			}
			if !found {
				cfg.Options = append(cfg.Options, DHCPOption{ID: id, Val: val})
			}

			update := dhcpOptionsUpdate{
				BootServer: cfg.BootServer,
				BootFile:   cfg.BootFile,
				Options:    cfg.Options,
			}
			var updated DHCPConfig
			if err := c.Put(ctx, "/dhcp/config/", update, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Option DHCP '%s' définie à '%s'.", id, val)), nil
		},
	)

	// ── Supprimer une option DHCP custom ─────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_option_delete",
			mcp.WithDescription("Supprime une option DHCP personnalisée par son identifiant. Nécessite la permission 'settings'. Fait un read-modify-write sur /dhcp/config/."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("Identifiant de l'option à supprimer (ex: tftp_server_name, bootfile_name)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetString("id", "")
			if id == "" {
				return mcp.NewToolResultError("id requis"), nil
			}

			var cfg DHCPConfig
			if err := c.Get(ctx, "/dhcp/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			filtered := cfg.Options[:0]
			deleted := false
			for _, opt := range cfg.Options {
				if opt.ID == id {
					deleted = true
				} else {
					filtered = append(filtered, opt)
				}
			}
			if !deleted {
				return mcp.NewToolResultError(fmt.Sprintf("Option DHCP '%s' introuvable.", id)), nil
			}

			update := dhcpOptionsUpdate{
				BootServer: cfg.BootServer,
				BootFile:   cfg.BootFile,
				Options:    filtered,
			}
			var updated DHCPConfig
			if err := c.Put(ctx, "/dhcp/config/", update, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Option DHCP '%s' supprimée.", id)), nil
		},
	)
}
