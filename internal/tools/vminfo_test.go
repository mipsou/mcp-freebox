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

func newVMInfoServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerVMInfo(s, mock)
	return s
}

func TestVMInfo_OK(t *testing.T) {
	s := newVMInfoServer(t, mockGetter{
		"/vm/info": VMInfo{
			UsedCPUs:    0,
			TotalCPUs:   2,
			UsedMemory:  0,
			TotalMemory: 1024,
			UsbUsed:     false,
			SataUsed:    false,
			UsbPorts:    []string{"usb-external-type-a", "usb-external-type-c"},
			SataPorts:   []string{"sata-internal-p0", "sata-internal-p1"},
		},
	})
	result := callTool(t, s, "freebox_vm_info")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	out := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(out, `"total_cpus": 2`) {
		t.Errorf("missing total_cpus: %s", out)
	}
	if !strings.Contains(out, `"sata-internal-p0"`) {
		t.Errorf("missing sata port: %s", out)
	}
}

func TestVMInfo_APIError(t *testing.T) {
	s := newVMInfoServer(t, mockGetter{})
	result := callTool(t, s, "freebox_vm_info")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestVMDistros_OK(t *testing.T) {
	s := newVMInfoServer(t, mockGetter{
		"/vm/distros/": []VMDistro{
			{
				Name: "Ubuntu 24.04 LTS (Noble)",
				OS:   "ubuntu",
				URL:  "http://ftp.free.fr/.private/ubuntu-cloud/releases/noble/release/ubuntu-24.04-server-cloudimg-arm64.img",
				Hash: "http://ftp.free.fr/.private/ubuntu-cloud/releases/noble/release/SHA256SUMS",
			},
			{
				Name: "Debian 12 (Bookworm)",
				OS:   "debian",
				URL:  "https://cloud.debian.org/images/cloud/bookworm/daily/latest/debian-12-generic-arm64-daily.qcow2",
				Hash: "https://cloud.debian.org/images/cloud/bookworm/daily/latest/SHA512SUMS",
			},
		},
	})
	result := callTool(t, s, "freebox_vm_distros")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	out := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(out, `"os": "ubuntu"`) {
		t.Errorf("missing os: %s", out)
	}
	if !strings.Contains(out, `"name": "Debian 12 (Bookworm)"`) {
		t.Errorf("missing name: %s", out)
	}
}

func TestVMDistros_APIError(t *testing.T) {
	s := newVMInfoServer(t, mockGetter{})
	result := callTool(t, s, "freebox_vm_distros")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
