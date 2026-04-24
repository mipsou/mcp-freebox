//go:build ignore

/*
 * mdns-debug — outil de diagnostic mDNS pour freebox-mcp
 *
 * Teste les 3 mécanismes de découverte Freebox dans l'ordre :
 *   1. Écoute passive multicast (224.0.0.251:5353) — reçoit-on du trafic _fbx-api._tcp ?
 *   2. Query active QU (unicast-response) — la Freebox répond-elle en unicast ?
 *   3. Query active multicast classique — reçoit-on la réponse via le groupe multicast ?
 *   4. Résolution mafreebox.freebox.fr — fallback DNS local Freebox
 *
 * Usage : go run cmd/mdns-debug/main.go
 */

package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (
	mdnsGroup   = "224.0.0.251"
	mdnsPort    = 5353
	serviceType = "_fbx-api._tcp.local."
	timeout     = 5 * time.Second
)

func main() {
	fmt.Println("=== mdns-debug : diagnostic découverte Freebox ===")
	fmt.Println()

	// ── Test 1 : Écoute passive sur le groupe multicast ──────────────────────
	fmt.Println("── TEST 1 : Écoute passive multicast 224.0.0.251:5353 (5s) ──")
	fmt.Println("  → Rejoint le groupe multicast et écoute sans envoyer de query")
	passiveListen()
	fmt.Println()

	// ── Test 2 : Query active avec bit QU (unicast-response) ─────────────────
	fmt.Println("── TEST 2 : Query active QU (réponse unicast attendue) ──────")
	fmt.Println("  → Envoie PTR query avec bit QU=1, écoute unicast")
	sendQuery(true)
	fmt.Println()

	// ── Test 3 : Query active multicast classique ─────────────────────────────
	fmt.Println("── TEST 3 : Query active multicast classique (QU=0) ─────────")
	fmt.Println("  → Envoie PTR query avec QU=0, écoute port 5353 multicast")
	sendQueryMulticast()
	fmt.Println()

	// ── Test 4 : Résolution DNS mafreebox.freebox.fr ─────────────────────────
	fmt.Println("── TEST 4 : Résolution mafreebox.freebox.fr ─────────────────")
	testDNSFallback()
	fmt.Println()

	fmt.Println("=== Fin du diagnostic ===")
}

// passiveListen rejoint le groupe multicast 224.0.0.251:5353 et écoute
// passivement le trafic mDNS pendant timeout.
func passiveListen() {
	group := &net.UDPAddr{IP: net.ParseIP(mdnsGroup), Port: mdnsPort}
	conn, err := net.ListenMulticastUDP("udp4", nil, group)
	if err != nil {
		fmt.Printf("  ERREUR ListenMulticastUDP: %v\n", err)
		fmt.Println("  → Windows: le pare-feu bloque peut-être le port 5353")
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(timeout)) //nolint:errcheck

	fmt.Println("  Socket multicast ouvert — écoute en cours...")
	received := 0
	buf := make([]byte, 65536)
	for {
		n, src, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			fmt.Printf("  ERREUR lecture: %v\n", err)
			break
		}
		received++
		var msg dns.Msg
		if parseErr := msg.Unpack(buf[:n]); parseErr != nil {
			fmt.Printf("  Paquet #%d depuis %s — parse échoué: %v\n", received, src, parseErr)
			continue
		}
		fmt.Printf("  Paquet #%d depuis %s — %d questions, %d réponses\n",
			received, src, len(msg.Question), len(msg.Answer))
		for _, q := range msg.Question {
			fmt.Printf("    Question: %s\n", q.String())
		}
		for _, rr := range append(msg.Answer, msg.Extra...) {
			s := rr.String()
			if strings.Contains(s, "fbx") || strings.Contains(s, "freebox") ||
				strings.Contains(s, "Freebox") || strings.Contains(s, "_fbx") {
				fmt.Printf("    [FREEBOX] %s\n", s)
			} else {
				fmt.Printf("    %s\n", s)
			}
		}
	}
	if received == 0 {
		fmt.Println("  RÉSULTAT: aucun paquet multicast reçu en 5s")
		fmt.Println("  → Cause probable: pare-feu Windows bloque 224.0.0.251:5353")
		fmt.Println("  → Ou: aucun trafic mDNS sur ce LAN pendant la fenêtre d'écoute")
	} else {
		fmt.Printf("  RÉSULTAT: %d paquet(s) reçus\n", received)
	}
}

