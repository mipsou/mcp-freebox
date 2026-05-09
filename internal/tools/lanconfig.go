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

// LanHostUpdate is the body for PUT /api/v15/lan/browser/pub/{id}.
// Tous les champs sont omitempty : le PUT accepte les patchs partiels sur cet
// endpoint (différent de /vm/{id} qui exige un body complet — cf #80).
type LanHostUpdate struct {
	PrimaryName string `json:"primary_name,omitempty"`
	HostType    string `json:"host_type,omitempty"`
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
			mcp.WithDescription("Renomme un équipement du réseau local (modifie le nom affiché dans l'interface Freebox et retourné par freebox_lan_hosts). Pour aussi changer le type, utiliser freebox_lan_host_update."),
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

	// ── Mettre à jour un équipement (nom + type) ─────────────────────────────
	// Patch partiel autorisé sur cet endpoint : on n'envoie que les champs
	// fournis. Pratique pour corriger les classifications automatiques erronées
	// (ex : Home Assistant détecté en "workstation" au lieu de "iot").
	s.AddTool(
		mcp.NewTool("freebox_lan_host_update",
			mcp.WithDescription("Met à jour les attributs modifiables d'un équipement LAN : nom (primary_name) et/ou type (host_type). Au moins un des deux doit être fourni."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("ID de l'équipement (champ 'id' dans freebox_lan_hosts)")),
			mcp.WithString("primary_name",
				mcp.Description("Nouveau nom (optionnel)")),
			mcp.WithString("host_type",
				mcp.Description("Nouveau type. Liste officielle dev.freebox.fr/sdk/os/lan/ + valeurs validées runtime sur firmware 4.9.18.1. Pas de 'iot' ni 'tv' : utiliser 'multimedia_device' pour streaming/audio/vidéo connectés, 'appliances' pour électroménager, 'networking_device' pour passerelles Zigbee/Hue/Trådfri."),
				mcp.Enum("workstation", "laptop", "smartphone", "tablet", "printer",
					"vg_console", "television", "nas", "ip_camera", "ip_phone",
					"freebox_player", "freebox_pop", "freebox_hd",
					"networking_device", "multimedia_device",
					"appliances", "other")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetString("id", "")
			if id == "" {
				return mcp.NewToolResultError("id : paramètre requis"), nil
			}
			body := LanHostUpdate{
				PrimaryName: req.GetString("primary_name", ""),
				HostType:    req.GetString("host_type", ""),
			}
			if body.PrimaryName == "" && body.HostType == "" {
				return mcp.NewToolResultError("au moins un de primary_name ou host_type doit être fourni"), nil
			}
			var updated LanHost
			if err := c.Put(ctx, fmt.Sprintf("/lan/browser/pub/%s", id), body, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)
}
