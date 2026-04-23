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
	Mac        string `json:"mac"`
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	AssignTime int64  `json:"assign_time"`
	ExpireTime int64  `json:"expire_time"`
	IsStatic   bool   `json:"is_static"`
}

func registerDHCP(s *server.MCPServer, c getter) {
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
}
