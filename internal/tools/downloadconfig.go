/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// DownloadConfig reflects GET /api/v4/downloads/config/
type DownloadConfig struct {
	MaxDownloadingTasks  int     `json:"max_downloading_tasks"`
	DownloadDirB64       string  `json:"download_dir"`
	WatchDirB64          string  `json:"watch_dir,omitempty"`
	UseWatchDir          bool    `json:"use_watch_dir,omitempty"`
	SeedRatio            float64 `json:"seed_ratio"`
	StopSeedingOnBattery bool    `json:"stop_seeding_on_battery"`
	ScheduledDownload    bool    `json:"scheduled_download"`
	BWNormal             bool    `json:"bw_normal"`
	MaxDownloadSpeed     int64   `json:"max_downloading_speed"`
	MaxUploadSpeed       int64   `json:"max_uploading_speed"`
}

// validThrottlingModes lists the accepted values for the throttling.mode field.
var validThrottlingModes = map[string]bool{
	"normal":    true,
	"slow":      true,
	"hibernate": true,
	"schedule":  true,
}

func registerDownloadConfig(s *server.MCPServer, c writer) {
	// ── Lecture ───────────────────────────────────────────────────────────────
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

	// ── Écriture (partial update) ─────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_download_config_set",
			mcp.WithDescription("Modifie la configuration du gestionnaire de téléchargements Freebox. Tous les paramètres sont optionnels (mise à jour partielle). Utile en PRA pour corriger un download_dir invalide sans passer par l'interface web."),
			mcp.WithString("download_dir",
				mcp.Description("Nouveau dossier de destination des téléchargements (chemin absolu, ex: /Disque 1/Téléchargements). Encodé en base64 en interne.")),
			mcp.WithString("watch_dir",
				mcp.Description("Nouveau dossier surveillé pour auto-ajout .torrent/.nzb (chemin absolu).")),
			mcp.WithBoolean("use_watch_dir",
				mcp.Description("Active ou désactive la surveillance automatique du watch_dir.")),
			mcp.WithNumber("max_downloading_tasks",
				mcp.Description("Nombre maximum de téléchargements simultanés (entier ≥ 1).")),
			mcp.WithString("throttling_mode",
				mcp.Description("Mode de limitation de bande passante : normal | slow | hibernate | schedule.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			body := make(map[string]any)

			if dir := req.GetString("download_dir", ""); dir != "" {
				clean, err := sanitizeFSPath(dir)
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				body["download_dir"] = base64.StdEncoding.EncodeToString([]byte(clean))
			}

			if dir := req.GetString("watch_dir", ""); dir != "" {
				clean, err := sanitizeFSPath(dir)
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				body["watch_dir"] = base64.StdEncoding.EncodeToString([]byte(clean))
			}

			// use_watch_dir: check raw args to distinguish "not provided" from "false"
			if _, ok := req.GetArguments()["use_watch_dir"]; ok {
				body["use_watch_dir"] = req.GetBool("use_watch_dir", false)
			}

			if n := req.GetInt("max_downloading_tasks", 0); n > 0 {
				body["max_downloading_tasks"] = n
			}

			if mode := req.GetString("throttling_mode", ""); mode != "" {
				if !validThrottlingModes[mode] {
					return mcp.NewToolResultError(
						fmt.Sprintf("throttling_mode invalide : %q (valeurs : normal, slow, hibernate, schedule)", mode),
					), nil
				}
				body["throttling"] = map[string]any{"mode": mode}
			}

			if len(body) == 0 {
				return mcp.NewToolResultError("aucun paramètre fourni — spécifiez au moins un champ à modifier"), nil
			}

			var updated DownloadConfig
			if err := c.Put(ctx, "/downloads/config/", body, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)
}
