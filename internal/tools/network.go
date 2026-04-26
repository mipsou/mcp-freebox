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

// StaticRoute reflects one entry from GET /api/v4/network/route/ipv4/
type StaticRoute struct {
	ID      string `json:"id"`
	IP      string `json:"ip"`
	Mask    string `json:"mask"`
	Gateway string `json:"gw"`
	Active  bool   `json:"active"`
	Comment string `json:"comment"`
}

// StaticRoute6 reflects one entry from GET /api/v4/network/route/ipv6/
type StaticRoute6 struct {
	ID      string `json:"id"`
	Dest    string `json:"dest"`
	Prefix  int    `json:"prefix"`
	Gateway string `json:"gw"`
	Active  bool   `json:"active"`
	Comment string `json:"comment"`
}

func registerNetwork(s *server.MCPServer, c writer) {
	// ── Routes statiques IPv4 ─────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_routes_ipv4",
			mcp.WithDescription("Liste les routes statiques IPv4 configurées sur la Freebox (destination, masque, gateway, état actif). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var routes []StaticRoute
			if err := c.Get(ctx, "/network/route/ipv4/", &routes); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(routes)
		},
	)

	// ── Routes statiques IPv6 ─────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_routes_ipv6",
			mcp.WithDescription("Liste les routes statiques IPv6 configurées sur la Freebox (destination, préfixe, gateway, état actif). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var routes []StaticRoute6
			if err := c.Get(ctx, "/network/route/ipv6/", &routes); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(routes)
		},
	)

	// ── Ajouter route IPv4 ────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_route_add",
			mcp.WithDescription("Ajoute une route statique IPv4 sur la Freebox."),
			mcp.WithString("ip",
				mcp.Required(),
				mcp.Description("Adresse réseau de destination (ex: 10.0.0.0)")),
			mcp.WithString("mask",
				mcp.Required(),
				mcp.Description("Masque de sous-réseau (ex: 255.255.255.0)")),
			mcp.WithString("gw",
				mcp.Required(),
				mcp.Description("Adresse gateway (ex: 192.168.100.11)")),
			mcp.WithString("comment",
				mcp.Description("Commentaire (ex: CoreOS-11 VMs)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			body := StaticRoute{
				IP:      req.GetString("ip", ""),
				Mask:    req.GetString("mask", ""),
				Gateway: req.GetString("gw", ""),
				Comment: req.GetString("comment", ""),
				Active:  true,
			}
			var created StaticRoute
			if err := c.Post(ctx, "/network/route/ipv4/", body, &created); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(created)
		},
	)

	// ── Supprimer route IPv4 ──────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_route_delete",
			mcp.WithDescription("Supprime une route statique IPv4 de la Freebox."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("ID de la route à supprimer (voir freebox_routes_ipv4)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetString("id", "")
			if err := c.Delete(ctx, fmt.Sprintf("/network/route/ipv4/%s", id)); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Route IPv4 %s supprimée.", id)), nil
		},
	)
}
