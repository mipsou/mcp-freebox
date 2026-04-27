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

// WifiBss reflects one entry from GET /api/v4/wifi/bss/
// A BSS (Basic Service Set) is a SSID broadcast on one AP.
type WifiBss struct {
	ID         string `json:"id"`
	BSSID      string `json:"bssid"`
	SSID       string `json:"ssid"`
	Band       string `json:"band"`
	Enabled    bool   `json:"enabled"`
	HideSSID   bool   `json:"hide_ssid"`
	Encryption string `json:"encryption"` // wpa2_psk | wpa3 | ...
	APID       int    `json:"ap_id"`
}

// WifiStation reflects one entry from GET /api/v4/wifi/stations/
// A station is a client currently associated to a BSS.
type WifiStation struct {
	MAC        string `json:"mac"`
	BSSID      string `json:"bssid"`
	Band       string `json:"band"`
	Signal     int    `json:"signal"`  // dBm
	RxRate     int    `json:"rx_rate"` // kbps
	TxRate     int    `json:"tx_rate"` // kbps
	Authorized bool   `json:"authorized"`
	Active     bool   `json:"active"`
}

func registerWifiBSS(s *server.MCPServer, c writer) {
	// ── SSIDs (BSS) ───────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_wifi_ssids",
			mcp.WithDescription("Liste les SSIDs WiFi de la Freebox : nom, bande, chiffrement, état caché/visible. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var bss []WifiBss
			if err := c.Get(ctx, "/wifi/bss/", &bss); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(bss)
		},
	)

	// ── Clients WiFi connectés ────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_wifi_stations",
			mcp.WithDescription("Liste les clients WiFi actuellement connectés à la Freebox : MAC, SSID, bande, signal (dBm), débit Rx/Tx. Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var stations []WifiStation
			if err := c.Get(ctx, "/wifi/stations/", &stations); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(stations)
		},
	)

	// ── Toggle SSID ───────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_wifi_ssid_toggle",
			mcp.WithDescription("Active ou désactive un SSID spécifique (BSS) sans toucher aux autres SSIDs."),
			mcp.WithString("id",
				mcp.Required(),
				mcp.Description("ID du BSS (champ 'id' dans freebox_wifi_ssids)")),
			mcp.WithBoolean("enabled",
				mcp.Required(),
				mcp.Description("true = activer le SSID, false = désactiver")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetString("id", "")
			body := map[string]any{"enabled": req.GetBool("enabled", false)}
			var updated WifiBss
			if err := c.Put(ctx, fmt.Sprintf("/wifi/bss/%s", id), body, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)
}
