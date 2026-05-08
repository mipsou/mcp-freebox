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

// SwitchPortConfig reflects GET /api/v4/switch/port/{id}
type SwitchPortConfig struct {
	ID         int    `json:"id"`
	DuplexMode string `json:"duplex"` // auto | full | half
	SpeedMode  string `json:"speed"`  // auto | 10 | 100 | 1000
}

// SwitchStats reflects GET /api/v4/switch/port/{id}/stats
type SwitchStats struct {
	RxBytesRate int64 `json:"rx_bytes_rate"`
	TxBytesRate int64 `json:"tx_bytes_rate"`
	RxBroadcast int64 `json:"rx_broadcast_packets"`
	TxBroadcast int64 `json:"tx_broadcast_packets"`
	RxErrors    int64 `json:"rx_err_packets"`
}

func registerSwitchConfig(s *server.MCPServer, c getter) {
	// ── Config d'un port ──────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_switch_port_config",
			mcp.WithDescription("Lit la configuration d'un port du switch LAN Freebox : activé, duplex (auto/full/half), vitesse (auto/10/100/1000 Mbps). Lecture seule."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Numéro du port (1-4 selon le modèle)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 1)
			var cfg SwitchPortConfig
			if err := c.Get(ctx, fmt.Sprintf("/switch/port/%d", id), &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Statistiques d'un port ────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_switch_port_stats",
			mcp.WithDescription("Statistiques temps réel d'un port du switch : débits Rx/Tx (bytes/s), paquets broadcast Rx/Tx, paquets erronés Rx."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Numéro du port (1-4)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 1)
			var stats SwitchStats
			if err := c.Get(ctx, fmt.Sprintf("/switch/port/%d/stats", id), &stats); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(stats)
		},
	)
}
