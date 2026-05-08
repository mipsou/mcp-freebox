/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

// DHCPOptions is a named slice of DHCPOption with a flexible JSON decoder.
// The Freebox API returns {} (empty object) instead of [] when no custom options
// are configured — confirmed empirically on 2026-05-08 (issue #60).
type DHCPOptions []DHCPOption

func (d *DHCPOptions) UnmarshalJSON(data []byte) error {
	// Locate the first non-whitespace byte to distinguish object from array.
	for _, b := range data {
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			continue
		}
		if b == '{' {
			// Empty-object sentinel returned by Freebox — treat as empty slice.
			*d = DHCPOptions{}
			return nil
		}
		break
	}
	// Standard JSON array (or null).
	type plain []DHCPOption
	var arr plain
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if arr == nil {
		*d = DHCPOptions{}
	} else {
		*d = DHCPOptions(arr)
	}
	return nil
}

// DHCPConfig reflects GET /api/v4/dhcp/config/
type DHCPConfig struct {
	Enabled              bool        `json:"enabled"`
	StickyAssign         bool        `json:"sticky_assign"` // toujours attribuer la même IP à un hôte donné
	BootServer           string      `json:"boot_server"`   // serveur TFTP / next-server BOOTP (siaddr)
	BootFile             string      `json:"boot_file"`     // fichier de boot (option 67)
	GatewayIP            string      `json:"gateway"`       // lecture seule
	NetmaskIP            string      `json:"netmask"`       // lecture seule
	IPRangeStart         string      `json:"ip_range_start"`
	IPRangeEnd           string      `json:"ip_range_end"`
	AlwaysBroadcast      bool        `json:"always_broadcast"`
	IgnoreOutOfRangeHint bool        `json:"ignore_out_of_range_hint"` // ignorer hint hors plage
	DNSServers           []string    `json:"dns"`
	Options              DHCPOptions `json:"options"` // options DHCP personnalisées RFC2132
}

// dhcpOptionsUpdate is the minimal payload accepted by PUT /api/v4/dhcp/config/
// for modifying only the custom options (partial update).
type dhcpOptionsUpdate struct {
	BootServer string      `json:"boot_server"`
	BootFile   string      `json:"boot_file"`
	Options    DHCPOptions `json:"options"`
}

// dhcpConfigUpdate is the write payload for PUT /api/v4/dhcp/config/
// — excludes gateway and netmask which are read-only.
type dhcpConfigUpdate struct {
	Enabled              bool        `json:"enabled"`
	StickyAssign         bool        `json:"sticky_assign"`
	BootServer           string      `json:"boot_server"`
	BootFile             string      `json:"boot_file"`
	IPRangeStart         string      `json:"ip_range_start"`
	IPRangeEnd           string      `json:"ip_range_end"`
	AlwaysBroadcast      bool        `json:"always_broadcast"`
	IgnoreOutOfRangeHint bool        `json:"ignore_out_of_range_hint"`
	DNSServers           []string    `json:"dns"`
	Options              DHCPOptions `json:"options"`
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
				BootServer string      `json:"boot_server"`
				BootFile   string      `json:"boot_file"`
				Options    DHCPOptions `json:"options"`
			}{
				BootServer: cfg.BootServer,
				BootFile:   cfg.BootFile,
				Options:    cfg.Options,
			}
			return jsonResult(result)
		},
	)

	// ── Modifier la config DHCP ───────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_config_set",
			mcp.WithDescription("Modifie la configuration du serveur DHCP (activé, plage IP, DNS, boot PXE, etc.). Read-modify-write sur /dhcp/config/. Nécessite la permission 'settings'."),
			mcp.WithBoolean("enabled",
				mcp.Description("Activer ou désactiver le serveur DHCP")),
			mcp.WithBoolean("sticky_assign",
				mcp.Description("Toujours attribuer la même IP à un hôte connu")),
			mcp.WithBoolean("always_broadcast",
				mcp.Description("Toujours répondre en broadcast")),
			mcp.WithBoolean("ignore_out_of_range_hint",
				mcp.Description("Ignorer les hints d'IP hors de la plage DHCP")),
			mcp.WithString("boot_server",
				mcp.Description("Serveur TFTP/PXE (next-server / siaddr). Passer une chaîne vide pour effacer.")),
			mcp.WithString("boot_file",
				mcp.Description("Fichier de boot PXE (option 67). Passer une chaîne vide pour effacer.")),
			mcp.WithString("ip_range_start",
				mcp.Description("Début de la plage DHCP (ex: 192.168.1.10)"),
				mcp.Pattern(IPv4Pattern)),
			mcp.WithString("ip_range_end",
				mcp.Description("Fin de la plage DHCP (ex: 192.168.1.200)"),
				mcp.Pattern(IPv4Pattern)),
			mcp.WithString("dns",
				mcp.Description("Serveurs DNS séparés par des virgules (ex: 192.168.1.1,192.168.1.2)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := req.GetArguments()

			// Vérification rapide avant GET : au moins un champ reconnu doit être présent.
			recognized := 0
			for _, name := range []string{
				"enabled", "sticky_assign", "always_broadcast", "ignore_out_of_range_hint",
				"boot_server", "boot_file", "ip_range_start", "ip_range_end", "dns",
			} {
				if _, ok := args[name]; ok {
					recognized++
				}
			}
			if recognized == 0 {
				return mcp.NewToolResultError("aucun champ à modifier"), nil
			}

			// Lire la config actuelle (read-modify-write).
			var cfg DHCPConfig
			if err := c.Get(ctx, "/dhcp/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Appliquer les champs fournis.
			if v, ok := args["enabled"].(bool); ok {
				cfg.Enabled = v
			}
			if v, ok := args["sticky_assign"].(bool); ok {
				cfg.StickyAssign = v
			}
			if v, ok := args["always_broadcast"].(bool); ok {
				cfg.AlwaysBroadcast = v
			}
			if v, ok := args["ignore_out_of_range_hint"].(bool); ok {
				cfg.IgnoreOutOfRangeHint = v
			}
			if v, ok := args["boot_server"].(string); ok {
				cfg.BootServer = v
			}
			if v, ok := args["boot_file"].(string); ok {
				cfg.BootFile = v
			}
			if v, ok := args["ip_range_start"].(string); ok && v != "" {
				cfg.IPRangeStart = v
			}
			if v, ok := args["ip_range_end"].(string); ok && v != "" {
				cfg.IPRangeEnd = v
			}
			if v, ok := args["dns"].(string); ok && v != "" {
				parts := strings.Split(v, ",")
				dns := make([]string, 0, len(parts))
				for _, p := range parts {
					if s := strings.TrimSpace(p); s != "" {
						dns = append(dns, s)
					}
				}
				cfg.DNSServers = dns
			}

			update := dhcpConfigUpdate{
				Enabled:              cfg.Enabled,
				StickyAssign:         cfg.StickyAssign,
				BootServer:           cfg.BootServer,
				BootFile:             cfg.BootFile,
				IPRangeStart:         cfg.IPRangeStart,
				IPRangeEnd:           cfg.IPRangeEnd,
				AlwaysBroadcast:      cfg.AlwaysBroadcast,
				IgnoreOutOfRangeHint: cfg.IgnoreOutOfRangeHint,
				DNSServers:           cfg.DNSServers,
				Options:              cfg.Options,
			}
			var updated DHCPConfig
			if err := c.Put(ctx, "/dhcp/config/", update, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
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
