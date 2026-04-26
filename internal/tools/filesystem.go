/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// FSEntry reflects one entry from GET /api/v6/fs/ls/{path}
// type : dir | file | link
// path : base64url-encoded full path on the Freebox storage
type FSEntry struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	Modification int64  `json:"modification"` // Unix timestamp
	Path         string `json:"path"`         // base64url encoded
	MimeType     string `json:"mimetype"`
}

// FSTask reflects the async task returned by mkdir/rm/mv/cp operations
type FSTask struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`  // mkdir | rm | mv | cp
	State    string `json:"state"` // queued | running | done | failed
	Error    string `json:"error"`
	To       string `json:"to"`
	From     string `json:"from"`
	Progress int    `json:"progress"`
}

// encodeFSPath encodes an absolute Freebox path to base64url (no padding)
// as required by the /fs/ls/ endpoint.
func encodeFSPath(path string) string {
	// Ensure leading slash, trim trailing slash
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = strings.TrimRight(path, "/")
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(path))
}

func registerFilesystem(s *server.MCPServer, c writer) {
	s.AddTool(
		mcp.NewTool("freebox_fs_list",
			mcp.WithDescription("Liste le contenu d'un répertoire sur le stockage de la Freebox (disque optionnel, clé USB…). Utile en PRA pour vérifier les images qcow2 disponibles dans /Freebox/VMs/."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description(`Chemin absolu sur le stockage Freebox, ex: /Freebox/VMs ou /Freebox`),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path := req.GetString("path", "/Freebox")
			encoded := encodeFSPath(path)
			var entries []FSEntry
			if err := c.Get(ctx, "/fs/ls/"+encoded, &entries); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(entries)
		},
	)

	// ── Créer un répertoire ───────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_fs_mkdir",
			mcp.WithDescription("Crée un répertoire sur le stockage de la Freebox. Retourne l'identifiant de la tâche asynchrone créée."),
			mcp.WithString("parent",
				mcp.Required(),
				mcp.Description("Chemin absolu du répertoire parent, ex: /Freebox/VMs"),
			),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Nom du nouveau répertoire à créer"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			parent := req.GetString("parent", "/Freebox")
			name := req.GetString("name", "")
			body := map[string]string{
				"parent":  encodeFSPath(parent),
				"dirname": name,
			}
			var task FSTask
			if err := c.Post(ctx, "/fs/mkdir/", body, &task); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(task)
		},
	)

	// ── Supprimer des fichiers/répertoires ────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_fs_delete",
			mcp.WithDescription("⚠️ Supprime un fichier ou un répertoire sur le stockage de la Freebox. Opération irréversible. Retourne l'identifiant de la tâche asynchrone."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Chemin absolu du fichier ou répertoire à supprimer, ex: /Freebox/Downloads/fichier.iso"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			path := req.GetString("path", "")
			body := map[string]any{
				"files": []string{encodeFSPath(path)},
			}
			var task FSTask
			if err := c.Post(ctx, "/fs/rm/", body, &task); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(task)
		},
	)
}
