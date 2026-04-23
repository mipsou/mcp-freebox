/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package client

import "fmt"

// APIError represents a Freebox OS API error response.
type APIError struct {
	ErrorCode string `json:"error_code"`
	Msg       string `json:"msg"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("freebox api error %s: %s", e.ErrorCode, e.Msg)
}

// apiResponse is the generic Freebox OS API envelope.
type apiResponse[T any] struct {
	Success   bool      `json:"success"`
	Msg       string    `json:"msg,omitempty"`
	ErrorCode string    `json:"error_code,omitempty"`
	Result    T         `json:"result,omitempty"`
}
