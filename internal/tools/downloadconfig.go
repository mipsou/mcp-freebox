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

// DownloadConfig reflects GET /api/v4/downloads/config/
type DownloadConfig struct {
	MaxDownloadingTasks  int     `json:"max_downloading_tasks"`
	DownloadDirB64       string  `json:"download_dir"`
	SeedRatio            float64 `json:"seed_ratio"`
	StopSeedingOnBattery bool    `json:"stop_seeding_on_battery"`
	ScheduledDownload    bool    `json:"scheduled_download"`
	BWNormal             bool    `json:"bw_normal"`
	MaxDownloadSpeed     int64   `json:"max_downloading_speed"`
	MaxUploadSpeed       int64   `json:"max_uploading_speed"`
}

func registerDownloadConfig(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_download_config",
			mcp.WithDescription("Lit la configuration du gestionnaire de téléchargements Freebox : dossier de destination, vitesses max Rx/Tx, ratio seeding, tâches simultanées, téléchargement planifié."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var cfg DownloadConfig
			if err := c.Get(ctx, "/downloads/config/", &cfg); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(cfg)
		},
	)
}
