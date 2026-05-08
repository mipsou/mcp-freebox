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
	"net/url"
	"os"
	"sync"
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

// errPairing is returned by lazyClient when auto-pairing is in progress.
var errPairing = errors.New(
	"appairage en cours — approuver dans Freebox OS → Gestion des accès → Applications puis réessayer",
)

// lazyClient implements tools.writer + tools.discoverer.
// When pairing is in progress, all calls return errPairing immediately.
// Once the token is obtained, setClient() installs the real client and
// all subsequent calls are forwarded normally — no server restart needed.
type lazyClient struct {
	mu   sync.RWMutex
	real *client.Client
}

func (l *lazyClient) get() (*client.Client, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.real == nil {
		return nil, errPairing
	}
	return l.real, nil
}

func (l *lazyClient) setClient(c *client.Client) {
	l.mu.Lock()
	l.real = c
	l.mu.Unlock()
}

func (l *lazyClient) Get(ctx context.Context, path string, dst any) error {
	c, err := l.get()
	if err != nil {
		return err
	}
	return c.Get(ctx, path, dst)
}

func (l *lazyClient) Post(ctx context.Context, path string, body, dst any) error {
	c, err := l.get()
	if err != nil {
		return err
	}
	return c.Post(ctx, path, body, dst)
}

func (l *lazyClient) PostForm(ctx context.Context, path string, values url.Values, dst any) error {
	c, err := l.get()
	if err != nil {
		return err
	}
	return c.PostForm(ctx, path, values, dst)
}

func (l *lazyClient) Put(ctx context.Context, path string, body, dst any) error {
	c, err := l.get()
	if err != nil {
		return err
	}
	return c.Put(ctx, path, body, dst)
}

func (l *lazyClient) Delete(ctx context.Context, path string) error {
	c, err := l.get()
	if err != nil {
		return err
	}
	return c.Delete(ctx, path)
}

func (l *lazyClient) DiscoverAPI(ctx context.Context, dst any) error {
	c, err := l.get()
	if err != nil {
		return err
	}
	return c.DiscoverAPI(ctx, dst)
}

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

	lazy := &lazyClient{}

	// Fast path: existing valid token → install client synchronously.
	if tok, ok := tryExistingToken(cfg, httpClient); ok {
		lazy.setClient(newClient(cfg, httpClient, tok))
	} else {
		// Slow path: no valid token → start server now, pair in background.
		fmt.Fprintln(os.Stderr, "freebox-mcp: aucun token valide — démarrage du serveur, appairage en arrière-plan")
		go func() {
			tok, err := autoPair(cfg, httpClient)
			if err != nil {
				fmt.Fprintf(os.Stderr, "freebox-mcp: appairage échoué: %v\n", err)
				return
			}
			lazy.setClient(newClient(cfg, httpClient, tok))
			fmt.Fprintln(os.Stderr, "freebox-mcp: appairage terminé — outils disponibles")
		}()
	}

	s := server.NewMCPServer("freebox-mcp", version,
		server.WithToolCapabilities(false),
	)
	tools.RegisterAll(s, lazy, lazy)

	fmt.Fprintf(os.Stderr, "freebox-mcp: starting (host=%s, version=%s)\n", cfg.Host, version)

	if err := server.ServeStdio(s); err != nil {
		if errors.Is(err, auth.ErrTokenRevoked) {
			fmt.Fprintln(os.Stderr, "freebox-mcp: token révoqué — suppression du credential, re-pair au prochain démarrage")
			_ = wincred.Delete(credTarget)
		}
		fmt.Fprintf(os.Stderr, "freebox-mcp: server error: %v\n", err)
		os.Exit(1)
	}
}

// newClient constructs a fully initialized API client from a valid token.
func newClient(cfg *config.Config, httpClient *http.Client, tok string) *client.Client {
	mgr := auth.New(cfg.BaseURL(), cfg.AppID, tok, httpClient)
	return client.New(cfg.BaseURL(), cfg.DiscoveryURL(), mgr, httpClient)
}

// tryExistingToken attempts to return a valid token from wincred or env var.
// Returns ("", false) if no valid token is available (absent, revoked, network error on first try).
func tryExistingToken(cfg *config.Config, httpClient *http.Client) (string, bool) {
	// 1. Credential Manager — with eager validation.
	if tok, err := wincred.Read(credTarget); err == nil && tok != "" {
		if verr := validateToken(cfg, httpClient, tok); verr == nil {
			fmt.Fprintln(os.Stderr, "freebox-mcp: token valide — démarrage normal")
			return tok, true
		} else if errors.Is(verr, auth.ErrTokenRevoked) {
			fmt.Fprintln(os.Stderr, "freebox-mcp: token révoqué côté Freebox — suppression et re-pair en arrière-plan")
			_ = wincred.Delete(credTarget)
		} else {
			// Network error — trust the token, let session retry later.
			fmt.Fprintf(os.Stderr, "freebox-mcp: validation impossible (%v) — démarrage avec token existant\n", verr)
			return tok, true
		}
	}

	// 2. Env var (override / CI) — no validation, trusted source.
	if tok := os.Getenv("FREEBOX_APP_TOKEN"); tok != "" {
		fmt.Fprintln(os.Stderr, "freebox-mcp: token chargé depuis FREEBOX_APP_TOKEN")
		return tok, true
	}

	return "", false
}

// validateToken does a lightweight auth check to confirm the token is still accepted.
func validateToken(cfg *config.Config, httpClient *http.Client, tok string) error {
	mgr := auth.New(cfg.BaseURL(), cfg.AppID, tok, httpClient)
	_, err := mgr.Token()
	return err
}

// autoPair sends a pairing request and waits for the user to approve it.
// Saves the token to wincred on success.
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
	fmt.Fprintf(os.Stderr, "  ║  Vous avez %.0f secondes.                                  ║\n", pairTimeout.Seconds())
	fmt.Fprintln(os.Stderr, "  ╚══════════════════════════════════════════════════════════╝")
	fmt.Fprintln(os.Stderr, "")

	appToken, err := pair.WaitForGrant(cfg, httpClient, req, pairTimeout)
	if err != nil {
		return "", fmt.Errorf("auto-pair: %w", err)
	}

	if werr := wincred.Write(credTarget, credUser, appToken); werr != nil {
		fmt.Fprintf(os.Stderr, "freebox-mcp: avertissement — sauvegarde wincred échouée (%v)\n", werr)
	} else {
		fmt.Fprintln(os.Stderr, "freebox-mcp: token sauvegardé dans Credential Manager")
	}

	return appToken, nil
}
