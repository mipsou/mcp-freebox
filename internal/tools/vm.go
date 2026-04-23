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

	s.AddTool(
		mcp.NewTool("freebox_vm_start",
			mcp.WithDescription("Démarre une VM arrêtée (PRA : remontée de service)."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := int(req.GetArguments()["id"].(float64))
			if err := c.Post(ctx, fmt.Sprintf("/vm/%d/start", id), nil, nil); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(`{"started": true, "id": %d}`, id)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("freebox_vm_kill",
			mcp.WithDescription("Force l'arrêt d'une VM bloquée (équivalent coupure secteur)."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("Identifiant de la VM (voir freebox_vm_list)"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := int(req.GetArguments()["id"].(float64))
			if err := c.Post(ctx, fmt.Sprintf("/vm/%d/kill", id), nil, nil); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf(`{"killed": true, "id": %d}`, id)), nil
		},
	)
}
