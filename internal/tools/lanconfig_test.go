/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newLANConfigServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerLANConfig(s, mock)
	return s
}

func TestLanConfig_OK(t *testing.T) {
	s := newLANConfigServer(t, mockGetter{
		"/lan/config/": LanConfig{
			IP: "192.168.1.254", Name: "Freebox",
			NameDNS: "freebox", NameMDNS: "freebox", Type: "router",
		},
	})
	result := callTool(t, s, "freebox_lan_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "router") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestLanConfig_APIError(t *testing.T) {
	s := newLANConfigServer(t, mockGetter{})
	result := callTool(t, s, "freebox_lan_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestLanHostRename_OK(t *testing.T) {
	s := newLANConfigServer(t, mockGetter{
		"/lan/browser/pub/ether-aa:bb:cc:dd:ee:ff": LanHost{
			ID: "ether-aa:bb:cc:dd:ee:ff", PrimaryName: "CoreOS-11",
		},
	})
	result := callToolWithArgs(t, s, "freebox_lan_host_rename", map[string]any{
		"id":   "ether-aa:bb:cc:dd:ee:ff",
		"name": "CoreOS-11",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestLanHostRename_NoError(t *testing.T) {
	// mockGetter.Put always returns nil — rename succeeds with empty struct result.
	s := newLANConfigServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_lan_host_rename", map[string]any{
		"id":   "ether-xx:xx:xx:xx:xx:xx",
		"name": "Test",
	})
	if result.IsError {
		t.Errorf("unexpected error: %v", result.Content)
	}
}

// ── lan_host_update : nom + type via une seule API ─────────────────────────

func newLANConfigWriterServer(t *testing.T, mock mockWriter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerLANConfig(s, mock)
	return s
}

func TestLanHostUpdate_BothFields_BodyShape(t *testing.T) {
	puts := map[string]any{}
	s := newLANConfigWriterServer(t, mockWriter{
		mockGetter: mockGetter{},
		putBodies:  puts,
	})
	callToolWithArgs(t, s, "freebox_lan_host_update", map[string]any{
		"id":           "ether-aa:bb",
		"primary_name": "HomeAssistant",
		"host_type":    "iot",
	})
	body := puts["/lan/browser/pub/ether-aa:bb"].(LanHostUpdate)
	if body.PrimaryName != "HomeAssistant" {
		t.Errorf("primary_name = %q, want HomeAssistant", body.PrimaryName)
	}
	if body.HostType != "iot" {
		t.Errorf("host_type = %q, want iot", body.HostType)
	}
}

func TestLanHostUpdate_OnlyType_OmitsName(t *testing.T) {
	puts := map[string]any{}
	s := newLANConfigWriterServer(t, mockWriter{
		mockGetter: mockGetter{},
		putBodies:  puts,
	})
	callToolWithArgs(t, s, "freebox_lan_host_update", map[string]any{
		"id":        "ether-cc:dd",
		"host_type": "nas",
	})
	body := puts["/lan/browser/pub/ether-cc:dd"].(LanHostUpdate)
	if body.PrimaryName != "" {
		t.Errorf("primary_name should be empty when not provided, got %q", body.PrimaryName)
	}
	if body.HostType != "nas" {
		t.Errorf("host_type = %q, want nas", body.HostType)
	}
}

func TestLanHostUpdate_RejectsEmpty(t *testing.T) {
	s := newLANConfigWriterServer(t, mockWriter{mockGetter: mockGetter{}})
	r := callToolWithArgs(t, s, "freebox_lan_host_update", map[string]any{
		"id": "ether-xx",
	})
	if !r.IsError {
		t.Error("must reject when neither primary_name nor host_type is provided")
	}
}

func TestLanHostUpdate_RejectsMissingID(t *testing.T) {
	s := newLANConfigWriterServer(t, mockWriter{mockGetter: mockGetter{}})
	r := callToolWithArgs(t, s, "freebox_lan_host_update", map[string]any{
		"host_type": "iot",
	})
	if !r.IsError {
		t.Error("must reject when id is missing")
	}
}
