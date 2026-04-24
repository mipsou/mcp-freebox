/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

// Package mdns discovers a Freebox on the local network via multicast DNS.
// The Freebox Delta publishes a _fbx-api._tcp.local service whose TXT records
// carry the API host, HTTPS port and base URL — no hardcoded IP required.
//
// Implementation uses miekg/dns directly (already a transitive dependency)
// to avoid grandcat/zeroconf which is unmaintained since 2020.
package mdns

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (
	mdnsAddr       = "224.0.0.251:5353"
	mdnsAddrIPv6   = "[ff02::fb]:5353"
	serviceType    = "_fbx-api._tcp.local."
	DefaultTimeout = 5 * time.Second
)

// FreeboxInfo holds the discovered Freebox coordinates.
type FreeboxInfo struct {
	Host      string // IP or hostname usable as FREEBOX_HOST
	HTTPSPort int    // HTTPS port (usually 443)
	APIBase   string // e.g. "/api/v4"
}

// Discover browses for _fbx-api._tcp.local and returns the first Freebox found.
func Discover(ctx context.Context) (*FreeboxInfo, error) {
	deadline := time.Now().Add(DefaultTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("mdns: listen: %w", err)
	}
	defer conn.Close()
	conn.SetDeadline(deadline) //nolint:errcheck

	// Send PTR query for _fbx-api._tcp.local.
	// QU bit (0x8000) requests a unicast response — the Freebox replies directly
	// to our ephemeral port instead of multicasting to 224.0.0.251:5353.
	// This avoids the need to join the multicast group for receiving (which
	// Windows Firewall blocks by default on non-domain networks).
	msg := new(dns.Msg)
	msg.SetQuestion(serviceType, dns.TypePTR)
	msg.RecursionDesired = false
	msg.Question[0].Qclass = dns.ClassINET | 0x8000 // QU: unicast-response requested

	packed, err := msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("mdns: pack query: %w", err)
	}

	dst := &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	if _, err := conn.WriteTo(packed, dst); err != nil {
		return nil, fmt.Errorf("mdns: send query: %w", err)
	}

	// Listen for responses.
	buf := make([]byte, 65536)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil, fmt.Errorf("mdns: scan timeout — aucune Freebox trouvée sur le LAN")
			}
			return nil, fmt.Errorf("mdns: read: %w", err)
		}

		var resp dns.Msg
		if err := resp.Unpack(buf[:n]); err != nil {
			continue
		}

		info := parseResponse(&resp)
		if info != nil {
			return info, nil
		}
	}
}

// parseResponse extracts FreeboxInfo from an mDNS response message.
//
// Host priority: api_domain (TXT) > A/AAAA record > SRV target.
// api_domain is preferred because it carries a valid TLS certificate issued by
// Freebox CA, enabling proper HTTPS without InsecureSkipVerify in the future.
//
// The SRV port is the HTTP port (80) — it is intentionally ignored here.
// The HTTPS port comes from the TXT record "https_port" (e.g. 42460).
func parseResponse(msg *dns.Msg) *FreeboxInfo {
	info := &FreeboxInfo{HTTPSPort: 443}

	var apiDomain, aRecord, srvTarget string

	for _, rr := range append(msg.Answer, msg.Extra...) {
		switch r := rr.(type) {
		case *dns.SRV:
			// SRV target = HTTP hostname; SRV port = HTTP port (not HTTPS).
			// Kept as last-resort hostname fallback only.
			if srvTarget == "" {
				srvTarget = strings.TrimSuffix(r.Target, ".")
			}
		case *dns.A:
			if aRecord == "" {
				aRecord = r.A.String()
			}
		case *dns.AAAA:
			if aRecord == "" && !r.AAAA.IsLinkLocalUnicast() {
				aRecord = r.AAAA.String()
			}
		case *dns.TXT:
			for _, txt := range r.Txt {
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
					apiDomain = kv[1]
				}
			}
		}
	}

	// Apply host priority: api_domain > A/AAAA > SRV target.
	switch {
	case apiDomain != "":
		info.Host = apiDomain
	case aRecord != "":
		info.Host = aRecord
	case srvTarget != "":
		info.Host = srvTarget
	}

	if info.Host == "" {
		return nil
	}
	return info
}
