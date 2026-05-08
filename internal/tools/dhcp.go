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

// DhcpStaticLease reflects one entry from GET /api/v4/dhcp/static_lease/
type DhcpStaticLease struct {
	ID       string `json:"id"`
	Mac      string `json:"mac"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Comment  string `json:"comment"`
}

// DhcpDynamicLease reflects one entry from GET /api/v4/dhcp/dynamic_lease/
type DhcpDynamicLease struct {
	Mac            string `json:"mac"`
	Hostname       string `json:"hostname"`
	IP             string `json:"ip"`
	AssignTime     int64  `json:"assign_time"`
	RefreshTime    int64  `json:"refresh_time"`
	LeaseRemaining int    `json:"lease_remaining"` // secondes restantes avant renouvellement
	IsStatic       bool   `json:"is_static"`
}

func registerDHCP(s *server.MCPServer, c writer) {
	// ── Liste statique ───────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_static",
			mcp.WithDescription("Liste les réservations DHCP statiques configurées sur la Freebox (MAC → IP fixe)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var leases []DhcpStaticLease
			if err := c.Get(ctx, "/dhcp/static_lease/", &leases); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(leases)
		},
	)

	// ── Liste dynamique ──────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_leases",
			mcp.WithDescription("Liste les baux DHCP dynamiques actifs (clients connectés avec IP assignée)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var leases []DhcpDynamicLease
			if err := c.Get(ctx, "/dhcp/dynamic_lease/", &leases); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(leases)
		},
	)

	// ── Créer réservation statique ───────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_static_create",
			mcp.WithDescription("Crée une réservation DHCP statique (MAC → IP fixe). L'équipement obtiendra toujours la même IP."),
			mcp.WithString("mac",
				mcp.Required(),
				mcp.Description("Adresse MAC (ex: aa:bb:cc:dd:ee:ff)"),
				mcp.Pattern(MACAddrPattern)),
			mcp.WithString("ip",
				mcp.Required(),
				mcp.Description("IP à réserver — ne pas utiliser .0, .1, .254, .255 (ex: 192.168.1.50)"),
				mcp.Pattern(IPv4Pattern)),
			mcp.WithString("hostname",
				mcp.Description("Nom d'hôte (optionnel)")),
			mcp.WithString("comment",
				mcp.Description("Commentaire (optionnel)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			mac := req.GetString("mac", "")
			if err := validateMAC(mac); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			ip := req.GetString("ip", "")
			if err := validateDHCPIP(ip); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			hostname := req.GetString("hostname", "")
			comment := req.GetString("comment", "")
			body := DhcpStaticLease{Mac: mac, IP: ip, Hostname: hostname, Comment: comment}
			var created DhcpStaticLease
			if err := c.Post(ctx, "/dhcp/static_lease/", body, &created); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(created)
		},
	)

	// ── Supprimer réservation statique ───────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_dhcp_static_delete",
			mcp.WithDescription("Supprime une réservation DHCP statique par son ID (voir freebox_dhcp_static)."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("ID de la réservation à supprimer")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetString("id", "")
			if err := c.Delete(ctx, fmt.Sprintf("/dhcp/static_lease/%s", id)); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Réservation DHCP %s supprimée.", id)), nil
		},
	)

}
