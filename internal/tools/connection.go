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
	State         string `json:"state"`
	Type          string `json:"type"`
	Media         string `json:"media"`
	IPv4          string `json:"ipv4"`
	IPv6          string `json:"ipv6"`
	RateDown      int64  `json:"rate_down"`
	RateUp        int64  `json:"rate_up"`
	BandwidthDown int64  `json:"bandwidth_down"`
	BandwidthUp   int64  `json:"bandwidth_up"`
	BytesDown     int64  `json:"bytes_down"`
	BytesUp       int64  `json:"bytes_up"`
	IPv4PortRange [2]int `json:"ipv4_port_range"`
}

// XdslLine holds per-direction xDSL stats (down or up).
type XdslLine struct {
	ES         int `json:"es"`
	SES        int `json:"ses"`
	SNR        int `json:"snr"`
	Attn       int `json:"attn"`
	SNR10      int `json:"snr_10"`
	Attn10     int `json:"attn_10"`
	FEC        int `json:"fec"`
	CRC        int `json:"crc"`
	HEC        int `json:"hec"`
	RtMax      int `json:"rt_max"`
	RtMin      int `json:"rt_min"`
	RtAct      int `json:"rt_act"`
	Rxmt       int `json:"rxmt"`
	RxmtCorr   int `json:"rxmt_corr"`
	RxmtUncorr int `json:"rxmt_uncorr"`
	GinpFEC    int `json:"ginp_fec"`
	GinpUncorr int `json:"ginp_uncorr"`
}

// XdslStatus reflects GET /api/v4/connection/xdsl/
type XdslStatus struct {
	Status struct {
		Status     string `json:"status"`
		Modulation string `json:"modulation"`
		Uptime     int    `json:"uptime"`
	} `json:"status"`
	Down XdslLine `json:"down"`
	Up   XdslLine `json:"up"`
}

// FtthStatus reflects GET /api/v4/connection/ftth/
type FtthStatus struct {
	SfpPresent        bool   `json:"sfp_present"`
	SfpAlimOK         bool   `json:"sfp_alim_ok"`
	SfpHasPowerReport bool   `json:"sfp_has_power_report"`
	SfpHasSignal      bool   `json:"sfp_has_signal"`
	Link              bool   `json:"link"`
	SfpSerial         string `json:"sfp_serial"`
	SfpVendor         string `json:"sfp_vendor"`
	SfpPwrTx          int    `json:"sfp_pwr_tx"`
	SfpPwrRx          int    `json:"sfp_pwr_rx"`
}

// DynDNSEntry reflects one entry from GET /api/v4/dynDns/
type DynDNSEntry struct {
	Enabled  bool   `json:"enabled"`
	Hostname string `json:"hostname"`
	Type     string `json:"type"`
	User     string `json:"user"`
	State    string `json:"state"`
	StateMsg string `json:"state_msg"`
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
			mcp.WithDescription("Statistiques ligne xDSL (SNR, atténuation, erreurs FEC/CRC, uptime)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var st XdslStatus
			if err := c.Get(ctx, "/connection/xdsl/", &st); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(st)
		},
	)

	s.AddTool(
		mcp.NewTool("freebox_connection_ftth",
			mcp.WithDescription("Statistiques connexion FTTH (présence SFP, signal optique Tx/Rx en dBm×100)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var st FtthStatus
			if err := c.Get(ctx, "/connection/ftth/", &st); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(st)
		},
	)

	s.AddTool(
		mcp.NewTool("freebox_dyndns_list",
			mcp.WithDescription("Liste des entrées DynDNS configurées sur la Freebox."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var entries []DynDNSEntry
			if err := c.Get(ctx, "/dynDns/", &entries); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(entries)
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
