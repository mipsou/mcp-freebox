/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// mockWriter extends mockGetter with configurable mutation behaviour.
type mockWriter struct {
	mockGetter
	postErrs     map[string]error
	putErrs      map[string]error
	deleteErrs   map[string]error
	postFormVals map[string]url.Values // captures PostForm calls for assertion
	postBodies   map[string]any        // captures Post body for assertion
}

func (m mockWriter) Post(_ context.Context, path string, body, _ any) error {
	if m.postBodies != nil {
		m.postBodies[path] = body
	}
	if err, ok := m.postErrs[path]; ok {
		return err
	}
	return nil
}

// PostForm captures the form values for later assertion and returns nil unless
// an error is configured via postErrs for the same path.
func (m mockWriter) PostForm(_ context.Context, path string, values url.Values, _ any) error {
	if m.postFormVals != nil {
		m.postFormVals[path] = values
	}
	if err, ok := m.postErrs[path]; ok {
		return err
	}
	return nil
}

func (m mockWriter) Put(_ context.Context, path string, _, _ any) error {
	if err, ok := m.putErrs[path]; ok {
		return err
	}
	return nil
}

func (m mockWriter) Delete(_ context.Context, path string) error {
	if err, ok := m.deleteErrs[path]; ok {
		return err
	}
	return nil
}

func newVMServer(t *testing.T, mock mockWriter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerVM(s, mock)
	return s
}

func callToolWithArgs(t *testing.T, s *server.MCPServer, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	st := s.GetTool(name)
	if st == nil {
		t.Fatalf("tool %q not registered", name)
	}
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	result, err := st.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	return result
}

