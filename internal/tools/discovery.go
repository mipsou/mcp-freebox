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

// discoverer is the interface required for the discovery tool (unauthenticated).
type discoverer interface {
	DiscoverAPI(ctx context.Context, dst any) error
}

// ApiVersionInfo reflects GET http://{host}/api_version (no auth, no envelope).
type ApiVersionInfo struct {
	UID            string `json:"uid"`
	DeviceName     string `json:"device_name"`
	APIVersion     string `json:"api_version"`
	APIBaseURL     string `json:"api_base_url"`
	DeviceType     string `json:"device_type"`
	HTTPSAvailable bool   `json:"https_available"`
	HTTPSPort      int    `json:"https_port"`
	BoxModel       string `json:"box_model"`
	BoxModelName   string `json:"box_model_name"`
}

func registerDiscovery(s *server.MCPServer, d discoverer) {
	s.AddTool(
		mcp.NewTool("freebox_discover",
			mcp.WithDescription("Découvre la Freebox sur le réseau local sans connaître son IP (mDNS via mafreebox.freebox.fr). Retourne modèle, version API, port HTTPS."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var info ApiVersionInfo
			if err := d.DiscoverAPI(ctx, &info); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(info)
		},
	)
}
