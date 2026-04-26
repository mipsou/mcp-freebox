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

func newContactsServer(t *testing.T, mock mockGetter) *server.MCPServer {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.0")
	registerContacts(s, mock)
	return s
}

func TestContacts_OK(t *testing.T) {
	s := newContactsServer(t, mockGetter{
		"/contact/": []Contact{
			{ID: 1, DisplayName: "Jean Dupont", FirstName: "Jean", LastName: "Dupont",
				Numbers: []ContactAddr{{Type: "home", Value: "0612345678"}}},
		},
	})
	result := callTool(t, s, "freebox_contacts")
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "Jean Dupont") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestContacts_APIError(t *testing.T) {
	s := newContactsServer(t, mockGetter{})
	result := callTool(t, s, "freebox_contacts")
	if !result.IsError {
		t.Error("expected tool error result")
	}
}

func TestContactGet_OK(t *testing.T) {
	s := newContactsServer(t, mockGetter{
		"/contact/1/": Contact{ID: 1, DisplayName: "Jean Dupont", FirstName: "Jean", LastName: "Dupont"},
	})
	result := callToolWithArgs(t, s, "freebox_contact_get", map[string]any{"id": float64(1)})
	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}
	if !strings.Contains(result.Content[0].(mcp.TextContent).Text, "Jean Dupont") {
		t.Errorf("unexpected result: %s", result.Content[0].(mcp.TextContent).Text)
	}
}

func TestContactGet_APIError(t *testing.T) {
	s := newContactsServer(t, mockGetter{})
	result := callToolWithArgs(t, s, "freebox_contact_get", map[string]any{"id": float64(99)})
	if !result.IsError {
		t.Error("expected tool error result")
	}
}
