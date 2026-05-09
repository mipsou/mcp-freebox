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
	putBodies    map[string]any        // captures Put body for assertion
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

func (m mockWriter) Put(_ context.Context, path string, body, _ any) error {
	if m.putBodies != nil {
		m.putBodies[path] = body
	}
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

// TestVMUpdate_ReadModifyWrite_PreservesUnchanged guards #80: l'API Freebox
// PUT /vm/{id} exige un body complet. Le handler doit GET l'état courant,
// patcher les champs demandés, et PUT le tout — sans perdre les champs non
// patchés (disk_path, os, etc.).
func TestVMUpdate_ReadModifyWrite_PreservesUnchanged(t *testing.T) {
	bodies := map[string]any{}
	current := VM{
		ID: 3, Name: "old", Status: "stopped",
		Memory: 256, Vcpus: 1,
		DiskPath: "L0Rpc3F1ZSAxL1ZNcy90ZXN0LnFjb3cy", DiskType: "qcow2", OS: "debian",
		EnableScreen: false,
	}
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{"/vm/3": current},
		putBodies:  bodies,
	})
	callToolWithArgs(t, s, "freebox_vm_update", map[string]any{
		"id": float64(3), "name": "renamed", "memory": float64(512),
	})
	body, ok := bodies["/vm/3"].(VM)
	if !ok {
		t.Fatalf("PUT body type = %T, want VM (full struct)", bodies["/vm/3"])
	}
	if body.Name != "renamed" {
		t.Errorf("name = %q, want renamed", body.Name)
	}
	if body.Memory != 512 {
		t.Errorf("memory = %d, want 512", body.Memory)
	}
	if body.Vcpus != 1 {
		t.Errorf("vcpus = %d, want 1 (unchanged)", body.Vcpus)
	}
	if body.DiskPath != current.DiskPath {
		t.Errorf("disk_path = %q, want preserved %q", body.DiskPath, current.DiskPath)
	}
	if body.OS != "debian" {
		t.Errorf("os = %q, want preserved debian", body.OS)
	}
	if body.ID != 3 {
		t.Errorf("id = %d, want 3 (echoed back)", body.ID)
	}
}

// TestVMUpdate_EnableScreenToggle vérifie que enable_screen passe correctement
// de false à true dans le PUT body après modification.
func TestVMUpdate_EnableScreenToggle(t *testing.T) {
	bodies := map[string]any{}
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{"/vm/0": VM{ID: 0, Name: "vm", EnableScreen: false}},
		putBodies:  bodies,
	})
	callToolWithArgs(t, s, "freebox_vm_update", map[string]any{
		"id": float64(0), "enable_screen": true,
	})
	body := bodies["/vm/0"].(VM)
	if !body.EnableScreen {
		t.Errorf("enable_screen = %v, want true", body.EnableScreen)
	}
}

// TestVMUpdate_GetError propage proprement l'erreur du GET initial.
func TestVMUpdate_GetError(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{}, // /vm/99 absent → erreur GET
	})
	result := callToolWithArgs(t, s, "freebox_vm_update", map[string]any{
		"id": float64(99), "name": "x",
	})
	if !result.IsError {
		t.Error("expected error when GET /vm/99 fails")
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

// ── vm_disk_resize / vm_disk_task (#85) ──────────────────────────────────────

func TestVMDiskResize_PostBodyShape(t *testing.T) {
	bodies := map[string]any{}
	const diskPath = "L0Rpc3F1ZSAxL1ZNcy90ZXN0LnFjb3cy" // base64 du chemin
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{
			"/vm/3":           VM{ID: 3, Status: "stopped", DiskPath: diskPath},
			"/vm/disk/resize": VMDiskTask{ID: 42, Type: "resize_disk", State: "queued"},
		},
		postBodies: bodies,
	})
	result := callToolWithArgs(t, s, "freebox_vm_disk_resize", map[string]any{
		"id":           float64(3),
		"size_gb":      float64(20),
		"allow_shrink": false,
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	body, ok := bodies["/vm/disk/resize"].(vmDiskResizeRequest)
	if !ok {
		t.Fatalf("POST body type = %T, want vmDiskResizeRequest", bodies["/vm/disk/resize"])
	}
	if body.DiskPath != diskPath {
		t.Errorf("disk_path = %q, want %q (preserved from VM)", body.DiskPath, diskPath)
	}
	if body.Size != 20*1024*1024*1024 {
		t.Errorf("size = %d, want 21474836480 (20 GiB)", body.Size)
	}
	if body.ShrinkAllow {
		t.Error("shrink_allow should be false by default")
	}
}

func TestVMDiskResize_RejectsRunningVM(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{
			"/vm/0": VM{ID: 0, Status: "running", DiskPath: "x"},
		},
	})
	result := callToolWithArgs(t, s, "freebox_vm_disk_resize", map[string]any{
		"id": float64(0), "size_gb": float64(10),
	})
	if !result.IsError {
		t.Error("expected error for running VM")
	}
}

