/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// allowedDownloadSchemes lists URL schemes accepted by freebox_download_add.
// file://, gopher://, and other non-standard schemes are rejected to prevent SSRF.
var allowedDownloadSchemes = map[string]bool{
	"http":   true,
	"https":  true,
	"magnet": true,
	"nzb":    true,
}

// validateDownloadURL checks that the URL scheme is in the allowed list and
// that it does not target loopback or link-local addresses (SSRF prevention).
func validateDownloadURL(raw string) error {
	// Magnet links have a special format — allow them directly
	if strings.HasPrefix(raw, "magnet:") {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("URL invalide : %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if !allowedDownloadSchemes[scheme] {
		return fmt.Errorf("schéma URL interdit '%s' : seuls http, https, magnet, nzb sont autorisés", u.Scheme)
	}
	host := strings.ToLower(u.Hostname())
	// Block SSRF targets: loopback, link-local (AWS metadata), internal RFC1918 not needed here
	// (the Freebox handles the actual download, so we protect against accidental leaks)
	ssrfBlocked := []string{"localhost", "127.", "0.0.0.0", "169.254.", "::1", "[::1]"}
	for _, blocked := range ssrfBlocked {
		if strings.HasPrefix(host, blocked) || host == blocked {
			return fmt.Errorf("URL interdite : cible '%s' non autorisée (loopback/link-local)", host)
		}
	}
	return nil
}

// Download reflects one entry from GET /api/v4/downloads/
// status : stopped | seeding | downloading | done | error | checking | repairing | extracting | retry
type Download struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	Type             string `json:"type"` // http | bt | nzb
	Size             int64  `json:"size"`
	RxBytes          int64  `json:"rx_bytes"`
	TxBytes          int64  `json:"tx_bytes"`
	DownloadDir      string `json:"download_dir"`
	CreatedTimestamp int64  `json:"created_ts"`
	StoppedTimestamp int64  `json:"stopped_ts"`
	Error            string `json:"error"`
}

// DownloadAdd is the body for POST /api/v4/downloads/add/
type DownloadAdd struct {
	DownloadURL string `json:"download_url"`
	DownloadDir string `json:"download_dir,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
}

func registerDownloads(s *server.MCPServer, c writer) {
	// ── Liste ────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_downloads",
			mcp.WithDescription("Liste les téléchargements en cours et terminés sur la Freebox : nom, statut, type (HTTP/BitTorrent/NZB), taille, progression, répertoire de destination."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var downloads []Download
			if err := c.Get(ctx, "/downloads/", &downloads); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(downloads)
		},
	)

	// ── Ajouter ──────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_download_add",
			mcp.WithDescription("Ajoute un téléchargement à la Freebox (URL HTTP, magnet BitTorrent, ou lien NZB)."),
			mcp.WithString("url",
				mcp.Required(),
				mcp.Description("URL du fichier, lien magnet ou URL NZB")),
			mcp.WithString("download_dir",
				mcp.Description("Répertoire de destination (chemin sur le NAS Freebox, optionnel)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			rawURL := req.GetString("url", "")
			if err := validateDownloadURL(rawURL); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := DownloadAdd{
				DownloadURL: rawURL,
				DownloadDir: req.GetString("download_dir", ""),
			}
			var created Download
			if err := c.Post(ctx, "/downloads/add/", body, &created); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(created)
		},
	)

	// ── Pause / Reprise ───────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_download_toggle",
			mcp.WithDescription("Stoppe ou reprend un téléchargement existant sur la Freebox."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("ID du téléchargement (voir freebox_downloads)")),
			mcp.WithString("status",
				mcp.Required(),
				mcp.Description("Nouvel état : stopped (pause) ou downloading (reprendre)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			status := req.GetString("status", "")
			body := map[string]any{"status": status}
			var updated Download
			if err := c.Put(ctx, fmt.Sprintf("/downloads/%d", id), body, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)

	// ── Supprimer ────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_download_delete",
			mcp.WithDescription("Supprime un téléchargement de la liste (le fichier téléchargé n'est pas effacé du disque)."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("ID du téléchargement à supprimer")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Delete(ctx, fmt.Sprintf("/downloads/%d", id)); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Téléchargement %d supprimé.", id)), nil
		},
	)
}
