# Cyber Audit — freebox-mcp

**Date** : 2026-04-26  
**Comité cyber** : 3 rôles (red team offensif · credential/supply chain · injection OWASP)  
**Contexte** : Serveur MCP exposant 65 outils de contrôle de la Freebox à Claude AI.  
**Modèle de menace principal** : prompt injection → attaquant contrôle Claude → appels MCP arbitraires.

---

## Résumé exécutif

| Sévérité  | Findings | Fixés |
|-----------|----------|-------|
| CRITICAL  | 3        | 3 ✅  |
| HIGH      | 7        | 5 ✅  |
| MEDIUM    | 5        | 0     |
| LOW       | 1        | 0     |

---

## RED TEAM — Surface d'attaque

### [FIXED ✅] C1 — Path Traversal dans `encodeFSPath`
- **Vecteur** : `path="/../etc/passwd"` → encodé base64 → envoyé à l'API `/fs/ls/`
- **Impact** : CRITICAL — lecture/suppression de fichiers système si l'API Freebox accepte les traversées
- **Fix** : `sanitizeFSPath()` — rejette `..`, valide via `path.Clean`, refuse `/`
- **Tests** : `TestSanitizeFSPath_TraversalRejected`, `TestFSList_TraversalBlocked`, `TestFSDelete_TraversalBlocked`

### [FIXED ✅] C2 — SSRF via `freebox_download_add`
- **Vecteur** : `url="file:///etc/passwd"` ou `url="http://169.254.169.254/latest/meta-data/"`
- **Impact** : CRITICAL — lecture fichiers locaux Freebox, accès metadata cloud si Freebox est en DMZ
- **Fix** : `validateDownloadURL()` — whitelist schemes (`http`, `https`, `magnet`, `nzb`), blocage loopback et link-local (169.254.x.x)
- **Tests** : `TestValidateDownloadURL_FileSchemeBlocked`, `TestValidateDownloadURL_LinkLocalBlocked`, etc.

### [FIXED ✅] C3 — Injection via adresse MAC invalide (WoL)
- **Vecteur** : `mac="'; DROP TABLE--"` ou `password=<10MB string>`
- **Impact** : HIGH — injection indirecte si l'API parse le champ MAC, DoS mémoire sur Freebox
- **Fix** : `validateMAC()` regex `([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}`, `validateSecureOn()` limite 17 chars
- **Tests** : `TestValidateMAC_Invalid`, `TestValidateSecureOn_TooLong`, `TestWOL_InvalidMAC`

### [MEDIUM — BACKLOG] C4 — `InsecureSkipVerify: true` dans le client HTTP

> ⚠️ **Correction (2026-04-27)** : Classification initiale CRITICAL incorrecte — reclassifié MEDIUM après analyse des pratiques TLS réelles de la Freebox.

