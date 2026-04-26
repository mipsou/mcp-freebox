/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package tools

import (
	"context"
	"fmt"
	"regexp"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	macRegex      = regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)
	secureOnRegex = regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)
)

func validateMAC(mac string) error {
	if !macRegex.MatchString(mac) {
		return fmt.Errorf("adresse MAC invalide '%s' : format attendu aa:bb:cc:dd:ee:ff", mac)
	}
	return nil
}

func validateSecureOn(password string) error {
	if password == "" {
		return nil
	}
	if len(password) > 17 { // "xx:xx:xx:xx:xx:xx" = 17 chars max
		return fmt.Errorf("mot de passe SecureOn trop long : maximum 17 caractères (format xx:xx:xx:xx:xx:xx)")
	}
	if !secureOnRegex.MatchString(password) {
		return fmt.Errorf("mot de passe SecureOn invalide '%s' : format attendu aa:bb:cc:dd:ee:ff", password)
	}
	return nil
}

// WolRequest is the body for POST /api/v4/lan/wol/{iface}/
type WolRequest struct {
	MACAddr  string `json:"mac"`
	Password string `json:"password,omitempty"` // SecureOn password optionnel
}

func registerWOL(s *server.MCPServer, c writer) {
	s.AddTool(
		mcp.NewTool("freebox_wol",
			mcp.WithDescription("Envoie un paquet Wake-on-LAN (magic packet) à un équipement du réseau local via son adresse MAC. Utile pour démarrer un NAS ou un PC à distance."),
			mcp.WithString("mac",
				mcp.Required(),
				mcp.Description("Adresse MAC de l'équipement à réveiller (ex: aa:bb:cc:dd:ee:ff)")),
			mcp.WithString("iface",
				mcp.Description("Interface réseau (défaut: pub — interface LAN principale)")),
			mcp.WithString("password",
				mcp.Description("Mot de passe SecureOn optionnel (6 octets hex, ex: 00:11:22:33:44:55)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			mac := req.GetString("mac", "")
			if err := validateMAC(mac); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			iface := req.GetString("iface", "pub")
			password := req.GetString("password", "")
			if err := validateSecureOn(password); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			body := WolRequest{MACAddr: mac, Password: password}
			var result any
			if err := c.Post(ctx, "/lan/wol/"+iface+"/", body, &result); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText("Paquet Wake-on-LAN envoyé à " + mac + " via " + iface + "."), nil
		},
	)
}
