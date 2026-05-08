/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"encoding/base64"
	"net/url"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newDownloadsServer(t *testing.T, mock mockWriter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerDownloads(s, mock)
	return s
}

func TestDownloads_OK(t *testing.T) {
	s := newDownloadsServer(t, mockWriter{mockGetter: mockGetter{
		"/downloads/": []Download{
			{ID: 1, Name: "ubuntu.iso", Status: "done", Type: "http", Size: 1024 * 1024 * 1024},
		},
	}})
	result := callTool(t, s, "freebox_downloads")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "ubuntu.iso") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestDownloads_APIError(t *testing.T) {
	s := newDownloadsServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callTool(t, s, "freebox_downloads")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestDownloadAdd_OK(t *testing.T) {
	captured := make(map[string]url.Values)
	s := newDownloadsServer(t, mockWriter{
		mockGetter:   mockGetter{},
		postFormVals: captured,
	})
	result := callToolWithArgs(t, s, "freebox_download_add", map[string]any{
		"url": "https://example.com/fichier.zip",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	// Verifier que le param s'appelle "download_url" (pas "url")
	vals, ok := captured["/downloads/add/"]
	if !ok {
		t.Fatal("PostForm not called for /downloads/add/")
	}
	if got := vals.Get("download_url"); got != "https://example.com/fichier.zip" {
		t.Errorf("download_url = %q, want %q", got, "https://example.com/fichier.zip")
	}
	if vals.Get("download_dir") != "" {
		t.Errorf("download_dir should be absent when not provided, got %q", vals.Get("download_dir"))
	}
}

func TestDownloadAdd_WithDir_OK(t *testing.T) {
	captured := make(map[string]url.Values)
	s := newDownloadsServer(t, mockWriter{
		mockGetter:   mockGetter{},
		postFormVals: captured,
	})
	result := callToolWithArgs(t, s, "freebox_download_add", map[string]any{
		"url":          "https://example.com/fichier.zip",
		"download_dir": "/Disque dur/Téléchargements",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	vals := captured["/downloads/add/"]
	// Verifier que download_dir est base64-encode
	wantB64 := base64.StdEncoding.EncodeToString([]byte("/Disque dur/Téléchargements"))
	if got := vals.Get("download_dir"); got != wantB64 {
		t.Errorf("download_dir = %q, want base64 %q", got, wantB64)
	}
}

func TestDownloadAdd_APIError(t *testing.T) {
	s := newDownloadsServer(t, mockWriter{
		mockGetter: mockGetter{},
		postErrs:   map[string]error{"/downloads/add/": &notFoundErr{"/downloads/add/"}},
	})
	result := callToolWithArgs(t, s, "freebox_download_add", map[string]any{
		"url": "https://example.com/fichier.zip",
	})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestDownloadToggle_OK(t *testing.T) {
	s := newDownloadsServer(t, mockWriter{mockGetter: mockGetter{
		"/downloads/1": Download{ID: 1, Status: "stopped"},
	}})
	result := callToolWithArgs(t, s, "freebox_download_toggle", map[string]any{
		"id":     float64(1),
		"status": "stopped",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestDownloadDelete_OK(t *testing.T) {
	s := newDownloadsServer(t, mockWriter{mockGetter: mockGetter{}})
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

// ── downloadAddForm unit tests ────────────────────────────────────────────────

func TestDownloadAddForm_NoDir(t *testing.T) {
	form := downloadAddForm("https://example.com/f.iso", "")
	if got := form.Get("download_url"); got != "https://example.com/f.iso" {
		t.Errorf("download_url = %q", got)
	}
	if form.Get("download_dir") != "" {
		t.Errorf("download_dir should be absent, got %q", form.Get("download_dir"))
	}
}

func TestDownloadAddForm_WithDir(t *testing.T) {
	form := downloadAddForm("https://example.com/f.iso", "/Disque dur/VMs")
	want := base64.StdEncoding.EncodeToString([]byte("/Disque dur/VMs"))
	if got := form.Get("download_dir"); got != want {
		t.Errorf("download_dir = %q, want %q", got, want)
	}
}