- **Vecteur** : Attaquant MITM sur LAN → interception des appels API
- **Périmètre réel** :
  - HTTP est utilisé **uniquement** pour la découverte (`GET /api_version`) — endpoint public, non authentifié, par conception de l'API Freebox OS
  - HTTPS est utilisé pour **tous les appels authentifiés** vers `mafreebox.freebox.fr`
  - Le cert Freebox est auto-signé par la **propre CA de Free** (pas une CA publique, pas Let's Encrypt)
  - **Free ne publie pas sa CA** → aucun bundle officiel n'est disponible pour le pinning standard
  - `InsecureSkipVerify: true` est **documenté intentionnellement** dans le code (commentaire explicite)
- **Menace réelle** : MITM sur réseau LAN local uniquement — l'API Freebox n'est jamais exposée sur Internet
- **Atténuation actuelle** : Acceptable comme tradeoff documenté tant que la CA Free reste non publiée
- **Piste future (non triviale)** : TOFU (Trust On First Use) — capturer et stocker le fingerprint SHA-256 du cert lors du premier pairing, puis valider à chaque connexion. Nécessite refactoring du flow `freebox_auth`.
- **Fichier** : `cmd/freebox-mcp/main.go` (TLS config)

### [CRITICAL — BACKLOG] C5 — Prompt injection → création règle NAT/firewall
- **Vecteur** : Attaquant injecte dans un prompt → Claude appelle `freebox_nat_create` avec `wan_port=22`, `lan_ip=<attacker-machine>` → SSH exposé sur Internet
- **Impact** : CRITICAL — intrusion réseau LAN complet
- **Mitigation proposée** : Pas de validation possible côté MCP sans connaître la politique de l'utilisateur. **Approche recommandée** : documenter explicitement dans la description des tools que ces opérations sont destructives et exigent confirmation. Envisager un outil `freebox_nat_preview` (lecture seule avant création).
- **Action** : Étude d'impact — "Confirmation gate" dans le server MCP

---

## CREDENTIALS & SUPPLY CHAIN

### [MEDIUM] D1 — `app_token` lisible par les processus admin Windows
- **Finding** : DPAPI protège le credential Windows Credential Manager, mais tout processus avec les droits `SYSTEM` ou `SeDebugPrivilege` peut lire via `mimikatz`-like ou `CryptUnprotectData`
- **Mitigation** : Documenter clairement que `freebox-mcp` doit tourner en utilisateur non-privilégié. Ne jamais lancer en tant qu'admin.
- **Action** : Ajouter dans README : "Sécurité — Exécution en utilisateur standard recommandée"

### [HIGH — BACKLOG] D2 — Session token exposable en cas de MITM LAN (lié à C4)
- **Finding** : Le header `X-Fbx-App-Auth` contient le session token. Si MITM LAN réussit via C4, le token peut être capturé → replay attack.
- **Contexte** : HTTPS est actif sur tous les appels authentifiés ; HTTP est réservé à `/api_version` (non authentifié). Le risque D2 se matérialise **uniquement** si l'attaquant réussit un MITM complet via C4 (LAN, `InsecureSkipVerify`).
- **Mitigation** : Dépend de la résolution TOFU de C4. Jusqu'à cette implémentation, le token reste protégé par HTTPS si aucun MITM actif n'est présent sur le LAN.

### [HIGH] D3 — `app_token` accepté via variable d'environnement `FREEBOX_APP_TOKEN`
- **Finding** : `os.Getenv("FREEBOX_APP_TOKEN")` dans `main.go` permet de bypasser Windows Credential Manager. Env vars visibles dans `ps aux`, logs CI, core dumps.
- **Mitigation** : Documenter : "Usage de la variable d'environnement réservé aux environnements de test — ne jamais utiliser en production"

### [MEDIUM] D4 — Dépendances tierces (supply chain)
- **Packages directs** : `mark3labs/mcp-go` v0.49.0, `miekg/dns` v1.1.72
- **Transitifs** : `google/jsonschema-go`, `yosida95/uritemplate/v3`, `google/uuid`, `spf13/cast`
- **Risque** : `mcp-go` est le cœur du serveur — une compromission serait catastrophique
- **Mitigation** : `go.sum` vérifié par CI (déjà en place). Envisager `govulncheck` dans le pipeline.
- **Action** : Ajouter `govulncheck ./...` dans `.github/workflows/ci.yml`

---

## INJECTION APPLICATIVE (OWASP)

### [FIXED ✅] I1 — NAT : validation `lan_ip` RFC1918 + ports
- **Finding** : `lan_ip` non validé → règle NAT vers IP publique possible. Ports hors plage non rejetés.
- **Fichier** : `internal/tools/nat.go`
- **Fix** : `validateRFC1918()` — IP publique rejetée. `validatePort()` — plage 1–65535 enforced. Contraintes schema : `mcp.Pattern(RFC1918Pattern)`, `mcp.Min(1)/Max(65535)`, `mcp.Enum("tcp", "udp")`.
- **Tests** : `TestValidateRFC1918_*`, `TestValidatePort_*`, `TestNATCreate_InvalidIP`, `TestNATCreate_InvalidPort`

### [FIXED ✅] I2 — DHCP statique : IP réservée (gateway, broadcast)
- **Finding** : Création bail DHCP pour `.1` (gateway) ou `.254` (Freebox) → conflit IP → réseau instable
- **Fichier** : `internal/tools/dhcp.go`
- **Fix** : `validateDHCPIP()` — rejette derniers octets `.0`, `.1`, `.254`, `.255`. `validateMAC()` sur le champ MAC. Contraintes schema : `mcp.Pattern(MACAddrPattern)`, `mcp.Pattern(IPv4Pattern)`.
- **Tests** : `TestValidateDHCPIP_ReservedRejected`, `TestDHCPStaticCreate_InvalidMAC/ReservedIP`

### [FIXED ✅] I3 — VM : `disk_path` non restreint → refonte `disk_name`
- **Finding** : `freebox_vm_create` acceptait n'importe quel `disk_path`. Accès hors `/Freebox/VMs/` possible.
- **Fichier** : `internal/tools/vm.go`
- **Fix (Security by Design)** : Paramètre `disk_path` supprimé et remplacé par `disk_name` (nom de fichier uniquement). Le chemin `/Freebox/VMs/<nom>` est construit **par le code**, jamais fourni par l'appelant. Injection de chemin arbitraire structurellement impossible.
- **Tests** : `TestValidateDiskName_Invalid`, `TestVMCreate_InvalidDiskName`, `TestVMCreate_DiskPathConstructed`

### [LOW] I4 — Session token TTL client/serveur non synchronisé
- **Finding** : TTL client hardcodé à 25 min ; TTL serveur Freebox = 30 min. Pas de rafraîchissement proactif.
- **Impact** : Faible — au pire une requête échoue et l'auto-retry renouvelle la session
- **Action** : Amélioration future

---

## Plan d'action (priorité)

### Immédiat (v0.18 — ce PR)
- [x] Fix C1 : path traversal `sanitizeFSPath`
- [x] Fix C2 : SSRF `validateDownloadURL`
- [x] Fix C3 : validation MAC/SecureOn WoL

### Court terme (v0.19) — ✅ LIVRÉ
- [x] Fix I1 : validation `lan_ip` RFC1918 + ports NAT (schema + runtime)
- [x] Fix I2 : validation IP DHCP statique (octets réservés)
- [x] Fix I3 : refonte `disk_path`→`disk_name` — security by design (chemin construit par le code)
- [x] Architecture : `validate.go` — module centralisé (fin du hardening dispersé)

### Moyen terme
- [ ] Fix C4 : TOFU cert pinning (capturer fingerprint SHA-256 au premier pairing — **Free ne publie pas sa CA**, cert pinning naïf impossible)
- [ ] Fix I2 : validation IP DHCP statique
- [ ] Documentation D1/D3 : sécurité opérationnelle README

### Long terme
- [ ] C5 : "Confirmation gate" pour opérations destructives (NAT, reboot, vm_delete, fs_delete)
- [ ] D2 : Dépend de C4

---

*Voir aussi : [QUALITY_AUDIT.md](QUALITY_AUDIT.md) pour les findings non-cyber (tests, architecture, CI).*
