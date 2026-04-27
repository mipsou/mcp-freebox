/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	// RFC 2202 Test Case 2: known independent reference vector.
	// Key: "Jefe", Data: "what do ya want for nothing?"
	// Expected HMAC-SHA1: effcdf6ae5eb2fa2d27416d5f184df9c259a7c79
	got := sign("Jefe", "what do ya want for nothing?")
	want := "effcdf6ae5eb2fa2d27416d5f184df9c259a7c79"
	if got != want {
		t.Errorf("sign() = %q, want %q", got, want)
	}
}

func TestSessionValid(t *testing.T) {
	s := &Session{Token: "tok", ExpiresAt: time.Now().Add(time.Minute)}
	if !s.Valid() {
		t.Error("expected valid session")
	}

	expired := &Session{Token: "tok", ExpiresAt: time.Now().Add(-time.Second)}
	if expired.Valid() {
		t.Error("expected expired session to be invalid")
	}

	var nilSession *Session
	if nilSession.Valid() {
		t.Error("nil session must be invalid")
	}
}

func TestRefresh(t *testing.T) {
	challenge := "testchallenge123"
	appToken := "testapptoken"
	sessionToken := "returned-session-token"

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v4/login/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"result":  map[string]any{"logged_in": false, "challenge": challenge},
			})
		case "/api/v4/login/session/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"result":  map[string]any{"session_token": sessionToken},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	m := New(srv.URL+"/api/v4", "test-app", appToken, srv.Client())
	tok, err := m.Token()
	if err != nil {
		t.Fatalf("Token() error: %v", err)
	}
	if tok != sessionToken {
		t.Errorf("Token() = %q, want %q", tok, sessionToken)
	}

	// Second call must reuse cached session.
	tok2, err := m.Token()
	if err != nil {
		t.Fatalf("Token() second call error: %v", err)
	}
	if tok2 != tok {
		t.Error("expected cached token to be reused")
	}
}
