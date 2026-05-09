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
	"net/url"
	"path"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// FSInfo reflects GET /api/v15/fs/info/?path={base64}.
// Note : ce endpoint ATTEND le path en query-string (URL-encodé), pas en
// segment d'URL comme /fs/ls/.
type FSInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	Type         string `json:"type"` // dir | file | link
	Size         int64  `json:"size"`
	Index        int    `json:"index"`
	Link         bool   `json:"link"`
	Parent       string `json:"parent"`
	Modification int64  `json:"modification"`
	Hidden       bool   `json:"hidden"`
	MimeType     string `json:"mimetype"`
}

// FSEntry reflects one entry from GET /api/v15/fs/ls/{path}
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

// FSListResult enveloppe la réponse de GET /api/v15/fs/ls/{path}.
// Sur firmware 4.9.18.1, l'API retourne {entries:[...], parent:{...}} et non
// pas un tableau brut — un unmarshal direct vers []FSEntry échoue avec
// "cannot unmarshal object into Go value of type []tools.FSEntry" (#93).
type FSListResult struct {
	Entries []FSEntry `json:"entries"`
	Parent  *FSEntry  `json:"parent,omitempty"`
}

// FSTask reflects the async task returned by rm/mv/cp operations.
// Note: mkdir returns a bare string task ID, not an FSTask object.
type FSTask struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`  // rm | mv | cp
	State    string `json:"state"` // queued | running | done | failed
	Error    string `json:"error"`
	From     string `json:"from"`
	To       string `json:"to"`
	Progress int    `json:"progress"`
}

// validFSDstModes lists accepted values for the dst_mode parameter of mv/cp.
var validFSDstModes = map[string]bool{
	"overwrite": true,
	"both":      true,
	"recent":    true,
	"skip":      true,
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
	// ── Info sur un fichier/répertoire ───────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_fs_info",
			mcp.WithDescription("Renvoie les métadonnées d'un fichier ou répertoire sur le stockage Freebox : type (file/dir/link), taille, mimetype, dernière modification, parent."),
			mcp.WithString("path",
				mcp.Required(),
				mcp.Description("Chemin absolu sur le stockage Freebox, ex: /Disque dur/Téléchargements/photo.jpg"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			raw := req.GetString("path", "")
			p, err := sanitizeFSPath(raw)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			// /fs/info/ exige path en query-string URL-encode (pas en segment).
			query := "?path=" + url.QueryEscape(encodeFSPath(p))
			var info FSInfo
			if err := c.Get(ctx, "/fs/info/"+query, &info); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(info)
		},
	)

	// ── Lister ───────────────────────────────────────────────────────────────
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
			var listing FSListResult
			if err := c.Get(ctx, "/fs/ls/"+encodeFSPath(p), &listing); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(listing.Entries)
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
			// /fs/mkdir/ returns a bare string task ID (not an FSTask object)
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
			mcp.WithDescription("⚠️ Supprime un fichier ou un répertoire sur le stockage de la Freebox. Opération irréversible. Retourne la tâche asynchrone créée."),
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
			// /fs/rm/ returns a FSTask object (unlike /fs/mkdir/ which returns a bare string)
			var task FSTask
			if err := c.Post(ctx, "/fs/rm/", body, &task); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(task)
		},
	)

	// ── Déplacer des fichiers/répertoires ─────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_fs_move",
			mcp.WithDescription("Déplace un ou plusieurs fichiers/répertoires vers un répertoire de destination sur le stockage Freebox. Retourne la tâche asynchrone créée."),
			mcp.WithArray("src_paths",
				mcp.Required(),
				mcp.Description("Liste des chemins sources absolus à déplacer, ex: [\"/Disque 1/Téléchargements/alma.qcow2\"]"),
				mcp.WithStringItems(),
			),
			mcp.WithString("dst_path",
				mcp.Required(),
				mcp.Description("Répertoire de destination absolu, ex: /Freebox/VMs"),
			),
			mcp.WithString("dst_mode",
				mcp.Description("Comportement si la destination existe : overwrite | both | recent | skip (défaut: skip)"),
				mcp.Enum("overwrite", "both", "recent", "skip"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			srcs := req.GetStringSlice("src_paths", nil)
			if len(srcs) == 0 {
				return mcp.NewToolResultError("src_paths ne peut pas être vide"), nil
			}
			rawDst := req.GetString("dst_path", "")
			dst, err := sanitizeFSPath(rawDst)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			mode := req.GetString("dst_mode", "skip")
			if !validFSDstModes[mode] {
				return mcp.NewToolResultError(
					fmt.Sprintf("dst_mode invalide : %q (valeurs : overwrite, both, recent, skip)", mode),
				), nil
			}
			encodedSrcs := make([]string, 0, len(srcs))
			for _, s := range srcs {
				clean, err := sanitizeFSPath(s)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("src_paths[%q] : %v", s, err)), nil
				}
				encodedSrcs = append(encodedSrcs, encodeFSPath(clean))
			}
			body := map[string]any{
				"files": encodedSrcs,
				"dst":   encodeFSPath(dst),
				"mode":  mode,
			}
			var task FSTask
			if err := c.Post(ctx, "/fs/mv/", body, &task); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(task)
		},
	)

	// ── Copier des fichiers/répertoires ───────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_fs_copy",
			mcp.WithDescription("Copie un ou plusieurs fichiers/répertoires vers un répertoire de destination sur le stockage Freebox. Retourne la tâche asynchrone créée."),
			mcp.WithArray("src_paths",
				mcp.Required(),
				mcp.Description("Liste des chemins sources absolus à copier, ex: [\"/Disque 1/Téléchargements/alma.qcow2\"]"),
				mcp.WithStringItems(),
			),
			mcp.WithString("dst_path",
				mcp.Required(),
				mcp.Description("Répertoire de destination absolu, ex: /Freebox/VMs"),
			),
			mcp.WithString("dst_mode",
				mcp.Description("Comportement si la destination existe : overwrite | both | recent | skip (défaut: skip)"),
				mcp.Enum("overwrite", "both", "recent", "skip"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			srcs := req.GetStringSlice("src_paths", nil)
			if len(srcs) == 0 {
				return mcp.NewToolResultError("src_paths ne peut pas être vide"), nil
			}
			rawDst := req.GetString("dst_path", "")
			dst, err := sanitizeFSPath(rawDst)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			mode := req.GetString("dst_mode", "skip")
			if !validFSDstModes[mode] {
				return mcp.NewToolResultError(
					fmt.Sprintf("dst_mode invalide : %q (valeurs : overwrite, both, recent, skip)", mode),
				), nil
			}
			encodedSrcs := make([]string, 0, len(srcs))
			for _, s := range srcs {
				clean, err := sanitizeFSPath(s)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("src_paths[%q] : %v", s, err)), nil
				}
				encodedSrcs = append(encodedSrcs, encodeFSPath(clean))
			}
			body := map[string]any{
				"files": encodedSrcs,
				"dst":   encodeFSPath(dst),
				"mode":  mode,
			}
			var task FSTask
			if err := c.Post(ctx, "/fs/cp/", body, &task); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(task)
		},
	)

	// ── Renommer un fichier (in-place) ────────────────────────────────────────
	// /fs/rename/ diffère de /fs/mv/ : il change le nom sans changer le
	// répertoire parent et est synchrone (retourne le nouveau path en base64,
	// pas une FSTask). Combler le manque #94.
	s.AddTool(
		mcp.NewTool("freebox_fs_rename",
			mcp.WithDescription("Renomme un fichier ou répertoire sur le stockage Freebox (in-place, sans changer de parent). Pour déplacer vers un autre dossier, utiliser freebox_fs_move."),
			mcp.WithString("src_path",
				mcp.Required(),
				mcp.Description("Chemin absolu du fichier/dossier à renommer, ex: /Disque 1/VMs/old.qcow2")),
			mcp.WithString("new_name",
				mcp.Required(),
				mcp.Description("Nouveau nom (basename uniquement, sans chemin), ex: new.qcow2")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			rawSrc := req.GetString("src_path", "")
			src, err := sanitizeFSPath(rawSrc)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("src_path : %v", err)), nil
			}
			newName := req.GetString("new_name", "")
			if newName == "" {
				return mcp.NewToolResultError("new_name : paramètre requis"), nil
			}
			if strings.ContainsAny(newName, "/\\") {
				return mcp.NewToolResultError("new_name : ne doit pas contenir de séparateur de chemin (utiliser freebox_fs_move pour déplacer)"), nil
			}
			if strings.Contains(newName, "..") {
				return mcp.NewToolResultError("new_name : séquence '..' interdite"), nil
			}
			body := map[string]any{
				"src": encodeFSPath(src),
				"dst": newName,
			}
			var newPath string
			if err := c.Post(ctx, "/fs/rename/", body, &newPath); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			// L'API retourne le nouveau chemin complet en base64 — on le décode
			// pour faciliter l'usage côté caller.
			decoded, decodeErr := base64.StdEncoding.DecodeString(newPath)
			if decodeErr != nil {
				// fallback : retourner le base64 brut si le décodage échoue
				return mcp.NewToolResultText(newPath), nil
			}
			return mcp.NewToolResultText(string(decoded)), nil
		},
	)
}
