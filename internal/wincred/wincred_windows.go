/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

// Package wincred provides read/write access to Windows Credential Manager
// using native advapi32.dll syscalls — no external module required.
package wincred

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

const (
	credTypeGeneric         = 1
	credPersistLocalMachine = 2
)

// credentialW mirrors the Windows CREDENTIALW structure.
type credentialW struct {
	Flags              uint32
	Type               uint32
	TargetName         *uint16
	Comment            *uint16
	LastWritten        syscall.Filetime
	CredentialBlobSize uint32
	CredentialBlob     uintptr
	Persist            uint32
	AttributeCount     uint32
	Attributes         uintptr
	TargetAlias        *uint16
	UserName           *uint16
}

var (
	advapi32   = syscall.NewLazyDLL("advapi32.dll")
	credReadW  = advapi32.NewProc("CredReadW")
	credWriteW = advapi32.NewProc("CredWriteW")
	credFree   = advapi32.NewProc("CredFree")
	credDeleteW = advapi32.NewProc("CredDeleteW")
)

// Read returns the password stored for the given generic credential target.
// Returns ("", ErrNotFound) if the credential does not exist.
func Read(target string) (string, error) {
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return "", fmt.Errorf("wincred read: %w", err)
	}

	var ptr uintptr
	ret, _, e := credReadW.Call(
		uintptr(unsafe.Pointer(targetPtr)),
		credTypeGeneric,
		0,
		uintptr(unsafe.Pointer(&ptr)),
	)
	if ret == 0 {
		if errno, ok := e.(syscall.Errno); ok && errno == 1168 { // ERROR_NOT_FOUND
			return "", ErrNotFound
		}
		return "", fmt.Errorf("wincred read CredReadW: %w", e)
	}
	defer credFree.Call(ptr) //nolint:errcheck

	c := (*credentialW)(unsafe.Pointer(ptr))
	if c.CredentialBlobSize == 0 {
		return "", nil
	}

	// CredentialBlob is UTF-16LE encoded bytes (no null terminator guaranteed).
	blobSize := int(c.CredentialBlobSize)
	blob := make([]byte, blobSize)
	for i := range blob {
		blob[i] = *(*byte)(unsafe.Pointer(c.CredentialBlob + uintptr(i)))
	}

	u16s := make([]uint16, blobSize/2)
	for i := range u16s {
		u16s[i] = binary.LittleEndian.Uint16(blob[i*2:])
	}
	return string(utf16.Decode(u16s)), nil
}

// Write stores a generic credential (UTF-16LE blob) in Windows Credential Manager.
// The credential persists across user sessions (DPAPI-protected).
func Write(target, username, password string) error {
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return fmt.Errorf("wincred write target: %w", err)
	}
	usernamePtr, err := syscall.UTF16PtrFromString(username)
	if err != nil {
		return fmt.Errorf("wincred write username: %w", err)
	}

	// Encode password as UTF-16LE without null terminator.
	u16s := utf16.Encode([]rune(password))
	blob := make([]byte, len(u16s)*2)
	for i, v := range u16s {
		binary.LittleEndian.PutUint16(blob[i*2:], v)
	}

	var blobPtr uintptr
	if len(blob) > 0 {
		blobPtr = uintptr(unsafe.Pointer(&blob[0]))
	}

	cred := credentialW{
		Type:               credTypeGeneric,
		TargetName:         targetPtr,
		UserName:           usernamePtr,
		CredentialBlobSize: uint32(len(blob)),
		CredentialBlob:     blobPtr,
		Persist:            credPersistLocalMachine,
	}

	ret, _, e := credWriteW.Call(uintptr(unsafe.Pointer(&cred)), 0)
	if ret == 0 {
		return fmt.Errorf("wincred write CredWriteW: %w", e)
	}
	return nil
}

// Delete removes a generic credential from Windows Credential Manager.
// Returns nil if the credential did not exist.
func Delete(target string) error {
	targetPtr, err := syscall.UTF16PtrFromString(target)
	if err != nil {
		return fmt.Errorf("wincred delete: %w", err)
	}
	ret, _, e := credDeleteW.Call(
		uintptr(unsafe.Pointer(targetPtr)),
		credTypeGeneric,
		0,
	)
	if ret == 0 {
		if errno, ok := e.(syscall.Errno); ok && errno == 1168 {
			return nil // already gone
		}
		return fmt.Errorf("wincred delete CredDeleteW: %w", e)
	}
	return nil
}

// ErrNotFound is returned by Read when the credential does not exist.
var ErrNotFound = fmt.Errorf("wincred: credential not found")
