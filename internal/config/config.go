/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package config

import (
	"fmt"
	"os"
	"time"
)

const (
	DefaultHost    = "mafreebox.freebox.fr"
	DefaultAppID   = "mcp-freebox"
	DefaultTimeout = 10 * time.Second
)

type Config struct {
	Host    string
	AppID   string
	Timeout time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Host:    env("FREEBOX_HOST", DefaultHost),
		AppID:   env("FREEBOX_APP_ID", DefaultAppID),
		Timeout: DefaultTimeout,
	}
	if cfg.Host == "" {
		return nil, fmt.Errorf("FREEBOX_HOST must not be empty")
	}
	return cfg, nil
}

func (c *Config) BaseURL() string {
	return "https://" + c.Host + "/api/v4"
}

// DiscoveryURL returns the unauthenticated API version endpoint (HTTP, no /api/v4 prefix).
func (c *Config) DiscoveryURL() string {
	return "http://" + c.Host + "/api_version"
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
