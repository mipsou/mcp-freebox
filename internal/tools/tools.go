/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"net/url"

	"github.com/mark3labs/mcp-go/server"
)

// getter is the read-only surface tools need from the API client.
type getter interface {
	Get(ctx context.Context, path string, dst any) error
}

// writer adds mutation methods (POST, PUT, DELETE).
// PostForm is required for endpoints that use application/x-www-form-urlencoded
// instead of application/json (e.g. /downloads/add/).
type writer interface {
	getter
	Post(ctx context.Context, path string, body, dst any) error
	PostForm(ctx context.Context, path string, values url.Values, dst any) error
	Put(ctx context.Context, path string, body, dst any) error
	Delete(ctx context.Context, path string) error
}

// RegisterAll wires every tool into the MCP server.
func RegisterAll(s *server.MCPServer, c writer, d discoverer) {
	registerDiscovery(s, d)
	registerConnection(s, c)
	registerConnectionLogs(s, c)
	registerConnectionIPv6(s, c)
	registerLAN(s, c)
	registerDHCP(s, c)
	registerNAT(s, c)
	registerWifi(s, c)
	registerWifiMacFilter(s, c)
	registerWifiPlanning(s, c)
	registerStorage(s, c)
	registerFilesystem(s, c)
	registerVM(s, c)
	registerVMInfo(s, c)
	registerSystem(s, c)
	registerSwitch(s, c)
	registerFirewall(s, c)
	registerVPN(s, c)
	registerNetshare(s, c)
	registerDownloads(s, c)
	registerCalls(s, c)
	registerParental(s, c)
	registerContacts(s, c)
	registerWOL(s, c)
	registerNetwork(s, c)
	registerSysAction(s, c)
	registerLANConfig(s, c)
	registerWifiBSS(s, c)
	registerTV(s, c)
	registerDHCPConfig(s, c)
	registerDHCPv6(s, c)
	registerTFTP(s, c)
	registerFTP(s, c)
	registerUPnP(s, c)
	registerLCD(s, c)
	registerSwitchConfig(s, c)
	registerFirmware(s, c)
	registerAirMedia(s, c)
	registerConnectionConfig(s, c)
	registerDownloadConfig(s, c)
}
