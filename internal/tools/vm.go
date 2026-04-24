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

// VM reflects one entry from GET /api/v4/vm/
type VM struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	Memory           int    `json:"memory"`
	Vcpus            int    `json:"vcpus"`
	DiskPath         string `json:"disk_path"`
	DiskType         string `json:"disk_type"`
	OS               string `json:"os"`
	EnableScreen     bool   `json:"enable_screen"`
	CloudinitEnabled bool   `json:"cloudinit_enabled"`
}

func registerVM(s *server.MCPServer, c writer) {
	// ── Liste ────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_list",
			mcp.WithDescription("Liste les machines virtuelles de la Freebox : nom, état (running/stopped), mémoire, vCPUs, OS."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var vms []VM
			if err := c.Get(ctx, "/vm/", &vms); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(vms)
		},
	)

	// ── Démarrer ─────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_start",
			mcp.WithDescription("Démarre une VM arrêtée (PRA : remontée de service)."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Post(ctx, fmt.Sprintf("/vm/%d/start", id), nil, nil); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(`{"started": true, "id": %d}`, id)), nil
		},
	)

	// ── Arrêt gracieux (ACPI) ─────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_stop",
			mcp.WithDescription("Envoie un signal d'arrêt ACPI à une VM (équivalent bouton Power). Préférer freebox_vm_kill si la VM ne répond plus."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Post(ctx, fmt.Sprintf("/vm/%d/stop", id), nil, nil); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(`{"stopping": true, "id": %d}`, id)), nil
		},
	)

	// ── Forcer l'arrêt ───────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_kill",
			mcp.WithDescription("Force l'arrêt d'une VM bloquée (équivalent coupure secteur). Risque de corruption des données."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Post(ctx, fmt.Sprintf("/vm/%d/kill", id), nil, nil); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(`{"killed": true, "id": %d}`, id)), nil
		},
	)

	// ── Créer ────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_create",
			mcp.WithDescription("Crée une nouvelle machine virtuelle sur la Freebox."),
			mcp.WithString("name",
				mcp.Required(),
				mcp.Description("Nom de la VM")),
			mcp.WithNumber("memory",
				mcp.Required(),
				mcp.Description("Mémoire allouée en Mo (ex: 2048)")),
			mcp.WithNumber("vcpus",
				mcp.Required(),
				mcp.Description("Nombre de vCPUs")),
			mcp.WithString("disk_path",
				mcp.Required(),
				mcp.Description("Chemin du disque sur le stockage Freebox (ex: Freebox/VMs/disk.qcow2)")),
			mcp.WithString("disk_type",
				mcp.Required(),
				mcp.Description("Type de disque : raw ou qcow2")),
			mcp.WithString("os",
				mcp.Description("OS invité : fedora, debian, ubuntu, unknown (défaut: unknown)")),
			mcp.WithBoolean("enable_screen",
				mcp.Description("Activer l'écran virtuel VNC (défaut: false)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name := req.GetString("name", "")
			memory := req.GetInt("memory", 0)
			vcpus := req.GetInt("vcpus", 0)
			diskPath := req.GetString("disk_path", "")
			diskType := req.GetString("disk_type", "")
			osName := req.GetString("os", "unknown")
			enableScreen := req.GetBool("enable_screen", false)
			body := VM{
				Name: name, Memory: memory, Vcpus: vcpus,
				DiskPath: diskPath, DiskType: diskType,
				OS: osName, EnableScreen: enableScreen,
			}
			var created VM
			if err := c.Post(ctx, "/vm/", body, &created); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(created)
		},
	)

	// ── Modifier ─────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_update",
			mcp.WithDescription("Modifie la configuration d'une VM arrêtée (nom, mémoire, vCPUs, écran VNC)."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)")),
			mcp.WithString("name",
				mcp.Description("Nouveau nom (optionnel)")),
			mcp.WithNumber("memory",
				mcp.Description("Nouvelle mémoire en Mo (optionnel)")),
			mcp.WithNumber("vcpus",
				mcp.Description("Nouveau nombre de vCPUs (optionnel)")),
			mcp.WithBoolean("enable_screen",
				mcp.Description("Activer/désactiver l'écran VNC (optionnel)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			patch := map[string]any{}
			args := req.GetArguments()
			if v, ok := args["name"].(string); ok && v != "" {
				patch["name"] = v
			}
			if v, ok := args["memory"]; ok && v != nil {
				patch["memory"] = int(toFloat(v))
			}
			if v, ok := args["vcpus"]; ok && v != nil {
				patch["vcpus"] = int(toFloat(v))
			}
			if v, ok := args["enable_screen"].(bool); ok {
				patch["enable_screen"] = v
			}
			var updated VM
			if err := c.Put(ctx, fmt.Sprintf("/vm/%d", id), patch, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)

	// ── Supprimer ────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_vm_delete",
			mcp.WithDescription("Supprime définitivement une VM et libère son disque virtuel."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM à supprimer (voir freebox_vm_list)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Delete(ctx, fmt.Sprintf("/vm/%d", id)); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("VM %d supprimée.", id)), nil
		},
	)
}
