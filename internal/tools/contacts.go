/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Contact reflects one entry from GET /api/v4/contact/
type Contact struct {
	ID          int           `json:"id"`
	DisplayName string        `json:"display_name"`
	FirstName   string        `json:"first_name"`
	LastName    string        `json:"last_name"`
	Company     string        `json:"company"`
	Emails      []ContactAddr `json:"emails"`
	Numbers     []ContactAddr `json:"numbers"`
	Addresses   []ContactAddr `json:"addresses"`
	URLs        []ContactAddr `json:"urls"`
}

// ContactAddr is a typed address (phone, email, URL, postal).
type ContactAddr struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func registerContacts(s *server.MCPServer, c getter) {
	// ── Liste ────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_contacts",
			mcp.WithDescription("Liste les contacts du répertoire téléphonique de la Freebox (nom, prénom, société, numéros, emails). Lecture seule."),
		),
		func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var contacts []Contact
			if err := c.Get(ctx, "/contact/", &contacts); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(contacts)
		},
	)

	// ── Détail ────────────────────────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("freebox_contact_get",
			mcp.WithDescription("Récupère le détail complet d'un contact du répertoire (nom, numéros, emails, adresses, URLs)."),
			mcp.WithNumber("id",
				mcp.Required(),
				mcp.Description("ID du contact (voir freebox_contacts)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			id := req.GetInt("id", 0)
			var contact Contact
			if err := c.Get(ctx, fmt.Sprintf("/contact/%d/", id), &contact); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return jsonResult(contact)
		},
	)
}
