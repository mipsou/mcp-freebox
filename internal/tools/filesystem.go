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

func registerFilesystem(s *server.MCPServer, c getter) {
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
}
