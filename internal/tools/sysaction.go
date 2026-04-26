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

func registerSysAction(s *server.MCPServer, c writer) {
	// ── Redémarrage Freebox ───────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_reboot",
			mcp.WithDescription("Redémarre la Freebox. ⚠️ Action irréversible — la connexion internet sera coupée pendant ~2 minutes. Demander confirmation explicite avant d'appeler cet outil."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var result any
			if err := c.Post(ctx, "/system/reboot/", nil, &result); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText("Redémarrage de la Freebox initié. La connexion sera interrompue ~2 minutes."), nil
		},
	)
}