func TestVMList_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{
		"/vm/": []VM{
			{ID: 0, Name: "HomeAssistant", Status: "running", Memory: 2048, Vcpus: 2, OS: "debian"},
		},
	}})
	result := callTool(t, s, "freebox_vm_list")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"name": "HomeAssistant"`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVMList_APIError(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callTool(t, s, "freebox_vm_list")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestVMStart_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_start", map[string]any{"id": float64(0)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"started": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVMStart_APIError(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{},
		postErrs:   map[string]error{"/vm/0/start": fmt.Errorf("permission denied")},
	})
	result := callToolWithArgs(t, s, "freebox_vm_start", map[string]any{"id": float64(0)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestVMStop_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_stop", map[string]any{"id": float64(1)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"stopping": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVMStop_APIError(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{},
		postErrs:   map[string]error{"/vm/1/stop": fmt.Errorf("vm not running")},
	})
	result := callToolWithArgs(t, s, "freebox_vm_stop", map[string]any{"id": float64(1)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestVMKill_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_kill", map[string]any{"id": float64(0)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, `"killed": true`) {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVMCreate_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{
		"/vm/": VM{ID: 2, Name: "test-vm", Status: "stopped", Memory: 1024, Vcpus: 1},
	}})
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name":      "test-vm",
		"memory":    float64(1024),
		"vcpus":     float64(1),
		"disk_name": "test.qcow2",
		"disk_dir":  "/Disque 1/VMs/",
		"disk_type": "qcow2",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestVMCreate_APIError(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{},
		postErrs:   map[string]error{"/vm/": fmt.Errorf("disk not found")},
	})
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name": "fail", "memory": float64(512), "vcpus": float64(1),
		"disk_name": "fail.qcow2", "disk_dir": "/Disque 1/VMs/", "disk_type": "qcow2",
	})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// ── Sécurité : disk_name — chemin structurellement contraint ─────────────────

func TestValidateDiskName_Valid(t *testing.T) {
	cases := []string{"fedora.qcow2", "debian-12.raw", "my_vm.qcow2", "Ubuntu-22.04.raw"}
	for _, c := range cases {
		if err := validateDiskName(c); err != nil {
			t.Errorf("validateDiskName(%q) unexpected error: %v", c, err)
		}
	}
}

func TestValidateDiskName_Invalid(t *testing.T) {
	cases := []string{
		"../etc/passwd",
		"/Freebox/VMs/hack.qcow2", // path separator
		"no-extension",
		"bad.vmdk",
		"",
	}
	for _, c := range cases {
		if err := validateDiskName(c); err == nil {
			t.Errorf("validateDiskName(%q) should have returned error", c)
		}
	}
}

func TestVMCreate_InvalidDiskName(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name": "hack", "memory": float64(512), "vcpus": float64(1),
		"disk_name": "../etc/qemu.qcow2", "disk_dir": "/Disque 1/VMs/", "disk_type": "qcow2",
	})
	if !result.IsError {
		t.Error("path traversal in disk_name should return error")
	}
}

func TestVMCreate_MissingDiskDir(t *testing.T) {
	// disk_dir est requis : pas de défaut hardcodé. L'absence du paramètre doit
	// retourner une erreur explicite plutôt que de masquer un mismatch de config
	// (les chemins de stockage varient selon le modèle de Freebox).
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name": "no-dir", "memory": float64(512), "vcpus": float64(1),
		"disk_name": "no-dir.qcow2", "disk_type": "qcow2",
	})
	if !result.IsError {
		t.Fatal("missing disk_dir should return error")
	}
	got := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(got, "disk_dir") {
		t.Errorf("error should mention disk_dir, got: %s", got)
	}
}

func TestVMCreate_DiskDir_DisqueExterne(t *testing.T) {
	// Caller-supplied disk_dir = /Disque 1/VMs/ (Freebox avec disque externe).
	bodies := map[string]any{}
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{"/vm/": VM{ID: 3, Name: "haos"}},
		postBodies: bodies,
	})
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name": "haos", "memory": float64(2048), "vcpus": float64(2),
		"disk_name": "haos.qcow2", "disk_dir": "/Disque 1/VMs/", "disk_type": "qcow2",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	creq, ok := bodies["/vm/"].(vmCreateRequest)
	if !ok {
		t.Fatalf("expected vmCreateRequest body, got %T", bodies["/vm/"])
	}
	want := base64.StdEncoding.EncodeToString([]byte("/Disque 1/VMs/haos.qcow2"))
	if creq.DiskPath != want {
		t.Errorf("DiskPath = %q, want base64 %q", creq.DiskPath, want)
	}
}

func TestVMCreate_DiskDir_StockageInterne(t *testing.T) {
	// Caller-supplied disk_dir = /Freebox/VMs/ (Freebox avec stockage interne).
	bodies := map[string]any{}
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{"/vm/": VM{ID: 4, Name: "haos-int"}},
		postBodies: bodies,
	})
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name": "haos-int", "memory": float64(2048), "vcpus": float64(2),
		"disk_name": "haos.qcow2", "disk_type": "qcow2",
		"disk_dir": "/Freebox/VMs/",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	creq, ok := bodies["/vm/"].(vmCreateRequest)
	if !ok {
		t.Fatalf("expected vmCreateRequest body, got %T", bodies["/vm/"])
	}
	want := base64.StdEncoding.EncodeToString([]byte("/Freebox/VMs/haos.qcow2"))
	if creq.DiskPath != want {
		t.Errorf("DiskPath = %q, want base64 %q", creq.DiskPath, want)
	}
}

func TestVMCreate_DiskDir_TraversalBlocked(t *testing.T) {
	// Verify path traversal in disk_dir is rejected.
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name": "hack", "memory": float64(512), "vcpus": float64(1),
		"disk_name": "hack.qcow2", "disk_type": "qcow2",
		"disk_dir": "/../etc/",
	})
	if !result.IsError {
		t.Error("path traversal in disk_dir should return error")
	}
}

func TestVMUpdate_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{
		"/vm/0": VM{ID: 0, Name: "renamed", Memory: 4096, Vcpus: 4},
	}})
	result := callToolWithArgs(t, s, "freebox_vm_update", map[string]any{
		"id": float64(0), "name": "renamed", "memory": float64(4096),
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestVMDelete_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_delete", map[string]any{"id": float64(3)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "VM 3 supprimée") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVMDelete_APIError(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{},
		deleteErrs: map[string]error{"/vm/3": fmt.Errorf("vm not found")},
	})
	result := callToolWithArgs(t, s, "freebox_vm_delete", map[string]any{"id": float64(3)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// ── vm_create : cloud-init + cd_path + bind_usb_ports ────────────────────────

func TestVMCreate_CloudInit_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{
		"/vm/": VM{ID: 5, Name: "alma-pra", CloudinitEnabled: true},
	}})
	userdata := "#cloud-config\nhostname: alma-pra\nssh_authorized_keys:\n  - ssh-ed25519 AAAA...\n"
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name":               "alma-pra",
		"memory":             float64(768),
		"vcpus":              float64(2),
		"disk_name":          "alma9.qcow2",
		"disk_dir":           "/Disque 1/VMs/",
		"disk_type":          "qcow2",
		"cloudinit_userdata": userdata,
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestVMCreate_CloudInit_TooLong(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	userdata := strings.Repeat("x", maxCloudinitLen+1)
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name":               "big-cloud",
		"memory":             float64(512),
		"vcpus":              float64(1),
		"disk_name":          "big.qcow2",
		"disk_dir":           "/Disque 1/VMs/",
		"disk_type":          "qcow2",
		"cloudinit_userdata": userdata,
	})
	if !result.IsError {
		t.Error("expected error for oversized cloudinit_userdata")
	}
}

func TestVMCreate_CDPath_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{
		"/vm/": VM{ID: 6, Name: "install-vm"},
	}})
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name":      "install-vm",
		"memory":    float64(1024),
		"vcpus":     float64(2),
		"disk_name": "debian.qcow2",
		"disk_dir":  "/Disque 1/VMs/",
		"disk_type": "qcow2",
		"cd_path":   "/Disque 1/ISO/debian-12-arm64.iso",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
}

func TestVMCreate_CDPath_TraversalBlocked(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_create", map[string]any{
		"name":      "hack",
		"memory":    float64(512),
		"vcpus":     float64(1),
		"disk_name": "hack.qcow2",
		"disk_dir":  "/Disque 1/VMs/",
		"disk_type": "qcow2",
		"cd_path":   "/../etc/passwd",
	})
	if !result.IsError {
		t.Error("path traversal in cd_path should return error")
	}
}

// ── BindUSBPorts : custom unmarshal (issue #76) ──────────────────────────────
//
// The Freebox API returns "" (empty string) instead of [] when no USB port is
// bound to a VM. The custom UnmarshalJSON must accept both shapes plus null.

func TestBindUSBPorts_Unmarshal_EmptyString(t *testing.T) {
	var v VM
	if err := json.Unmarshal([]byte(`{"bind_usb_ports": ""}`), &v); err != nil {
		t.Fatalf("unmarshal empty string failed: %v", err)
	}
	if len(v.BindUSBPorts) != 0 {
		t.Errorf("expected empty BindUSBPorts, got %v", v.BindUSBPorts)
	}
}

func TestBindUSBPorts_Unmarshal_Null(t *testing.T) {
	var v VM
	if err := json.Unmarshal([]byte(`{"bind_usb_ports": null}`), &v); err != nil {
		t.Fatalf("unmarshal null failed: %v", err)
	}
	if v.BindUSBPorts != nil {
		t.Errorf("expected nil BindUSBPorts, got %v", v.BindUSBPorts)
	}
}

func TestBindUSBPorts_Unmarshal_Array(t *testing.T) {
	var v VM
	if err := json.Unmarshal([]byte(`{"bind_usb_ports": ["usb-external-type-c", "usb-external-type-a"]}`), &v); err != nil {
		t.Fatalf("unmarshal array failed: %v", err)
	}
	want := []string{"usb-external-type-c", "usb-external-type-a"}
	if len(v.BindUSBPorts) != len(want) || v.BindUSBPorts[0] != want[0] || v.BindUSBPorts[1] != want[1] {
		t.Errorf("BindUSBPorts = %v, want %v", v.BindUSBPorts, want)
	}
}

func TestBindUSBPorts_Unmarshal_EmptyArray(t *testing.T) {
	var v VM
	if err := json.Unmarshal([]byte(`{"bind_usb_ports": []}`), &v); err != nil {
		t.Fatalf("unmarshal empty array failed: %v", err)
	}
	if len(v.BindUSBPorts) != 0 {
		t.Errorf("expected empty BindUSBPorts, got %v", v.BindUSBPorts)
	}
}

func TestBindUSBPorts_Unmarshal_InvalidString(t *testing.T) {
	var v VM
	err := json.Unmarshal([]byte(`{"bind_usb_ports": "not-empty"}`), &v)
	if err == nil {
		t.Error("expected unmarshal error for non-empty string, got nil")
	}
}

// Real-world payload captured during runtime test on Freebox firmware (PR #75
// validation): ensures the full VM list response decodes without error when
// bind_usb_ports comes back as an empty string.
func TestVMList_RealPayload_BindUSBPortsEmptyString(t *testing.T) {
	payload := `[{"id":0,"name":"test-vm","status":"stopped","memory":256,"vcpus":1,` +
		`"disk_path":"L0Rpc3F1ZSAxL1ZNcy90ZXN0LnFjb3cy","disk_type":"qcow2",` +
		`"os":"unknown","enable_screen":false,"cloudinit_enabled":false,` +
		`"bind_usb_ports":""}]`
	var vms []VM
	if err := json.Unmarshal([]byte(payload), &vms); err != nil {
		t.Fatalf("unmarshal real-world VM list failed: %v", err)
	}
	if len(vms) != 1 || vms[0].Name != "test-vm" {
		t.Errorf("unexpected decode: %+v", vms)
	}
}
