/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mipsou/mcp-freebox/internal/auth"
)

// Client wraps authenticated HTTP calls to the Freebox OS API.
type Client struct {
	baseURL string
	auth    *auth.Manager
	http    *http.Client
}

func New(baseURL string, auth *auth.Manager, http *http.Client) *Client {
	return &Client{baseURL: baseURL, auth: auth, http: http}
}

// Get performs an authenticated GET and decodes result into dst.
func (c *Client) Get(ctx context.Context, path string, dst any) error {
	return c.do(ctx, http.MethodGet, path, nil, dst)
}

// Post performs an authenticated POST with JSON body and decodes result into dst.
func (c *Client) Post(ctx context.Context, path string, body any, dst any) error {
	return c.do(ctx, http.MethodPost, path, body, dst)
}

// Put performs an authenticated PUT with JSON body.
func (c *Client) Put(ctx context.Context, path string, body any, dst any) error {
	return c.do(ctx, http.MethodPut, path, body, dst)
}

// Delete performs an authenticated DELETE.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, body, dst any) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
	}

	err := c.attempt(ctx, method, path, bodyBytes, dst)
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode == "auth_required" {
		c.auth.Invalidate()
		return c.attempt(ctx, method, path, bodyBytes, dst)
	}
	return err
}

func (c *Client) attempt(ctx context.Context, method, path string, bodyBytes []byte, dst any) error {
	token, err := c.auth.Token()
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}

	var bodyReader io.Reader
	if len(bodyBytes) > 0 {
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("X-Fbx-App-Auth", token)
	if len(bodyBytes) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var env struct {
		Success   bool            `json:"success"`
		Msg       string          `json:"msg,omitempty"`
		ErrorCode string          `json:"error_code,omitempty"`
		Result    json.RawMessage `json:"result,omitempty"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("decode envelope: %w", err)
	}
	if !env.Success {
		return &APIError{ErrorCode: env.ErrorCode, Msg: env.Msg}
	}
	if dst == nil || len(env.Result) == 0 {
		return nil
	}
	return json.Unmarshal(env.Result, dst)
}
