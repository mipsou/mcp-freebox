/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

// freebox-pair is a one-shot CLI that pairs this application with the Freebox.
// Run once: press the physical button on the Freebox when prompted,
// then save the printed app_token to your secret store (Credential Manager).
package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mipsou/mcp-freebox/internal/config"
	"github.com/mipsou/mcp-freebox/internal/mdns"
)

// permission describes one Freebox OS permission scope.
type permission struct {
	name      string
	access    string // "Lecture" or "Lecture/Écriture" or "Contrôle"
	usedByMCP bool
	desc      string
}

// freeboxPermissions lists all permissions the Freebox may grant.
// usedByMCP = true si un outil MCP existant l'utilise.
var freeboxPermissions = []permission{
	{"Connexion (WAN, débits, xDSL, FTTH)", "Lecture", true, "État ligne, IPs publiques, DynDNS"},
	{"Modification des réglages de la Freebox", "Lecture/Écriture", true, "Système (uptime, firmware, temp), switch LAN"},
	{"Accès aux fichiers de la Freebox", "Lecture/Écriture", false, "NAS, stockage interne"},
	{"Accès à la base de contacts", "Lecture/Écriture", false, "Répertoire téléphonique"},
	{"Accès au journal d'appels", "Lecture", false, "Historique des appels"},
	{"Accès au guide TV", "Lecture", false, "Programme télévisé"},
	{"Programmation des enregistrements", "Lecture/Écriture", false, "Enregistrements TV"},
	{"Contrôle de la VM", "Contrôle", true, "Démarrage/arrêt machines virtuelles (PRA)"},
	{"Accès au gestionnaire de téléchargements", "Lecture/Écriture", false, "Ajout/suppression de téléchargements"},
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-pair: config error: %v\n", err)
		os.Exit(1)
	}

	// Si FREEBOX_HOST n'est pas défini, tenter la découverte mDNS.
	// En cas d'échec : fallback sur mafreebox.freebox.fr (défaut config).
	if os.Getenv("FREEBOX_HOST") == "" {
		fmt.Fprintln(os.Stderr, "freebox-pair: découverte mDNS en cours...")
		discovered, err := mdns.Discover(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "freebox-pair: mDNS indisponible (%v) — fallback %s\n", err, cfg.Host)
		} else {
			cfg.Host = discovered.Host
			if discovered.HTTPSPort != 0 && discovered.HTTPSPort != 443 {
				cfg.Host = fmt.Sprintf("%s:%d", discovered.Host, discovered.HTTPSPort)
			}
			fmt.Fprintf(os.Stderr, "freebox-pair: Freebox trouvée → %s\n", cfg.Host)
		}
	}

	printPermissions()

	if !confirm("Continuer avec ce pairing ?") {
		fmt.Fprintln(os.Stderr, "freebox-pair: annulé.")
		os.Exit(0)
	}

	//nolint:gosec
	httpClient := &http.Client{
		Timeout: 90 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	fmt.Fprintln(os.Stderr, "\nfreebox-pair: envoi de la demande d'autorisation...")
	token, err := requestToken(cfg, httpClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-pair: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "\n*** La Freebox attend votre validation physique ***")
	fmt.Fprintln(os.Stderr, "Le voyant de la Freebox clignote.")
	if !confirm("Prêt à appuyer sur le bouton de la Freebox ?") {
		fmt.Fprintln(os.Stderr, "freebox-pair: annulé. Aucun token n'a été enregistré.")
		os.Exit(0)
	}

	fmt.Fprintln(os.Stderr, "freebox-pair: en attente de la validation (60s)...")
	_, appToken, err := waitForGrant(cfg, httpClient, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-pair: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "\nfreebox-pair: SUCCESS — token ci-dessous (stdout)")

	// app_token sur stdout pour piping vers un secret store.
	fmt.Println(appToken)
}

func printPermissions() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "  PERMISSIONS QUE LA FREEBOX VA ACCORDER À CETTE APPLICATION")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "")

	fmt.Fprintln(os.Stderr, "  NÉCESSAIRES pour freebox-mcp v0.4 :")
	for _, p := range freeboxPermissions {
		if p.usedByMCP {
			fmt.Fprintf(os.Stderr, "    [✓] %-42s  %-22s  %s\n", p.name, p.access, p.desc)
		}
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  ACCORDÉES PAR DÉFAUT (non nécessaires pour v0.4) :")
	for _, p := range freeboxPermissions {
		if !p.usedByMCP {
			fmt.Fprintf(os.Stderr, "    [ ] %-42s  %-22s  %s\n", p.name, p.access, p.desc)
		}
	}
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "─────────────────────────────────────────────────────────")
	fmt.Fprintln(os.Stderr, "  ACTION REQUISE APRÈS LE PAIRING :")
	fmt.Fprintln(os.Stderr, "  La Freebox accorde TOUS ces droits par défaut.")
	fmt.Fprintln(os.Stderr, "  Vous devrez décocher les droits non marqués [✓] dans :")
	fmt.Fprintln(os.Stderr, "  Freebox OS → Paramètres → Gestion des accès → Applications")
	fmt.Fprintln(os.Stderr, "─────────────────────────────────────────────────────────")
	fmt.Fprintln(os.Stderr, "")
}

var stdinScanner = bufio.NewScanner(os.Stdin)

func confirm(prompt string) bool {
	fmt.Fprintf(os.Stderr, "  → %s [y/N] : ", prompt)
	if !stdinScanner.Scan() {
		return false
	}
	return strings.ToLower(strings.TrimSpace(stdinScanner.Text())) == "y"
}

type authorizeRequest struct {
	AppID      string `json:"app_id"`
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
	DeviceName string `json:"device_name"`
}

type authorizeResult struct {
	AppToken string `json:"app_token"`
	TrackID  int    `json:"track_id"`
}

func requestToken(cfg *config.Config, c *http.Client) (*authorizeResult, error) {
	hostname, _ := os.Hostname()
	body, _ := json.Marshal(authorizeRequest{
		AppID:      cfg.AppID,
		AppName:    "MCP Freebox",
		AppVersion: "1.0",
		DeviceName: hostname,
	})

	resp, err := c.Post(cfg.BaseURL()+"/login/authorize/",
		"application/json", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var env struct {
		Success bool            `json:"success"`
		Msg     string          `json:"msg"`
		Result  authorizeResult `json:"result"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, err
	}
	if !env.Success {
		return nil, fmt.Errorf("authorize failed: %s", env.Msg)
	}
	return &env.Result, nil
}

func waitForGrant(cfg *config.Config, c *http.Client, ar *authorizeResult) (int, string, error) {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := c.Get(fmt.Sprintf("%s/login/authorize/%d", cfg.BaseURL(), ar.TrackID))
		if err != nil {
			return 0, "", err
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var env struct {
			Success bool   `json:"success"`
			Msg     string `json:"msg"`
			Result  struct {
				Status string `json:"status"`
			} `json:"result"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			return 0, "", err
		}
		if !env.Success {
			return 0, "", fmt.Errorf("poll authorize: %s", env.Msg)
		}

		switch env.Result.Status {
		case "granted":
			return ar.TrackID, ar.AppToken, nil
		case "denied":
			return 0, "", fmt.Errorf("accès refusé par l'utilisateur")
		case "timeout":
			return 0, "", fmt.Errorf("délai d'autorisation dépassé")
		}
		time.Sleep(2 * time.Second)
	}
	return 0, "", fmt.Errorf("timeout client : bouton non pressé dans les 60s")
}
