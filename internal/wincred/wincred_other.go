// Copyright (c) 2026 Mipsou <chpujol@gmail.com>
//
// SPDX-License-Identifier: EUPL-1.2

//go:build !windows

// Package wincred provides read/write access to a system credential store.
// On non-Windows platforms all operations are no-ops — tokens must be supplied
// via the FREEBOX_TOKEN environment variable instead.
package wincred

import "errors"

var errNotSupported = errors.New("wincred: credential store not available on this platform (use FREEBOX_TOKEN env var)")

// Read always returns an error on non-Windows platforms.
func Read(_ string) (string, error) {
	return "", errNotSupported
}

// Write is a no-op on non-Windows platforms.
func Write(_, _, _ string) error {
	return nil
}

// Delete is a no-op on non-Windows platforms.
func Delete(_ string) error {
	return nil
}
