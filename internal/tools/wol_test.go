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

func newWOLServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerWOL(s, mock)
	return s
}

func TestWOL_OK(t *testing.T) {
	s := newWOLServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_wol", map[string]any{
		"mac": "aa:bb:cc:dd:ee:ff",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "aa:bb:cc:dd:ee:ff") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestWOL_WithIface(t *testing.T) {
	s := newWOLServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_wol", map[string]any{
		"mac":   "11:22:33:44:55:66",
		"iface": "pub",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "pub") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

// ── Sécurité : validation MAC et SecureOn ─────────────────────────────────────

func TestValidateMAC_Valid(t *testing.T) {
	macs := []string{"aa:bb:cc:dd:ee:ff", "AA:BB:CC:DD:EE:FF", "00:11:22:33:44:55"}
	for _, mac := range macs {
		if err := validateMAC(mac); err != nil {
			t.Errorf("validateMAC(%q) unexpected error: %v", mac, err)
		}
	}
}

func TestValidateMAC_Invalid(t *testing.T) {
	macs := []string{"not-a-mac", "aa:bb:cc:dd:ee", "'; DROP TABLE--", "gg:hh:ii:jj:kk:ll"}
	for _, mac := range macs {
		if err := validateMAC(mac); err == nil {
			t.Errorf("validateMAC(%q) should have returned error", mac)
		}
	}
}

func TestValidateSecureOn_Valid(t *testing.T) {
	if err := validateSecureOn(""); err != nil {
		t.Errorf("empty SecureOn should be valid: %v", err)
	}
	if err := validateSecureOn("aa:bb:cc:dd:ee:ff"); err != nil {
		t.Errorf("valid SecureOn should pass: %v", err)
	}
}

func TestValidateSecureOn_TooLong(t *testing.T) {
	if err := validateSecureOn(strings.Repeat("a", 1000)); err == nil {
		t.Error("oversized SecureOn should be rejected")
	}
}

func TestWOL_InvalidMAC(t *testing.T) {
	s := newWOLServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_wol", map[string]any{
		"mac": "not-a-mac-address",
	})
	if !result.IsError {
		t.Error("invalid MAC should return error")
	}
}
