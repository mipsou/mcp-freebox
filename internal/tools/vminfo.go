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

// VMInfo reflects GET /api/v15/vm/info (sans trailing slash) — vue agrégée des
// ressources VM disponibles et utilisées sur la Freebox.
// SataPorts/UsbPorts = labels exposés par le firmware (ex: "sata-internal-p0").
type VMInfo struct {
	UsedCPUs    int      `json:"used_cpus"`
	TotalCPUs   int      `json:"total_cpus"`
	UsedMemory  int      `json:"used_memory"`
	TotalMemory int      `json:"total_memory"`
	UsbUsed     bool     `json:"usb_used"`
	SataUsed    bool     `json:"sata_used"`
	UsbPorts    []string `json:"usb_ports"`
	SataPorts   []string `json:"sata_ports"`
}

// VMDistro reflects une entrée de GET /api/v15/vm/distros/.
// Hash pointe vers un fichier de checksums (SHA256SUMS / SHA512SUMS) hébergé
// par Free pour valider l'image téléchargée.
type VMDistro struct {
	Name string `json:"name"`
	OS   string `json:"os"`
	URL  string `json:"url"`
	Hash string `json:"hash"`
}

func registerVMInfo(s *server.MCPServer, c getter) {
	s.AddTool(
		mcp.NewTool("freebox_vm_info",
			mcp.WithDescription("Vue agrégée des ressources de la fonctionnalité VM : CPU/RAM utilisés vs disponibles, ports SATA et USB exposés et leur statut d'occupation."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var info VMInfo
			if err := c.Get(ctx, "/vm/info", &info); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(info)
		},
	)

	s.AddTool(
		mcp.NewTool("freebox_vm_distros",
			mcp.WithDescription("Liste des distributions Linux installables sur la Freebox VM (Ubuntu, Debian, Fedora…) — name, os, URL de l'image cloudimg, URL des checksums."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var distros []VMDistro
			if err := c.Get(ctx, "/vm/distros/", &distros); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(distros)
		},
	)
}
