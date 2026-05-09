/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func newFSServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerFilesystem(s, mock)
	return s
}

func TestEncodeFSPath(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"/Freebox/VMs", encodeFSPath("/Freebox/VMs")},
		{"Freebox/VMs", encodeFSPath("/Freebox/VMs")},   // leading slash added
		{"/Freebox/VMs/", encodeFSPath("/Freebox/VMs")}, // trailing slash trimmed
	}
	for _, tc := range cases {
		got := encodeFSPath(tc.input)
		if got != tc.want {
			t.Errorf("encodeFSPath(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFSList_OK(t *testing.T) {
	p := "/Freebox/VMs"
	encoded := encodeFSPath(p)
	// L'API enveloppe les entries dans {entries:[...], parent:{...}} (#93).
	s := newFSServer(t, mockGetter{
		"/fs/ls/" + encoded: FSListResult{
			Entries: []FSEntry{
				{Name: "Fedora.qcow2", Type: "file", Size: 5368709120, Path: encodeFSPath("/Freebox/VMs/Fedora.qcow2")},
				{Name: "Fedora-Server-KVM-40-1.14.aarch64.qcow2", Type: "file", Size: 1073741824, Path: encodeFSPath("/Freebox/VMs/Fedora-Server-KVM-40-1.14.aarch64.qcow2")},
			},
			Parent: &FSEntry{Name: "VMs", Type: "dir", Path: encoded},
		},
	})
	req := callToolWithArgs(t, s, "freebox_fs_list", map[string]any{"path": p})
	if req.IsError {
		t.Fatalf("tool returned error: %v", req.Content)
	}
	if !strings.Contains(req.Content[0].(mcp.TextContent).Text, `"name": "Fedora.qcow2"`) {
		t.Errorf("unexpected result: %s", req.Content[0].(mcp.TextContent).Text)
	}
}

// TestFSList_APIWrappedShape (#93) reproduit la cause racine : l'API retourne
// un objet {entries:[...], parent:{...}}, pas un tableau brut. Avant le fix,
// le décodage échouait avec "cannot unmarshal object into Go value of type
// []tools.FSEntry". Ce test garantit que le wrapper FSListResult absorbe la
// shape réelle de l'API.
func TestFSList_APIWrappedShape(t *testing.T) {
	p := "/Disque 1/VMs"
	encoded := encodeFSPath(p)
	rawJSON := `{"entries":[{"name":"vm.qcow2","type":"file","size":1024,"path":"x","mimetype":"application/octet-stream"}],"parent":{"name":"VMs","type":"dir","path":"x"}}`
	var listing FSListResult
	if err := json.Unmarshal([]byte(rawJSON), &listing); err != nil {
		t.Fatalf("FSListResult unmarshal failed on real API shape: %v", err)
	}
	if len(listing.Entries) != 1 || listing.Entries[0].Name != "vm.qcow2" {
		t.Errorf("entries not decoded correctly: %+v", listing)
	}
	if listing.Parent == nil || listing.Parent.Name != "VMs" {
		t.Errorf("parent not decoded correctly: %+v", listing.Parent)
	}
	_ = encoded
}

func TestFSList_APIError(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_list", map[string]any{"path": "/Freebox/VMs"})
	if !req.IsError {
		t.Error("expected tool error result")
	}
}

// ── fs_mkdir : /fs/mkdir/ returns a bare string task ID ───────────────────────

func TestFSMkdir_NoError(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_mkdir", map[string]any{
		"parent": "/Freebox/Downloads",
		"name":   "test-dir",
	})
	if req.IsError {
		t.Errorf("unexpected error: %v", req.Content)
	}
}

// ── fs_delete : /fs/rm/ returns a FSTask object ───────────────────────────────

