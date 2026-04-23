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

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ConnectionStatus reflects GET /api/v4/connection/
type ConnectionStatus struct {
	State           string  `json:"state"`
	Type            string  `json:"type"`
	Media           string  `json:"media"`
	IPv4            string  `json:"ipv4"`
	IPv6            string  `json:"ipv6"`
	RateDown        int64   `json:"rate_down"`
	RateUp          int64   `json:"rate_up"`
	BandwidthDown   int64   `json:"bandwidth_down"`
	BandwidthUp     int64   `json:"bandwidth_up"`
	BytesDown       int64   `json:"bytes_down"`
	BytesUp         int64   `json:"bytes_up"`
	IPv4PortRange   [2]int  `json:"ipv4_port_range"`
}

func registerConnection(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_connection_status",
			mcp.WithDescription("État de la connexion WAN Freebox : état ligne, IP publique IPv4/IPv6, débits."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var st ConnectionStatus
			if err := c.Get(ctx, "/connection/", &st); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(st)
		},
	)

	s.AddTool(
		mcp.NewTool("freebox_connection_xdsl",
			mcp.WithDescription("Statistiques ligne xDSL (SNR, atténuation, erreurs)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var result json.RawMessage
			if err := c.Get(ctx, "/connection/xdsl/", &result); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("freebox_connection_ftth",
			mcp.WithDescription("Statistiques connexion FTTH (SFP, signal optique)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var result json.RawMessage
			if err := c.Get(ctx, "/connection/ftth/", &result); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("freebox_dyndns_list",
			mcp.WithDescription("Liste des entrées DynDNS configurées sur la Freebox."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var result json.RawMessage
			if err := c.Get(ctx, "/dynDns/", &result); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		},
	)
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return mcp.NewToolResultText(string(b)), nil
}
