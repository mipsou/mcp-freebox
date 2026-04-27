/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mipsou/mcp-freebox/internal/auth"
)

// stubAuth returns a fixed token without hitting any server.
type stubAuth string

// newTestClient builds a Client backed by a test server that always
// returns the provided payload as {"success":true,"result":<payload>}.
func newTestClient(t *testing.T, path string, payload any) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/login/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"result":  map[string]any{"logged_in": false, "challenge": "ch"},
			})
		case "/api/v4/login/session/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"result":  map[string]any{"session_token": "test-tok"},
			})
		case path:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"result":  payload,
			})
		default:
			http.NotFound(w, r)
		}
	}))

	mgr := auth.New(srv.URL+"/api/v4", "test-app", "test-token", srv.Client())
	c := New(srv.URL+"/api/v4", srv.URL+"/api_version", mgr, srv.Client())
	return c, srv
}

func TestGet(t *testing.T) {
	type thing struct{ Name string }
	c, srv := newTestClient(t, "/api/v4/thing/", thing{Name: "hello"})
	defer srv.Close()

	var got thing
	if err := c.Get(context.Background(), "/thing/", &got); err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got.Name != "hello" {
		t.Errorf("Name = %q, want hello", got.Name)
	}
}

func TestAPIError(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/login/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"result":  map[string]any{"logged_in": false, "challenge": "ch"},
			})
		case "/api/v4/login/session/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"result":  map[string]any{"session_token": "test-tok"},
			})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success":    false,
				"msg":        "permission denied",
				"error_code": "insufficient_rights",
			})
		}
	}))
	defer srv.Close()

	mgr := auth.New(srv.URL+"/api/v4", "test-app", "test-token", srv.Client())
	c := New(srv.URL+"/api/v4", srv.URL+"/api_version", mgr, srv.Client())

	var dst any
	err := c.Get(context.Background(), "/protected/", &dst)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.ErrorCode != "insufficient_rights" {
		t.Errorf("ErrorCode = %q, want insufficient_rights", apiErr.ErrorCode)
	}
}
