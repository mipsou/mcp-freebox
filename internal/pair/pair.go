/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

// Package pair handles Freebox OS application authorization (pairing).
// It provides the API calls to request and poll for an app token.
// Interactive prompts live in cmd/freebox-pair; this package is UI-agnostic.
package pair

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mipsou/mcp-freebox/internal/config"
)

// Request holds the pending authorization state returned by Start.
type Request struct {
	AppToken string // app_token to save on success
	TrackID  int    // track_id for polling
}

// Start sends an authorization request to the Freebox.
// The Freebox OS will display a pending notification (LED blink or web UI).
// The caller must then call WaitForGrant and direct the user to approve.
func Start(cfg *config.Config, c *http.Client) (*Request, error) {
	hostname, _ := os.Hostname()
	body, _ := json.Marshal(map[string]string{
		"app_id":      cfg.AppID,
		"app_name":    "MCP Freebox",
		"app_version": "1.0",
		"device_name": hostname,
	})

	resp, err := c.Post(cfg.BaseURL()+"/login/authorize/",
		"application/json", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("pair start: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var env struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		Result  struct {
			AppToken string `json:"app_token"`
			TrackID  int    `json:"track_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("pair start parse: %w", err)
	}
	if !env.Success {
		return nil, fmt.Errorf("pair start: %s", env.Msg)
	}
	return &Request{AppToken: env.Result.AppToken, TrackID: env.Result.TrackID}, nil
}

// WaitForGrant polls the Freebox until the user grants or denies access,
// or until timeout is reached. Returns the app_token on success.
func WaitForGrant(cfg *config.Config, c *http.Client, req *Request, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := c.Get(fmt.Sprintf("%s/login/authorize/%d", cfg.BaseURL(), req.TrackID))
		if err != nil {
			return "", fmt.Errorf("pair poll: %w", err)
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
			return "", fmt.Errorf("pair poll parse: %w", err)
		}
		if !env.Success {
			return "", fmt.Errorf("pair poll: %s", env.Msg)
		}

		switch env.Result.Status {
		case "granted":
			return req.AppToken, nil
		case "denied":
			return "", fmt.Errorf("pair: accès refusé par l'utilisateur")
		case "timeout":
			return "", fmt.Errorf("pair: délai d'autorisation dépassé côté Freebox")
		}
		time.Sleep(2 * time.Second)
	}
	return "", fmt.Errorf("pair: timeout client (%.0fs sans réponse)", timeout.Seconds())
}
