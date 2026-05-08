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
// type  : pptp | openvpn | ipsec | wireguard
// name  : pptp | openvpn_routed | openvpn_bridge | ipsec | wireguard
// state : started | stopped
type VPNServer struct {
	Type                string `json:"type"`
	Name                string `json:"name"`
	State               string `json:"state"`
	ConnectionCount     int    `json:"connection_count"`
	AuthConnectionCount int    `json:"auth_connection_count"`
}

// VPNConnection reflects one entry from GET /api/v4/vpn/connection/
type VPNConnection struct {
	ID            string `json:"id"`
	VPN           string `json:"vpn"`           // nom du serveur (ex: wireguard)
	User          string `json:"user"`          // utilisateur connecté
	Authenticated bool   `json:"authenticated"` // connexion authentifiée
	AuthTime      int64  `json:"auth_time"`     // timestamp Unix d'authentification
	SrcIP         string `json:"src_ip"`        // IP publique du client
	SrcPort       int    `json:"src_port"`      // port source du client
	LocalIP       string `json:"local_ip"`      // IP tunnel assignée
	RxBytes       int    `json:"rx_bytes"`      // octets reçus
	TxBytes       int    `json:"tx_bytes"`      // octets envoyés
}

// VPNClientConfig reflects one entry from GET /api/v4/vpn_client/config/
// type : wireguard | openvpn
type VPNClientConfig struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Active      bool   `json:"active"`
	Type        string `json:"type"`
}

func registerVPN(s *server.MCPServer, c getter) {
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
			mcp.WithDescription("Liste les clients VPN actuellement connectés à la Freebox : protocole, utilisateur, IP source/port, IP tunnel, octets Rx/Tx, timestamp d'authentification. Lecture seule."),
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
