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

// VPNServer reflects one entry from GET /api/v4/vpn/
// vpn_type : pptp | openvpn_routed | openvpn_bridged | ipsec | wireguard
type VPNServer struct {
	VPNType       string `json:"vpn_type"`
	Enabled       bool   `json:"enabled"`
	ActiveClients int    `json:"active_clients,omitempty"`
}

// VPNConnection reflects one entry from GET /api/v4/vpn/connection/
type VPNConnection struct {
	ID           string   `json:"id"`
	VPN          string   `json:"vpn"`           // vpn_type du serveur
	Login        string   `json:"login"`         // utilisateur connecté
	SrcIP        string   `json:"src_ip"`        // IP publique du client
	LocalIP      string   `json:"local_ip"`      // IP tunnel assignée
	PushedRoutes []string `json:"pushed_routes"` // routes poussées au client
	ConnectedAt  int64    `json:"connected_at"`  // timestamp Unix
}

// VPNClientConfig reflects one entry from GET /api/v4/vpn_client/config/
// type : wireguard | openvpn
type VPNClientConfig struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Type        string `json:"type"`
}

func registerVPN(s *server.MCPServer, c writer) {
	// ── Serveur VPN — état des protocoles ────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vpn_server_status",
			mcp.WithDescription("État des serveurs VPN de la Freebox : PPTP, OpenVPN Routé, OpenVPN Bridgé, IPsec IKEv2, WireGuard (activé/désactivé, clients actifs). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var servers []VPNServer
			if err := c.Get(ctx, "/vpn/", &servers); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(servers)
		},
	)

	// ── Serveur VPN — connexions actives ─────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vpn_connections",
			mcp.WithDescription("Liste les clients VPN actuellement connectés à la Freebox : protocole, login, IP source, IP tunnel, routes poussées, timestamp de connexion. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var conns []VPNConnection
			if err := c.Get(ctx, "/vpn/connection/", &conns); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(conns)
		},
	)

	// ── Client VPN — configurations sortantes ────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vpn_client_configs",
			mcp.WithDescription("Liste les configurations VPN client de la Freebox (connexions sortantes vers un serveur VPN externe) : description, type (wireguard/openvpn), état actif. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var configs []VPNClientConfig
			if err := c.Get(ctx, "/vpn_client/config/", &configs); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(configs)
		},
	)
}
