/*
 * Copyright (c) 2026 Mipsou <chpujol@gmail.com>
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

// Package tools — validate.go
// Single source of truth for all input validation in the tools package.
// Security is enforced here and only here; handlers call these functions
// at the entry point before any business logic.
package tools

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ── Regex patterns (also used as mcp.Pattern() schema constraints) ────────────

const (
	// MACAddrPattern is the canonical MAC address regex (aa:bb:cc:dd:ee:ff).
	MACAddrPattern = `^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`

	// IPv4Pattern matches any syntactically valid IPv4 address.
	IPv4Pattern = `^([0-9]{1,3}\.){3}[0-9]{1,3}$`

	// RFC1918Pattern matches RFC1918 private address ranges only.
	RFC1918Pattern = `^(10\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}|` +
		`172\.(1[6-9]|2[0-9]|3[01])\.[0-9]{1,3}\.[0-9]{1,3}|` +
		`192\.168\.[0-9]{1,3}\.[0-9]{1,3})$`

	// DiskNamePattern matches safe VM disk filenames (no path separators).
	DiskNamePattern = `^[a-zA-Z0-9][a-zA-Z0-9_.-]*\.(qcow2|raw)$`
)

var (
	macRegex      = regexp.MustCompile(MACAddrPattern)
	secureOnRegex = regexp.MustCompile(MACAddrPattern) // SecureOn uses same MAC format
	rfc1918Regex  = regexp.MustCompile(RFC1918Pattern)
	diskNameRegex = regexp.MustCompile(DiskNamePattern)
)

// ── MAC address ───────────────────────────────────────────────────────────────

// validateMAC rejects addresses that do not match aa:bb:cc:dd:ee:ff.
func validateMAC(mac string) error {
	if !macRegex.MatchString(mac) {
		return fmt.Errorf("adresse MAC invalide '%s' : format attendu aa:bb:cc:dd:ee:ff", mac)
	}
	return nil
}

// validateSecureOn accepts an empty password (feature disabled) or a valid
// MAC-formatted SecureOn password (6-byte hex).
func validateSecureOn(password string) error {
	if password == "" {
		return nil
	}
	if len(password) > 17 { // "xx:xx:xx:xx:xx:xx" = 17 chars max
		return fmt.Errorf("mot de passe SecureOn trop long : maximum 17 caractères (format xx:xx:xx:xx:xx:xx)")
	}
	if !secureOnRegex.MatchString(password) {
		return fmt.Errorf("mot de passe SecureOn invalide '%s' : format attendu aa:bb:cc:dd:ee:ff", password)
	}
	return nil
}

// ── Download URL (SSRF) ───────────────────────────────────────────────────────

// allowedDownloadSchemes lists URL schemes accepted by freebox_download_add.
var allowedDownloadSchemes = map[string]bool{
	"http":   true,
	"https":  true,
	"magnet": true,
	"nzb":    true,
}

// validateDownloadURL rejects unsafe URL schemes and loopback / link-local
// targets to prevent SSRF via the Freebox download engine.
func validateDownloadURL(raw string) error {
	if strings.HasPrefix(raw, "magnet:") {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("URL invalide : %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if !allowedDownloadSchemes[scheme] {
		return fmt.Errorf("schéma URL interdit '%s' : seuls http, https, magnet, nzb sont autorisés", u.Scheme)
	}
	host := strings.ToLower(u.Hostname())
	ssrfBlocked := []string{"localhost", "127.", "0.0.0.0", "169.254.", "::1", "[::1]"}
	for _, blocked := range ssrfBlocked {
		if strings.HasPrefix(host, blocked) || host == blocked {
			return fmt.Errorf("URL interdite : cible '%s' non autorisée (loopback/link-local)", host)
		}
	}
	return nil
}

// ── NAT / port forwarding ─────────────────────────────────────────────────────

// validateRFC1918 rejects any IP that is not in a private address range.
// NAT rules must target internal hosts only.
func validateRFC1918(ip string) error {
	if !rfc1918Regex.MatchString(ip) {
		return fmt.Errorf("lan_ip '%s' invalide : adresse RFC1918 requise (10.x, 172.16–31.x, 192.168.x)", ip)
	}
	return nil
}

// validatePort ensures a port number is within the valid TCP/UDP range 1–65535.
func validatePort(port int, name string) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s %d invalide : plage 1–65535 requise", name, port)
	}
	return nil
}

// ── DHCP static leases ────────────────────────────────────────────────────────

// validateDHCPIP rejects reserved last-octet values that would collide with
// the gateway (.1), the Freebox itself (.254), or network/broadcast (.0/.255).
func validateDHCPIP(ip string) error {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return fmt.Errorf("adresse IP invalide '%s'", ip)
	}
	last := parts[3]
	switch last {
	case "0":
		return fmt.Errorf("IP '%s' réservée : adresse réseau (.0 interdit)", ip)
	case "1":
		return fmt.Errorf("IP '%s' réservée : adresse gateway (.1 interdit)", ip)
	case "254":
		return fmt.Errorf("IP '%s' réservée : adresse Freebox (.254 interdit)", ip)
	case "255":
		return fmt.Errorf("IP '%s' réservée : adresse broadcast (.255 interdit)", ip)
	}
	return nil
}

// ── VM disk ───────────────────────────────────────────────────────────────────

// validateDiskName ensures the disk filename is safe: no path separators,
// no parent-directory traversal, valid extension.
// The full path (/Freebox/VMs/<name>) is constructed by the handler — callers
// never provide the directory prefix.
func validateDiskName(name string) error {
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("disk_name '%s' invalide : séparateurs de chemin '/' et '\\' interdits", name)
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("disk_name '%s' invalide : séquence '..' interdite", name)
	}
	if !diskNameRegex.MatchString(name) {
		return fmt.Errorf("disk_name '%s' invalide : format attendu nom.qcow2 ou nom.raw", name)
	}
	return nil
}
