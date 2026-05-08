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
	DefaultTimeout = 30 * time.Second // TLS cold-start to mafreebox.freebox.fr takes ~7.5 s
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
		Timeout: envDuration("FREEBOX_TIMEOUT", DefaultTimeout),
	}
	if cfg.Host == "" {
		return nil, fmt.Errorf("FREEBOX_HOST must not be empty")
	}
	return cfg, nil
}

func (c *Config) BaseURL() string {
	return "https://" + c.Host + "/api/v9"
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

// envDuration reads a Go duration string from an environment variable.
// Falls back to the default value if the variable is absent or unparseable.
func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}