// sendQuery envoie une PTR query avec le bit QU optionnel et écoute sur
// un port éphémère (réponse unicast si QU=1).
func sendQuery(quBit bool) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		fmt.Printf("  ERREUR listen: %v\n", err)
		return
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(timeout)) //nolint:errcheck

	msg := new(dns.Msg)
	msg.SetQuestion(serviceType, dns.TypePTR)
	msg.RecursionDesired = false
	qclass := uint16(dns.ClassINET)
	if quBit {
		qclass |= 0x8000
	}
	msg.Question[0].Qclass = qclass

	packed, err := msg.Pack()
	if err != nil {
		fmt.Printf("  ERREUR pack: %v\n", err)
		return
	}

	dst := &net.UDPAddr{IP: net.ParseIP(mdnsGroup), Port: mdnsPort}
	if _, err := conn.WriteTo(packed, dst); err != nil {
		fmt.Printf("  ERREUR envoi: %v\n", err)
		return
	}
	localAddr := conn.LocalAddr()
	fmt.Printf("  Query envoyée depuis %s vers %s (QU=%v)\n", localAddr, dst, quBit)

	buf := make([]byte, 65536)
	received := 0
	for {
		n, src, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			break
		}
		received++
		var resp dns.Msg
		if parseErr := resp.Unpack(buf[:n]); parseErr != nil {
			fmt.Printf("  Réponse #%d depuis %s — parse échoué\n", received, src)
			continue
		}
		fmt.Printf("  Réponse #%d depuis %s — %d answer(s)\n", received, src, len(resp.Answer))
		for _, rr := range append(resp.Answer, resp.Extra...) {
			fmt.Printf("    %s\n", rr.String())
		}
	}
	if received == 0 {
		fmt.Printf("  RÉSULTAT: aucune réponse en 5s (QU=%v)\n", quBit)
	} else {
		fmt.Printf("  RÉSULTAT: %d réponse(s) reçues\n", received)
	}
}

// sendQueryMulticast envoie une query classique (QU=0) et écoute sur
// le groupe multicast pour la réponse.
func sendQueryMulticast() {
	// Ouvre un socket multicast pour recevoir
	group := &net.UDPAddr{IP: net.ParseIP(mdnsGroup), Port: mdnsPort}
	recv, err := net.ListenMulticastUDP("udp4", nil, group)
	if err != nil {
		fmt.Printf("  ERREUR socket multicast: %v\n", err)
		return
	}
	defer recv.Close()
	recv.SetDeadline(time.Now().Add(timeout)) //nolint:errcheck

	// Envoie la query depuis un port éphémère
	send, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		fmt.Printf("  ERREUR socket envoi: %v\n", err)
		return
	}
	defer send.Close()

	msg := new(dns.Msg)
	msg.SetQuestion(serviceType, dns.TypePTR)
	msg.RecursionDesired = false
	msg.Question[0].Qclass = dns.ClassINET // QU=0

	packed, _ := msg.Pack()
	dst := &net.UDPAddr{IP: net.ParseIP(mdnsGroup), Port: mdnsPort}
	if _, err := send.WriteTo(packed, dst); err != nil {
		fmt.Printf("  ERREUR envoi: %v\n", err)
		return
	}
	fmt.Printf("  Query multicast envoyée, écoute sur %s:%d\n", mdnsGroup, mdnsPort)

	buf := make([]byte, 65536)
	received := 0
	for {
		n, src, err := recv.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			break
		}
		var resp dns.Msg
		if parseErr := resp.Unpack(buf[:n]); parseErr != nil {
			continue
		}
		// Filtre notre propre query
		if len(resp.Answer) == 0 && len(resp.Extra) == 0 {
			continue
		}
		received++
		fmt.Printf("  Réponse #%d depuis %s — %d answer(s)\n", received, src, len(resp.Answer))
		for _, rr := range append(resp.Answer, resp.Extra...) {
			fmt.Printf("    %s\n", rr.String())
		}
	}
	if received == 0 {
		fmt.Println("  RÉSULTAT: aucune réponse multicast en 5s")
	} else {
		fmt.Printf("  RÉSULTAT: %d réponse(s) reçues\n", received)
	}
}

// testDNSFallback vérifie que mafreebox.freebox.fr est résolvable.
func testDNSFallback() {
	host := "mafreebox.freebox.fr"
	addrs, err := net.LookupHost(host)
	if err != nil {
		fmt.Printf("  ERREUR résolution %s: %v\n", host, err)
		fmt.Println("  RÉSULTAT: fallback DNS NON disponible")
		return
	}
	fmt.Printf("  %s → %v\n", host, addrs)

	// Test HTTPS connectivity
	conn, err := net.DialTimeout("tcp", addrs[0]+":443", 3*time.Second)
	if err != nil {
		fmt.Printf("  Port 443 inaccessible: %v\n", err)
	} else {
		conn.Close()
		fmt.Printf("  Port 443 OK\n")
	}
	fmt.Println("  RÉSULTAT: fallback DNS opérationnel")
	_ = os.Stdout.Sync()
}