func TestFSDelete_NoError(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_delete", map[string]any{
		"path": "/Freebox/Downloads/old-file.iso",
	})
	if req.IsError {
		t.Errorf("unexpected error: %v", req.Content)
	}
	// Response must be a JSON object (FSTask), not a plain string
	text := req.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"state"`) {
		t.Errorf("expected FSTask JSON in response, got: %s", text)
	}
}

func TestFSDelete_TraversalBlocked(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_delete", map[string]any{"path": "/Freebox/../../etc"})
	if !req.IsError {
		t.Error("path traversal should return error")
	}
}

// ── fs_move ───────────────────────────────────────────────────────────────────

func TestFSMove_OK(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_move", map[string]any{
		"src_paths": []any{"/Disque 1/Téléchargements/alma.qcow2"},
		"dst_path":  "/Freebox/VMs",
	})
	if req.IsError {
		t.Errorf("unexpected error: %v", req.Content)
	}
}

func TestFSMove_DefaultMode_Skip(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_move", map[string]any{
		"src_paths": []any{"/Disque 1/file.iso"},
		"dst_path":  "/Freebox/VMs",
		"dst_mode":  "overwrite",
	})
	if req.IsError {
		t.Errorf("unexpected error: %v", req.Content)
	}
}

func TestFSMove_EmptySrcs(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_move", map[string]any{
		"src_paths": []any{},
		"dst_path":  "/Freebox/VMs",
	})
	if !req.IsError {
		t.Error("expected error for empty src_paths")
	}
}

func TestFSMove_TraversalInSrc(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_move", map[string]any{
		"src_paths": []any{"/../etc/passwd"},
		"dst_path":  "/Freebox/VMs",
	})
	if !req.IsError {
		t.Error("path traversal in src_paths should return error")
	}
}

func TestFSMove_TraversalInDst(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_move", map[string]any{
		"src_paths": []any{"/Disque 1/file.iso"},
		"dst_path":  "/Freebox/../../etc",
	})
	if !req.IsError {
		t.Error("path traversal in dst_path should return error")
	}
}

// ── fs_copy ───────────────────────────────────────────────────────────────────

func TestFSCopy_OK(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_copy", map[string]any{
		"src_paths": []any{"/Disque 1/Téléchargements/alma.qcow2"},
		"dst_path":  "/Freebox/VMs",
	})
	if req.IsError {
		t.Errorf("unexpected error: %v", req.Content)
	}
}

func TestFSCopy_TraversalBlocked(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_copy", map[string]any{
		"src_paths": []any{"/../etc/shadow"},
		"dst_path":  "/Freebox/VMs",
	})
	if !req.IsError {
		t.Error("path traversal should return error")
	}
}

// ── Sécurité : path traversal (commun) ───────────────────────────────────────

func TestSanitizeFSPath_Valid(t *testing.T) {
	cases := []string{"/Freebox/VMs", "/Freebox/Downloads/file.iso", "/mnt/usb"}
	for _, c := range cases {
		if _, err := sanitizeFSPath(c); err != nil {
			t.Errorf("sanitizeFSPath(%q) unexpected error: %v", c, err)
		}
	}
}

func TestSanitizeFSPath_TraversalRejected(t *testing.T) {
	cases := []string{"/../etc/passwd", "/Freebox/../../etc", ".."}
	for _, c := range cases {
		if _, err := sanitizeFSPath(c); err == nil {
			t.Errorf("sanitizeFSPath(%q) should have returned error", c)
		}
	}
}

func TestSanitizeFSPath_RootRejected(t *testing.T) {
	if _, err := sanitizeFSPath("/"); err == nil {
		t.Error("sanitizeFSPath('/') should have returned error")
	}
}

func TestFSList_TraversalBlocked(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_list", map[string]any{"path": "/../etc/passwd"})
	if !req.IsError {
		t.Error("path traversal should return error")
	}
}

// ── fs_info : path passe en query-string URL-encode ───────────────────────────

func TestFSInfo_OK(t *testing.T) {
	p := "/Disque dur"
	encoded := encodeFSPath(p)
	mockKey := "/fs/info/?path=" + url.QueryEscape(encoded)
	s := newFSServer(t, mockGetter{
		mockKey: FSInfo{
			Name:     "Disque dur",
			Path:     encoded,
			Type:     "dir",
			Size:     60,
			MimeType: "inode/directory",
		},
	})
	req := callToolWithArgs(t, s, "freebox_fs_info", map[string]any{"path": p})
	if req.IsError {
		t.Fatalf("tool returned error: %v", req.Content)
	}
	out := req.Content[0].(mcp.TextContent).Text
	if !strings.Contains(out, `"name": "Disque dur"`) {
		t.Errorf("missing name: %s", out)
	}
	if !strings.Contains(out, `"mimetype": "inode/directory"`) {
		t.Errorf("missing mimetype: %s", out)
	}
}

func TestFSInfo_TraversalBlocked(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_info", map[string]any{"path": "/Freebox/../../etc"})
	if !req.IsError {
		t.Error("path traversal should return error")
	}
}

func TestFSInfo_APIError(t *testing.T) {
	s := newFSServer(t, mockGetter{})
	req := callToolWithArgs(t, s, "freebox_fs_info", map[string]any{"path": "/Freebox/Downloads"})
	if !req.IsError {
		t.Error("expected tool error result")
	}
}

// ── fs_rename : /fs/rename/ rename in-place (#94) ────────────────────────────

// newFSWriterServer wraps registerFilesystem with a mockWriter that captures
// POST bodies so tests can verify the wire format.
func newFSWriterServer(t *testing.T, mock mockWriter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerFilesystem(s, mock)
	return s
}

func TestFSRename_PostBodyShape(t *testing.T) {
	bodies := map[string]any{}
	s := newFSWriterServer(t, mockWriter{
		mockGetter: mockGetter{},
		postBodies: bodies,
	})
	callToolWithArgs(t, s, "freebox_fs_rename", map[string]any{
		"src_path": "/Disque 1/VMs/old.qcow2",
		"new_name": "new.qcow2",
	})
	body, ok := bodies["/fs/rename/"].(map[string]any)
	if !ok {
		t.Fatalf("POST body type = %T, want map", bodies["/fs/rename/"])
	}
	wantSrc := encodeFSPath("/Disque 1/VMs/old.qcow2")
	if body["src"] != wantSrc {
		t.Errorf("src = %q, want %q (base64-encoded)", body["src"], wantSrc)
	}
	if body["dst"] != "new.qcow2" {
		t.Errorf("dst = %q, want %q (plain text basename)", body["dst"], "new.qcow2")
	}
}

func TestFSRename_RejectsPathInName(t *testing.T) {
	s := newFSWriterServer(t, mockWriter{mockGetter: mockGetter{}})
	cases := []string{"a/b", "a\\b", "../etc/passwd", "..hidden"}
	for _, name := range cases {
		req := callToolWithArgs(t, s, "freebox_fs_rename", map[string]any{
			"src_path": "/Disque 1/x",
			"new_name": name,
		})
		if !req.IsError {
			t.Errorf("new_name=%q should be rejected", name)
		}
	}
}

func TestFSRename_RejectsTraversalInSrc(t *testing.T) {
	s := newFSWriterServer(t, mockWriter{mockGetter: mockGetter{}})
	req := callToolWithArgs(t, s, "freebox_fs_rename", map[string]any{
		"src_path": "/Disque 1/../etc/passwd",
		"new_name": "x.qcow2",
	})
	if !req.IsError {
		t.Error("path traversal in src_path should be rejected")
	}
}
