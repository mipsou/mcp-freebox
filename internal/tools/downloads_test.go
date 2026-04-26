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

func newDownloadsServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerDownloads(s, mock)
	return s
}

func TestDownloads_OK(t *testing.T) {
	s := newDownloadsServer(t, mockGetter{
		"/downloads/": []Download{
			{ID: 1, Name: "ubuntu.iso", Status: "done", Type: "http", Size: 1024 * 1024 * 1024},
		},
	})
	result := callTool(t, s, "freebox_downloads")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "ubuntu.iso") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDownloads_APIError(t *testing.T) {
	s := newDownloadsServer(t, mockGetter{})
	result := callTool(t, s, "freebox_downloads")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestDownloadAdd_OK(t *testing.T) {
	s := newDownloadsServer(t, mockGetter{
		"/downloads/add/": Download{ID: 2, Name: "fichier.zip", Status: "downloading"},
	})
	result := callToolWithArgs(t, s, "freebox_download_add", map[string]any{
		"url": "https://example.com/fichier.zip",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestDownloadToggle_OK(t *testing.T) {
	s := newDownloadsServer(t, mockGetter{
		"/downloads/1": Download{ID: 1, Status: "stopped"},
	})
	result := callToolWithArgs(t, s, "freebox_download_toggle", map[string]any{
		"id":     float64(1),
		"status": "stopped",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestDownloadDelete_OK(t *testing.T) {
	s := newDownloadsServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_download_delete", map[string]any{
		"id": float64(1),
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "supprimé") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

// ── Sécurité : validation URL (SSRF) ─────────────────────────────────────────

func TestValidateDownloadURL_ValidHTTPS(t *testing.T) {
	if err := validateDownloadURL("https://example.com/file.zip"); err != nil {
		t.Errorf("unexpected error for valid HTTPS URL: %v", err)
	}
}

func TestValidateDownloadURL_ValidMagnet(t *testing.T) {
	if err := validateDownloadURL("magnet:?xt=urn:btih:abc123&dn=test"); err != nil {
		t.Errorf("unexpected error for valid magnet link: %v", err)
	}
}

func TestValidateDownloadURL_FileSchemeBlocked(t *testing.T) {
	if err := validateDownloadURL("file:///etc/passwd"); err == nil {
		t.Error("file:// scheme should be blocked")
	}
}

func TestValidateDownloadURL_GopherBlocked(t *testing.T) {
	if err := validateDownloadURL("gopher://evil.com/"); err == nil {
		t.Error("gopher:// scheme should be blocked")
	}
}

func TestValidateDownloadURL_LoopbackBlocked(t *testing.T) {
	if err := validateDownloadURL("http://127.0.0.1/secret"); err == nil {
		t.Error("loopback address should be blocked")
	}
}

func TestValidateDownloadURL_LinkLocalBlocked(t *testing.T) {
	if err := validateDownloadURL("http://169.254.169.254/latest/meta-data/"); err == nil {
		t.Error("link-local SSRF target should be blocked")
	}
}
