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
	"path"
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

// encodeFSPath encodes an absolute Freebox path to standard base64 (with padding)
// as required by the /fs/ API endpoints.
// The Freebox API spec explicitly uses standard base64 (RFC 4648 §4) with "=" padding.
// Example from doc: /Disque dur → L0Rpc3F1ZSBkdXI=
func encodeFSPath(p string) string {
	// Ensure leading slash, trim trailing slash
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = strings.TrimRight(p, "/")
	return base64.StdEncoding.EncodeToString([]byte(p))
}

// sanitizeFSPath validates and cleans a filesystem path to prevent traversal attacks.
// Returns an error if the path contains ".." components or is otherwise unsafe.
func sanitizeFSPath(p string) (string, error) {
	if strings.Contains(p, "..") {
		return "", fmt.Errorf("chemin invalide : séquence '..' interdite")
	}
	// Clean the path (resolves redundant slashes, etc.)
	clean := path.Clean("/" + strings.TrimLeft(p, "/"))
	// Reject bare root
	if clean == "/" {
		return "", fmt.Errorf("chemin invalide : la racine '/' n'est pas un chemin de fichier valide")
	}
	return clean, nil
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
			raw := req.GetString("path", "/Freebox")
			p, err := sanitizeFSPath(raw)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			var entries []FSEntry
			if err := c.Get(ctx, "/fs/ls/"+encodeFSPath(p), &entries); err != nil {
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
			rawParent := req.GetString("parent", "/Freebox")
			name := req.GetString("name", "")
			parent, err := sanitizeFSPath(rawParent)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			if strings.ContainsAny(name, "/\\..") {
				return mcp.NewToolResultError("nom de répertoire invalide : '/', '\\', '..' interdits"), nil
			}
			body := map[string]string{
				"parent":  encodeFSPath(parent),
				"dirname": name,
			}
			var taskID string
			if err := c.Post(ctx, "/fs/mkdir/", body, &taskID); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Répertoire '%s' créé. Tâche : %s", name, taskID)), nil
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
			raw := req.GetString("path", "")
			p, err := sanitizeFSPath(raw)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			body := map[string]any{
				"files": []string{encodeFSPath(p)},
			}
			var taskID string
			if err := c.Post(ctx, "/fs/rm/", body, &taskID); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Suppression de '%s' en cours. Tâche : %s", p, taskID)), nil
		},
	)
}
