/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newLANServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerLAN(s, mock)
	return s
}

func TestLanHosts_OK(t *testing.T) {
	s := newLANServer(t, mockGetter{
		"/lan/browser/pub/": []LanHost{
			{
				ID:          "ether-aa:bb:cc:dd:ee:ff",
				PrimaryName: "MyPC",
				Reachable:   true,
				L3Connectivities: []L3Connectivity{
					{Addr: "192.168.1.100", AF: "ipv4", Active: true},
				},
			},
		},
	})
	result := callTool(t, s, "freebox_lan_hosts")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"primary_name": "MyPC"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestLanHosts_APIError(t *testing.T) {
	s := newLANServer(t, mockGetter{})
	result := callTool(t, s, "freebox_lan_hosts")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestLanInterfaces_OK(t *testing.T) {
	s := newLANServer(t, mockGetter{
		"/lan/browser/interfaces/": []LanInterface{
			{Name: "pub", HostCount: 12},
			{Name: "guest", HostCount: 3},
		},
	})
	result := callTool(t, s, "freebox_lan_interfaces")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"name": "pub"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"host_count": 12`) {
		t.Errorf("missing host_count: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestLanInterfaces_APIError(t *testing.T) {
	s := newLANServer(t, mockGetter{})
	result := callTool(t, s, "freebox_lan_interfaces")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// L2Idents tolère que l'API renvoie l2ident en object single ou array.
// Sur firmware 4.9.18.1, observé en object pour certains hosts.

func TestL2Idents_UnmarshalArray(t *testing.T) {
	var l L2Idents
	if err := json.Unmarshal([]byte(`[{"id":"aa:bb","type":"mac"}]`), &l); err != nil {
		t.Fatalf("unmarshal array: %v", err)
	}
	if len(l) != 1 || l[0].ID != "aa:bb" {
		t.Errorf("got %+v", l)
	}
}

func TestL2Idents_UnmarshalSingleObject(t *testing.T) {
	var l L2Idents
	if err := json.Unmarshal([]byte(`{"id":"aa:bb","type":"mac"}`), &l); err != nil {
		t.Fatalf("unmarshal single object: %v", err)
	}
	if len(l) != 1 || l[0].ID != "aa:bb" || l[0].Type != "mac" {
		t.Errorf("single object → got %+v, want one entry wrapped", l)
	}
}

func TestL2Idents_UnmarshalNull(t *testing.T) {
	var l L2Idents
	if err := json.Unmarshal([]byte(`null`), &l); err != nil {
		t.Fatalf("unmarshal null: %v", err)
	}
	if l != nil {
		t.Errorf("null should yield nil slice, got %+v", l)
	}
}

func TestLanHost_DecodesL2IdentObjectShape(t *testing.T) {
	// Réplique la shape réelle observée en runtime sur firmware 4.9.18.1.
	payload := `{"id":"ether-aa:bb","primary_name":"x","host_type":"workstation",
		"l2ident":{"id":"AA:BB:CC:DD:EE:FF","type":"mac_address"},
		"l3connectivities":[]}`
	var h LanHost
	if err := json.Unmarshal([]byte(payload), &h); err != nil {
		t.Fatalf("LanHost decode failed on real-world l2ident shape: %v", err)
	}
	if len(h.L2Ident) != 1 || h.L2Ident[0].ID != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("L2Ident wrap failed: %+v", h.L2Ident)
	}
}
