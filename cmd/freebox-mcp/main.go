/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/server"

	"github.com/mipsou/mcp-freebox/internal/auth"
	"github.com/mipsou/mcp-freebox/internal/client"
	"github.com/mipsou/mcp-freebox/internal/config"
	"github.com/mipsou/mcp-freebox/internal/mdns"
	"github.com/mipsou/mcp-freebox/internal/pair"
	"github.com/mipsou/mcp-freebox/internal/tools"
	"github.com/mipsou/mcp-freebox/internal/wincred"
)

var version = "dev"

const (
	credTarget  = "freebox-mcp"
	credUser    = "app"
	pairTimeout = 3 * time.Minute
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-mcp: config error: %v\n", err)
		os.Exit(1)
	}

	// Découverte mDNS si FREEBOX_HOST non défini — fallback mafreebox.freebox.fr.
	if os.Getenv("FREEBOX_HOST") == "" {
		fmt.Fprintln(os.Stderr, "freebox-mcp: découverte mDNS...")
		discovered, err := mdns.Discover(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "freebox-mcp: mDNS indisponible (%v) — fallback %s\n", err, cfg.Host)
		} else {
			cfg.Host = discovered.Host
			if discovered.HTTPSPort != 0 && discovered.HTTPSPort != 443 {
				cfg.Host = fmt.Sprintf("%s:%d", discovered.Host, discovered.HTTPSPort)
			}
			fmt.Fprintf(os.Stderr, "freebox-mcp: Freebox trouvée → %s\n", cfg.Host)
		}
	}

	// Freebox uses a self-signed cert on mafreebox.freebox.fr;
	// InsecureSkipVerify is intentional and documented.
	//nolint:gosec
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	appToken, err := acquireToken(cfg, httpClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-mcp: impossible d'obtenir un token: %v\n", err)
		os.Exit(1)
	}

	mgr := auth.New(cfg.BaseURL(), cfg.AppID, appToken, httpClient)
	c := client.New(cfg.BaseURL(), cfg.DiscoveryURL(), mgr, httpClient)

	s := server.NewMCPServer("freebox-mcp", version,
		server.WithToolCapabilities(false),
	)
	tools.RegisterAll(s, c, c)

	fmt.Fprintf(os.Stderr, "freebox-mcp: starting (host=%s, version=%s)\n",
		cfg.Host, version)

	if err := server.ServeStdio(s); err != nil {
		// Detect token revocation during session: re-pair next launch.
		if errors.Is(err, auth.ErrTokenRevoked) {
			fmt.Fprintln(os.Stderr, "freebox-mcp: token révoqué — suppression du credential, re-pair au prochain démarrage")
			_ = wincred.Delete(credTarget)
		}
		fmt.Fprintf(os.Stderr, "freebox-mcp: server error: %v\n", err)
		os.Exit(1)
	}
}

// acquireToken returns a valid app token using this priority:
//  1. Windows Credential Manager (wincred) — persistent across launches
//  2. FREEBOX_APP_TOKEN env var — fallback / CI / override
//  3. Auto-pairing — first-run or after revocation
//
// After a successful auto-pair, the token is saved to wincred for future launches.
func acquireToken(cfg *config.Config, httpClient *http.Client) (string, error) {
	// 1. Credential Manager
	if tok, err := wincred.Read(credTarget); err == nil && tok != "" {
		fmt.Fprintln(os.Stderr, "freebox-mcp: token chargé depuis Credential Manager")
		return tok, nil
	}

	// 2. Env var (override / CI)
	if tok := os.Getenv("FREEBOX_APP_TOKEN"); tok != "" {
		fmt.Fprintln(os.Stderr, "freebox-mcp: token chargé depuis FREEBOX_APP_TOKEN")
		return tok, nil
	}

	// 3. Auto-pairing — first launch or after revocation
	return autoPair(cfg, httpClient)
}

// autoPair sends a pairing request and waits for the user to approve it
// in the Freebox OS web interface (Paramètres → Gestion des accès → Applications).
func autoPair(cfg *config.Config, httpClient *http.Client) (string, error) {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "freebox-mcp: aucun token trouvé — appairage automatique")
	fmt.Fprintln(os.Stderr, "freebox-mcp: envoi de la demande d'autorisation...")

	req, err := pair.Start(cfg, httpClient)
	if err != nil {
		return "", fmt.Errorf("auto-pair: %w", err)
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  ╔══════════════════════════════════════════════════════════╗")
	fmt.Fprintln(os.Stderr, "  ║  ACTION REQUISE — Freebox OS                             ║")
	fmt.Fprintln(os.Stderr, "  ║                                                          ║")
	fmt.Fprintln(os.Stderr, "  ║  Allez sur votre Freebox :                               ║")
	fmt.Fprintln(os.Stderr, "  ║  Paramètres → Gestion des accès → Applications           ║")
	fmt.Fprintln(os.Stderr, "  ║  puis acceptez la demande  « MCP Freebox »               ║")
	fmt.Fprintln(os.Stderr, "  ║                                                          ║")
	fmt.Fprintf(os.Stderr,  "  ║  Vous avez %.0f secondes.                                  ║\n", pairTimeout.Seconds())
	fmt.Fprintln(os.Stderr, "  ╚══════════════════════════════════════════════════════════╝")
	fmt.Fprintln(os.Stderr, "")

	appToken, err := pair.WaitForGrant(cfg, httpClient, req, pairTimeout)
	if err != nil {
		return "", fmt.Errorf("auto-pair: %w", err)
	}

	// Persist for future launches.
	if werr := wincred.Write(credTarget, credUser, appToken); werr != nil {
		fmt.Fprintf(os.Stderr, "freebox-mcp: avertissement — sauvegarde wincred échouée (%v)\n", werr)
	} else {
		fmt.Fprintln(os.Stderr, "freebox-mcp: appairage réussi — token sauvegardé dans Credential Manager")
	}

	return appToken, nil
}