func TestVMDiskResize_RejectsZeroSize(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{"/vm/0": VM{ID: 0, Status: "stopped"}},
	})
	result := callToolWithArgs(t, s, "freebox_vm_disk_resize", map[string]any{
		"id": float64(0), "size_gb": float64(0),
	})
	if !result.IsError {
		t.Error("expected error for size_gb=0")
	}
}

func TestVMDiskResize_AllowShrinkPropagates(t *testing.T) {
	bodies := map[string]any{}
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{
			"/vm/0":           VM{ID: 0, Status: "stopped", DiskPath: "x"},
			"/vm/disk/resize": VMDiskTask{ID: 1},
		},
		postBodies: bodies,
	})
	callToolWithArgs(t, s, "freebox_vm_disk_resize", map[string]any{
		"id": float64(0), "size_gb": float64(5), "allow_shrink": true,
	})
	body := bodies["/vm/disk/resize"].(vmDiskResizeRequest)
	if !body.ShrinkAllow {
		t.Error("shrink_allow=true should propagate")
	}
}

func TestVMDiskCreate_PostBodyShape(t *testing.T) {
	bodies := map[string]any{}
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{"/vm/disk/create": VMDiskTask{ID: 7, Type: "create_disk"}},
		postBodies: bodies,
	})
	result := callToolWithArgs(t, s, "freebox_vm_disk_create", map[string]any{
		"disk_name": "fresh.qcow2",
		"disk_dir":  "/Disque 1/VMs/",
		"size_gb":   float64(5),
		"disk_type": "qcow2",
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	body := bodies["/vm/disk/create"].(vmDiskCreateRequest)
	if body.Size != 5*1024*1024*1024 {
		t.Errorf("size = %d, want 5 GiB", body.Size)
	}
	if body.DiskType != "qcow2" {
		t.Errorf("disk_type = %q, want qcow2", body.DiskType)
	}
	// disk_path doit être base64 du chemin propre
	wantPath := base64.StdEncoding.EncodeToString([]byte("/Disque 1/VMs/fresh.qcow2"))
	if body.DiskPath != wantPath {
		t.Errorf("disk_path = %q, want %q", body.DiskPath, wantPath)
	}
}

func TestVMDiskCreate_RejectsBadDiskName(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_disk_create", map[string]any{
		"disk_name": "../etc/passwd.qcow2",
		"disk_dir":  "/Disque 1/VMs/",
		"size_gb":   float64(1),
		"disk_type": "qcow2",
	})
	if !result.IsError {
		t.Error("path traversal in disk_name should error")
	}
}

func TestVMDiskTask_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{
			"/vm/disk/task/42": VMDiskTask{ID: 42, Type: "resize_disk", State: "done", Progress: 100, Done: true},
		},
	})
	result := callToolWithArgs(t, s, "freebox_vm_disk_task", map[string]any{
		"task_id": float64(42),
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, `"state": "done"`) {
		t.Errorf("task content missing state=done: %s", text)
	}
}

func TestVMDiskTaskDelete_OK(t *testing.T) {
	s := newVMServer(t, mockWriter{mockGetter: mockGetter{}})
	result := callToolWithArgs(t, s, "freebox_vm_disk_task_delete", map[string]any{
		"task_id": float64(99),
	})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "99") {
		t.Errorf("expected task id in response, got: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestVMDiskTaskDelete_APIError(t *testing.T) {
	s := newVMServer(t, mockWriter{
		mockGetter: mockGetter{},
		deleteErrs: map[string]error{"/vm/disk/task/99": fmt.Errorf("task not found")},
	})
	result := callToolWithArgs(t, s, "freebox_vm_disk_task_delete", map[string]any{
		"task_id": float64(99),
	})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
