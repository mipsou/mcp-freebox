/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/server"
)

// getter is the read-only surface tools need from the API client.
type getter interface {
	Get(ctx context.Context, path string, dst any) error
}

// writer adds mutation methods (POST, PUT, DELETE).
type writer interface {
	getter
	Post(ctx context.Context, path string, body, dst any) error
	Put(ctx context.Context, path string, body, dst any) error
	Delete(ctx context.Context, path string) error
}

// RegisterAll wires every tool into the MCP server.
func RegisterAll(s *server.MCPServer, c writer) {
	registerConnection(s, c)
	// P1 — LAN, VM, Switch (to be added)
	// P2 — DHCP, TFTP, Firewall, WiFi, VPN, Storage, Netshare
	// P3 — System, Downloads, FS, Calls, Parental, FTP, AirMedia, LCD
}
