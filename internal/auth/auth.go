/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package auth

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // Freebox OS API mandates HMAC-SHA1
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Session holds an active session token (in-memory only, never persisted).
type Session struct {
	Token     string
	ExpiresAt time.Time
}

func (s *Session) Valid() bool {
	return s != nil && s.Token != "" && time.Now().Before(s.ExpiresAt)
}

// Manager handles Freebox OS session lifecycle.
type Manager struct {
	mu       sync.Mutex
	baseURL  string
	appID    string
	appToken string // loaded from Credential Manager, never logged
	client   *http.Client
	session  *Session
}

func New(baseURL, appID, appToken string, client *http.Client) *Manager {
	return &Manager{
		baseURL:  baseURL,
		appID:    appID,
		appToken: appToken,
		client:   client,
	}
}

// Token returns a valid session token, refreshing if needed. Safe for concurrent use.
func (m *Manager) Token() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.session.Valid() {
		return m.session.Token, nil
	}
	return m.refresh()
}

// Invalidate clears the cached session so the next Token() call forces a refresh.
// Called by the HTTP client when it receives error_code == "auth_required".
func (m *Manager) Invalidate() {
	m.mu.Lock()
	m.session = nil
	m.mu.Unlock()
}

// refresh must be called with m.mu held.
func (m *Manager) refresh() (string, error) {
	challenge, err := m.getChallenge()
	if err != nil {
		return "", fmt.Errorf("get challenge: %w", err)
	}

	password := sign(m.appToken, challenge)

	token, ttl, err := m.openSession(password)
	if err != nil {
		return "", fmt.Errorf("open session: %w", err)
	}

	m.session = &Session{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
	}
	return token, nil
}

// sign computes HMAC-SHA1(appToken, challenge) as lowercase hex.
func sign(appToken, challenge string) string {
	mac := hmac.New(sha1.New, []byte(appToken))
	mac.Write([]byte(challenge))
	return hex.EncodeToString(mac.Sum(nil))
}

type loginInfo struct {
	LoggedIn  bool   `json:"logged_in"`
	Challenge string `json:"challenge"`
}

func (m *Manager) getChallenge() (string, error) {
	resp, err := m.client.Get(m.baseURL + "/login/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var env struct {
		Success bool      `json:"success"`
		Result  loginInfo `json:"result"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return "", err
	}
	if !env.Success {
		return "", fmt.Errorf("login endpoint returned success=false")
	}
	return env.Result.Challenge, nil
}

type sessionResult struct {
	SessionToken string `json:"session_token"`
	PasswordSalt string `json:"password_salt"`
	Permissions  any    `json:"permissions"`
}

func (m *Manager) openSession(password string) (token string, ttl int, err error) {
	payload := fmt.Sprintf(`{"app_id":%q,"password":%q}`, m.appID, password)
	resp, err := m.client.Post(
		m.baseURL+"/login/session/",
		"application/json",
		strings.NewReader(payload),
	)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var env struct {
		Success bool          `json:"success"`
		Msg     string        `json:"msg"`
		Result  sessionResult `json:"result"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return "", 0, err
	}
	if !env.Success {
		return "", 0, fmt.Errorf("session error: %s", env.Msg)
	}
	// Freebox sessions last ~30 min; we use 25 min to be safe.
	return env.Result.SessionToken, 25 * 60, nil
}
