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

// PortForwarding reflects one entry from GET /api/v4/fw/redir/
type PortForwarding struct {
	ID           int    `json:"id"`
	Enabled      bool   `json:"enabled"`
	Comment      string `json:"comment"`
	LanPort      int    `json:"lan_port"`
	WanPortStart int    `json:"wan_port_start"`
	WanPortEnd   int    `json:"wan_port_end"`
	LanIP        string `json:"lan_ip"`
	IPProto      string `json:"ip_proto"`
	SrcIP        string `json:"src_ip"`
}

func registerNAT(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_nat_rules",
			mcp.WithDescription("Liste les règles de redirection de ports NAT (port forwarding) configurées sur la Freebox."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var rules []PortForwarding
			if err := c.Get(ctx, "/fw/redir/", &rules); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(rules)
		},
	)
}
