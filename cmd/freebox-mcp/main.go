/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"

	"github.com/mipsou/mcp-freebox/internal/auth"
	"github.com/mipsou/mcp-freebox/internal/client"
	"github.com/mipsou/mcp-freebox/internal/config"
	"github.com/mipsou/mcp-freebox/internal/tools"
)

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-mcp: config error: %v\n", err)
		os.Exit(1)
	}

	appToken, err := loadAppToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-mcp: token error: %v\n", err)
		fmt.Fprintf(os.Stderr, "freebox-mcp: run freebox-pair first to register this app\n")
		os.Exit(1)
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

	mgr := auth.New(cfg.BaseURL(), cfg.AppID, appToken, httpClient)
	c := client.New(cfg.BaseURL(), cfg.DiscoveryURL(), mgr, httpClient)

	s := server.NewMCPServer("freebox-mcp", version,
		server.WithToolCapabilities(false),
	)
	tools.RegisterAll(s, c, c)

	fmt.Fprintf(os.Stderr, "freebox-mcp: starting (host=%s, version=%s)\n",
		cfg.Host, version)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "freebox-mcp: server error: %v\n", err)
		os.Exit(1)
	}
}

// loadAppToken reads the app token from the FREEBOX_APP_TOKEN env var.
// Production: replace with Windows Credential Manager (wincred) lookup.
func loadAppToken() (string, error) {
	tok := os.Getenv("FREEBOX_APP_TOKEN")
	if tok == "" {
		return "", fmt.Errorf("FREEBOX_APP_TOKEN not set")
	}
	return tok, nil
}
