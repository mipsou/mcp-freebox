/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

// Package mdns discovers a Freebox on the local network via multicast DNS.
// The Freebox Delta publishes a _fbx-api._tcp.local service whose TXT records
// carry the API host, HTTPS port and base URL — no hardcoded IP required.
package mdns

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	serviceType = "_fbx-api._tcp"
	domain      = "local"
	// DefaultTimeout is how long we listen for mDNS responses before giving up.
	DefaultTimeout = 5 * time.Second
)

// FreeboxInfo holds the discovered Freebox coordinates.
type FreeboxInfo struct {
	Host      string // hostname or IP usable as FREEBOX_HOST
	HTTPSPort int    // HTTPS port (usually 443 or custom)
	APIBase   string // e.g. "/api/v4"
}

// Discover browses for _fbx-api._tcp.local and returns the first Freebox found.
// ctx controls the overall deadline; internally a 5-second scan window is used.
func Discover(ctx context.Context) (*FreeboxInfo, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, fmt.Errorf("mdns: resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry, 4)

	scanCtx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	if err := resolver.Browse(scanCtx, serviceType, domain, entries); err != nil {
		return nil, fmt.Errorf("mdns: browse: %w", err)
	}

	for {
		select {
		case entry, ok := <-entries:
			if !ok {
				return nil, fmt.Errorf("mdns: no Freebox found on the network")
			}
			info := parseEntry(entry)
			if info != nil {
				return info, nil
			}
		case <-scanCtx.Done():
			return nil, fmt.Errorf("mdns: scan timeout — no Freebox found (is it on the same LAN?)")
		}
	}
}

// parseEntry extracts FreeboxInfo from a zeroconf ServiceEntry.
func parseEntry(e *zeroconf.ServiceEntry) *FreeboxInfo {
	if e == nil {
		return nil
	}

	info := &FreeboxInfo{
		HTTPSPort: e.Port,
	}

	// Prefer AddrIPv4, fall back to hostname
	if len(e.AddrIPv4) > 0 {
		info.Host = e.AddrIPv4[0].String()
	} else if len(e.AddrIPv6) > 0 {
		// Wrap IPv6 in brackets for URL construction
		addr := e.AddrIPv6[0]
		if isLinkLocal(addr) && len(e.AddrIPv6) > 1 {
			addr = e.AddrIPv6[1]
		}
		info.Host = "[" + addr.String() + "]"
	} else if e.HostName != "" {
		info.Host = strings.TrimSuffix(e.HostName, ".")
	}

	// Parse TXT records: api_base_url, https_port, api_domain
	for _, txt := range e.Text {
		kv := strings.SplitN(txt, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch strings.ToLower(kv[0]) {
		case "api_base_url":
			info.APIBase = kv[1]
		case "https_port":
			if p, err := strconv.Atoi(kv[1]); err == nil {
				info.HTTPSPort = p
			}
		case "api_domain":
			if info.Host == "" {
				info.Host = kv[1]
			}
		}
	}

	if info.Host == "" {
		return nil
	}
	return info
}

// isLinkLocal reports whether addr is an IPv6 link-local address (fe80::/10).
func isLinkLocal(addr net.IP) bool {
	return addr != nil && addr.IsLinkLocalUnicast()
}
