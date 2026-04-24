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

// WifiApStatus reflects the runtime status of a WiFi access point.
type WifiApStatus struct {
	State               string `json:"state"`
	ChannelWidth        string `json:"channel_width"`
	PrimaryChannel      int    `json:"primary_channel"`
	SecondaryChannel    int    `json:"secondary_channel"`
	DfsCacRemainingTime int    `json:"dfs_cac_remaining_time"`
}

// WifiApConfig reflects the configuration of a WiFi access point.
type WifiApConfig struct {
	Band           string `json:"band"`
	ChannelWidth   string `json:"channel_width"`
	PrimaryChannel int    `json:"primary_channel"`
	Dfs            bool   `json:"dfs"`
	Enabled        bool   `json:"enabled"`
}

// WifiAp reflects one entry from GET /api/v4/wifi/ap/
type WifiAp struct {
	ID     int          `json:"id"`
	Name   string       `json:"name"`
	Status WifiApStatus `json:"status"`
	Config WifiApConfig `json:"config"`
}

// WifiGlobalConfig reflects GET /api/v4/wifi/config/
type WifiGlobalConfig struct {
	Enabled        bool   `json:"enabled"`
	MacFilterState string `json:"mac_filter_state"`
}

func registerWifi(s *server.MCPServer, c writer) {
	// ── Points d'accès ───────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_wifi_aps",
			mcp.WithDescription("Liste les points d'accès WiFi de la Freebox : bande (2g/5g/60g), canal, état, DFS."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var aps []WifiAp
			if err := c.Get(ctx, "/wifi/ap/", &aps); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(aps)
		},
	)

	// ── Config globale ───────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_wifi_config",
			mcp.WithDescription("Configuration WiFi globale de la Freebox (activé, état du filtre MAC)."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg WifiGlobalConfig
			if err := c.Get(ctx, "/wifi/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)

	// ── Toggle WiFi ──────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_wifi_toggle",
			mcp.WithDescription("Active ou désactive le WiFi global de la Freebox."),
			mcp.WithBoolean("enabled",
				mcp.Required(),
				mcp.Description("true = activer le WiFi, false = désactiver")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			enabled := req.GetBool("enabled", false)
			body := map[string]any{"enabled": enabled}
			var updated WifiGlobalConfig
			if err := c.Put(ctx, "/wifi/config/", body, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)
}
