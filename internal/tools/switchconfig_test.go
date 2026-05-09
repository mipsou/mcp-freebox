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

func newSwitchConfigServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerSwitchConfig(s, mock)
	return s
}

func TestSwitchPortConfig_OK(t *testing.T) {
	s := newSwitchConfigServer(t, mockGetter{
		"/switch/port/1": SwitchPortConfig{ID: 1, DuplexMode: "auto", SpeedMode: "auto"},
	})
	result := callToolWithArgs(t, s, "freebox_switch_port_config", map[string]any{"id": float64(1)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "auto") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSwitchPortConfig_APIError(t *testing.T) {
	s := newSwitchConfigServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_switch_port_config", map[string]any{"id": float64(1)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestSwitchPortStats_OK(t *testing.T) {
	s := newSwitchConfigServer(t, mockGetter{
		"/switch/port/2/stats": SwitchStats{RxBytesRate: 1000000, TxBytesRate: 500000},
	})
	result := callToolWithArgs(t, s, "freebox_switch_port_stats", map[string]any{"id": float64(2)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "1000000") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestSwitchPortStats_APIError(t *testing.T) {
	s := newSwitchConfigServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_switch_port_stats", map[string]any{"id": float64(1)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

// Verifie que le schema enrichi (firmware 4.9.18.1) expose les compteurs
// supplementaires : volumes Rx/Tx absolus, erreurs detaillees, pause frames.
func TestSwitchPortStats_RichSchema(t *testing.T) {
	s := newSwitchConfigServer(t, mockGetter{
		"/switch/port/1/stats": SwitchStats{
			RxBytes:            189258176443,
			TxBytes:            578548554716,
			RxPackets:          269342561,
			TxPackets:          426518466,
			RxFCSPackets:       0,
			TxCollisions:       0,
			RxOversizePackets:  0,
			RxUndersizePackets: 0,
			TxPause:            7,
		},
	})
	result := callToolWithArgs(t, s, "freebox_switch_port_stats", map[string]any{"id": float64(1)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	out := result.Content[0].(mcp.TextContent).Text
	for _, want := range []string{
		`"rx_good_bytes": 189258176443`,
		`"tx_bytes": 578548554716`,
		`"rx_good_packets": 269342561`,
		`"tx_pause": 7`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing field %q in: %s", want, out)
		}
	}
}
