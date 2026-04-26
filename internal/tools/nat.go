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

// PortForwarding reflects one entry from GET /api/v4/fw/redir/
type PortForwarding struct {
	ID           int    `json:"id"`
	Enabled      bool   `json:"enabled"`
	Comment      string `json:"comment"`
	LanPort      int    `json:"lan_port"`
	WanPortStart int    `json:"wan_port_start"`
	WanPortEnd   int    `json:"wan_port_end"`
	LanIP        string `json:"lan_ip"`
	IPProto      string `json:"ip_proto"`
	SrcIP        string `json:"src_ip"`
}

func registerNAT(s *server.MCPServer, c writer) {
	// ── Liste ────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_nat_rules",
			mcp.WithDescription("Liste les règles de redirection de ports NAT (port forwarding) configurées sur la Freebox."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var rules []PortForwarding
			if err := c.Get(ctx, "/fw/redir/", &rules); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(rules)
		},
	)

	// ── Créer ────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_nat_create",
			mcp.WithDescription("Crée une règle de redirection de port (port forwarding) sur la Freebox."),
			mcp.WithString("lan_ip",
				mcp.Required(),
				mcp.Description("IP locale de destination — adresse RFC1918 uniquement (ex: 192.168.100.11)"),
				mcp.Pattern(RFC1918Pattern)),
			mcp.WithNumber("lan_port",
				mcp.Required(),
				mcp.Description("Port local de destination (1–65535)"),
				mcp.Min(1), mcp.Max(65535)),
			mcp.WithNumber("wan_port_start",
				mcp.Required(),
				mcp.Description("Port WAN de début (1–65535)"),
				mcp.Min(1), mcp.Max(65535)),
			mcp.WithNumber("wan_port_end",
				mcp.Description("Port WAN de fin (1–65535, égal à wan_port_start si non précisé)"),
				mcp.Min(1), mcp.Max(65535)),
			mcp.WithString("ip_proto",
				mcp.Required(),
				mcp.Description("Protocole : tcp ou udp"),
				mcp.Enum("tcp", "udp")),
			mcp.WithString("comment",
				mcp.Description("Commentaire (ex: SSH CoreOS)")),
			mcp.WithBoolean("enabled",
				mcp.Description("Activer la règle immédiatement (défaut: true)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			lanIP := req.GetString("lan_ip", "")
			if err := validateRFC1918(lanIP); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			lanPort := req.GetInt("lan_port", 0)
			if err := validatePort(lanPort, "lan_port"); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			wanStart := req.GetInt("wan_port_start", 0)
			if err := validatePort(wanStart, "wan_port_start"); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			wanEnd := wanStart
			if args := req.GetArguments(); args != nil {
				if v, ok := args["wan_port_end"]; ok {
					wanEnd = int(toFloat(v))
					if err := validatePort(wanEnd, "wan_port_end"); err != nil {
						return mcp.NewToolResultError(err.Error()), nil
					}
				}
			}
			proto := req.GetString("ip_proto", "")
			comment := req.GetString("comment", "")
			enabled := true
			if args := req.GetArguments(); args != nil {
				if v, ok := args["enabled"].(bool); ok {
					enabled = v
				}
			}
			body := PortForwarding{
				LanIP:        lanIP,
				LanPort:      lanPort,
				WanPortStart: wanStart,
				WanPortEnd:   wanEnd,
				IPProto:      proto,
				Comment:      comment,
				Enabled:      enabled,
			}
			var created PortForwarding
			if err := c.Post(ctx, "/fw/redir/", body, &created); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(created)
		},
	)

	// ── Activer / Désactiver ─────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_nat_toggle",
			mcp.WithDescription("Active ou désactive une règle NAT existante sans la supprimer."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("ID de la règle (voir freebox_nat_rules)")),
			mcp.WithBoolean("enabled",
				mcp.Required(),
				mcp.Description("true = activer, false = désactiver")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			enabled := req.GetBool("enabled", false)
			body := map[string]any{"enabled": enabled}
			var updated PortForwarding
			if err := c.Put(ctx, fmt.Sprintf("/fw/redir/%d", id), body, &updated); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(updated)
		},
	)

	// ── Supprimer ────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_nat_delete",
			mcp.WithDescription("Supprime définitivement une règle de redirection de port NAT."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("ID de la règle à supprimer (voir freebox_nat_rules)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			if err := c.Delete(ctx, fmt.Sprintf("/fw/redir/%d", id)); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Règle NAT %d supprimée.", id)), nil
		},
	)
}

// toFloat converts interface{} to float64 safely.
func toFloat(v any) float64 {
	f, _ := v.(float64)
	return f
}
