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

// SwitchStats reflects GET /api/v15/switch/port/{id}/stats.
// Schéma exhaustif des compteurs Ethernet exposés par firmware 4.9.18.1 :
// volumes Rx/Tx (bytes, packets), taux instantanés (bytes/s, packets/s),
// erreurs (FCS, fragments, collisions, oversize, undersize, jabber, late,
// excessive, deferred, single, multiple), pause frames, broadcast, multicast,
// unicast, filtered, discard.
type SwitchStats struct {
	RxBytes            int64 `json:"rx_good_bytes"`
	RxBytesRate        int64 `json:"rx_bytes_rate"`
	RxPackets          int64 `json:"rx_good_packets"`
	RxPacketsRate      int64 `json:"rx_packets_rate"`
	RxBadBytes         int64 `json:"rx_bad_bytes"`
	RxBroadcast        int64 `json:"rx_broadcast_packets"`
	RxMulticastPackets int64 `json:"rx_multicast_packets"`
	RxUnicastPackets   int64 `json:"rx_unicast_packets"`
	RxDiscardPackets   int64 `json:"rx_discard_packets"`
	RxErrors           int64 `json:"rx_err_packets"`
	RxFCSPackets       int64 `json:"rx_fcs_packets"`
	RxFilteredPackets  int64 `json:"rx_filtered_packets"`
	RxFragmentsPackets int64 `json:"rx_fragments_packets"`
	RxJabberPackets    int64 `json:"rx_jabber_packets"`
	RxOversizePackets  int64 `json:"rx_oversize_packets"`
	RxUndersizePackets int64 `json:"rx_undersize_packets"`
	RxPause            int64 `json:"rx_pause"`
	TxBytes            int64 `json:"tx_bytes"`
	TxBytesRate        int64 `json:"tx_bytes_rate"`
	TxPackets          int64 `json:"tx_packets"`
	TxPacketsRate      int64 `json:"tx_packets_rate"`
	TxBroadcast        int64 `json:"tx_broadcast_packets"`
	TxMulticastPackets int64 `json:"tx_multicast_packets"`
	TxUnicastPackets   int64 `json:"tx_unicast_packets"`
	TxFilteredPackets  int64 `json:"tx_filtered_packets"`
	TxCollisions       int64 `json:"tx_collisions"`
	TxLate             int64 `json:"tx_late"`
	TxExcessive        int64 `json:"tx_excessive"`
	TxDeferred         int64 `json:"tx_deferred"`
	TxSingle           int64 `json:"tx_single"`
	TxMultiple         int64 `json:"tx_multiple"`
	TxFCS              int64 `json:"tx_fcs"`
	TxPause            int64 `json:"tx_pause"`
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
