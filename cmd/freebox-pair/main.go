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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mipsou/mcp-freebox/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-pair: config error: %v\n", err)
		os.Exit(1)
	}

	//nolint:gosec
	httpClient := &http.Client{
		Timeout: 90 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	fmt.Fprintln(os.Stderr, "freebox-pair: requesting app token...")
	token, err := requestToken(cfg, httpClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-pair: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "freebox-pair: *** PRESS THE BUTTON ON YOUR FREEBOX NOW ***")
	fmt.Fprintln(os.Stderr, "freebox-pair: waiting for authorization (60s timeout)...")

	trackID, appToken, err := waitForGrant(cfg, httpClient, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "freebox-pair: %v\n", err)
		os.Exit(1)
	}
	_ = trackID

	fmt.Fprintln(os.Stderr, "freebox-pair: SUCCESS")
	// Print app_token to stdout for piping into a secret store.
	fmt.Println(appToken)
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
			return 0, "", fmt.Errorf("access denied by user")
		case "timeout":
			return 0, "", fmt.Errorf("authorization timed out")
		}
		time.Sleep(2 * time.Second)
	}
	return 0, "", fmt.Errorf("client-side timeout waiting for button press")
}
