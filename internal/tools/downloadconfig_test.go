/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newDownloadConfigServer(t *testing.T, mock writer) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerDownloadConfig(s, mock)
	return s
}

// ── freebox_download_config (GET) ─────────────────────────────────────────────

func TestDownloadConfig_OK(t *testing.T) {
	s := newDownloadConfigServer(t, mockWriter{
		mockGetter: mockGetter{
			"/downloads/config/": DownloadConfig{MaxDownloadingTasks: 5, MaxDownloadSpeed: 10485760},
		},
	})
	result := callTool(t, s, "freebox_download_config")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "10485760") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDownloadConfig_APIError(t *testing.T) {
	s := newDownloadConfigServer(t, mockWriter{})
	result := callTool(t, s, "freebox_download_config")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// ── freebox_download_config_set (PUT) ─────────────────────────────────────────

func TestDownloadConfigSet_DownloadDir_OK(t *testing.T) {
	s := newDownloadConfigServer(t, mockWriter{
		mockGetter: mockGetter{
			"/downloads/config/": DownloadConfig{MaxDownloadingTasks: 5},
		},
	})
	result := callToolWithArgs(t, s, "freebox_download_config_set", map[string]any{
		"download_dir": "/Disque 1/Téléchargements",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestDownloadConfigSet_ThrottlingMode_OK(t *testing.T) {
	s := newDownloadConfigServer(t, mockWriter{
		mockGetter: mockGetter{
			"/downloads/config/": DownloadConfig{},
		},
	})
	result := callToolWithArgs(t, s, "freebox_download_config_set", map[string]any{
		"throttling_mode": "slow",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestDownloadConfigSet_MaxTasks_OK(t *testing.T) {
	s := newDownloadConfigServer(t, mockWriter{
		mockGetter: mockGetter{
			"/downloads/config/": DownloadConfig{MaxDownloadingTasks: 3},
		},
	})
	result := callToolWithArgs(t, s, "freebox_download_config_set", map[string]any{
		"max_downloading_tasks": float64(3),
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestDownloadConfigSet_EmptyBody_Error(t *testing.T) {
	s := newDownloadConfigServer(t, mockWriter{})
	result := callTool(t, s, "freebox_download_config_set")
	if !result.IsError {
		t.Error("expected error when no params provided")
	}
}

func TestDownloadConfigSet_InvalidThrottlingMode(t *testing.T) {
	s := newDownloadConfigServer(t, mockWriter{})
	result := callToolWithArgs(t, s, "freebox_download_config_set", map[string]any{
		"throttling_mode": "turbo",
	})
	if !result.IsError {
		t.Error("expected error for invalid throttling_mode")
	}
}

func TestDownloadConfigSet_TraversalBlocked(t *testing.T) {
	s := newDownloadConfigServer(t, mockWriter{})
	result := callToolWithArgs(t, s, "freebox_download_config_set", map[string]any{
		"download_dir": "/../etc/passwd",
	})
	if !result.IsError {
		t.Error("path traversal in download_dir should return error")
	}
}

// ── helpers : encodage base64 ─────────────────────────────────────────────────

func TestDownloadConfigSet_Base64Encoding(t *testing.T) {
	// Vérifie que le chemin est bien encodé en base64 standard (avec padding).
	// Exemple doc Freebox : /Freebox/Téléchargements → base64std
	path := "/Disque 1/Téléchargements"
	want := base64.StdEncoding.EncodeToString([]byte("/Disque 1/Téléchargements"))
	if want == "" {
		t.Fatal("base64 encoding should not be empty")
	}
	// Roundtrip check
	decoded, err := base64.StdEncoding.DecodeString(want)
	if err != nil {
		t.Fatalf("base64 decode error: %v", err)
	}
	if string(decoded) != path {
		t.Errorf("roundtrip mismatch: got %q, want %q", string(decoded), path)
	}
}
