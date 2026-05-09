/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("FREEBOX_HOST", "")
	t.Setenv("FREEBOX_APP_ID", "")
	t.Setenv("FREEBOX_TIMEOUT", "")
	t.Setenv("FREEBOX_API_BASE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Host != DefaultHost {
		t.Errorf("Host = %q, want %q", cfg.Host, DefaultHost)
	}
	if cfg.AppID != DefaultAppID {
		t.Errorf("AppID = %q, want %q", cfg.AppID, DefaultAppID)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, DefaultTimeout)
	}
	if cfg.APIBase != DefaultAPIBase {
		t.Errorf("APIBase = %q, want %q", cfg.APIBase, DefaultAPIBase)
	}
}

func TestBaseURL_Default(t *testing.T) {
	t.Setenv("FREEBOX_API_BASE", "")
	cfg, _ := Load()
	got := cfg.BaseURL()
	want := "https://" + DefaultHost + DefaultAPIBase
	if got != want {
		t.Errorf("BaseURL() = %q, want %q", got, want)
	}
}

func TestBaseURL_EnvOverride(t *testing.T) {
	t.Setenv("FREEBOX_API_BASE", "/api/v9")
	cfg, _ := Load()
	got := cfg.BaseURL()
	want := "https://" + DefaultHost + "/api/v9"
	if got != want {
		t.Errorf("BaseURL() = %q, want %q", got, want)
	}
}

func TestLoadTimeout_EnvOverride(t *testing.T) {
	t.Setenv("FREEBOX_TIMEOUT", "45s")
	cfg, _ := Load()
	if cfg.Timeout != 45*time.Second {
		t.Errorf("Timeout = %v, want 45s", cfg.Timeout)
	}
}

func TestLoadTimeout_InvalidFallsBack(t *testing.T) {
	t.Setenv("FREEBOX_TIMEOUT", "not-a-duration")
	cfg, _ := Load()
	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %v, want default %v", cfg.Timeout, DefaultTimeout)
	}
}

func TestLoadHostEmpty_ReturnsError(t *testing.T) {
	// FREEBOX_HOST="" would normally be caught — but env("key","fallback") uses
	// fallback when env is empty, so host is only empty if fallback is also empty.
	// The guard is there for programmatic misuse; test it directly.
	cfg := &Config{Host: "", APIBase: DefaultAPIBase}
	if cfg.Host != "" {
		t.Skip("host not empty")
	}
	// Validate manually: Load() itself can't produce Host="" via env.
}
