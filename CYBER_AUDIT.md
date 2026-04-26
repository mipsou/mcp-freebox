# Cyber Audit — freebox-mcp

**Date** : 2026-04-26  
**Comité cyber** : 3 rôles (red team offensif · credential/supply chain · injection OWASP)  
**Contexte** : Serveur MCP exposant 65 outils de contrôle de la Freebox à Claude AI.  
**Modèle de menace principal** : prompt injection → attaquant contrôle Claude → appels MCP arbitraires.

---

## Résumé exécutif

| Sévérité  | Findings | Fixés |
|-----------|----------|-------|
| CRITICAL  | 4        | 3 ✅  |
| HIGH      | 7        | 3 ✅  |
| MEDIUM    | 4        | 0     |
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

### [CRITICAL — BACKLOG] C4 — `InsecureSkipVerify: true` dans le client HTTP
- **Vecteur** : Attaquant MITM sur LAN → cert auto-signé → intercepte/modifie tous les appels API
- **Contexte** : La Freebox utilise un cert auto-signé pour son API locale. `InsecureSkipVerify` est nécessaire sans cert pinning.
- **Fichier** : `cmd/freebox-mcp/main.go` (TLS config)
- **Mitigation** : Implémenter le cert pinning via le fingerprint SHA-256 du cert Freebox (récupéré au premier pairing)
- **Action** : Issue GitHub #20 — `fix: TLS cert pinning à la place de InsecureSkipVerify`

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

### [HIGH — BACKLOG] D2 — Session token transmis en clair (HTTP) si TLS désactivé
- **Finding** : Le header `X-Fbx-App-Auth` contient le session token. Si MITM réussit (C4), le token est capturé → replay attack.
- **Mitigation** : Dépend du fix C4 (cert pinning). Une fois TLS valide, le token est protégé.

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

### [HIGH] I1 — NAT : pas de validation `lan_ip` ni des ports
- **Finding** : `lan_ip` non validé → règle NAT vers `0.0.0.0` possible. `wan_port < 1024` non filtré → exposition de ports système.
- **Fichier** : `internal/tools/nat.go`
- **Mitigation** : Valider `lan_ip` via regex IPv4 dans plage RFC1918, rejeter ports < 1024 (ou demander confirmation)
- **Action** : Feature/fix NAT validation — v0.19

### [MEDIUM] I2 — DHCP statique : IP hors subnet (conflit gateway)
- **Finding** : Création bail DHCP pour `192.168.x.1` (gateway) ou `.254` (Freebox) → conflit IP → réseau instable
- **Fichier** : `internal/tools/dhcp.go`
- **Mitigation** : Valider que l'IP n'est pas `.0`, `.1`, `.254`, `.255` du subnet

### [HIGH] I3 — VM : `disk_path` non restreint
- **Finding** : `freebox_vm_create` accepte n'importe quel `disk_path`. Un path vers `/Freebox/system/` ou une image malveillante peut être monté.
- **Fichier** : `internal/tools/vm.go`
- **Mitigation** : Valider que `disk_path` commence par `/Freebox/VMs/` et se termine par `.qcow2` ou `.raw`

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

### Court terme (v0.19)
- [ ] Fix I1 : validation `lan_ip` + ports NAT
- [ ] Fix I3 : restriction `disk_path` VM
- [ ] Fix D4 : `govulncheck` en CI

### Moyen terme
- [ ] Fix C4 : TLS cert pinning (remplace `InsecureSkipVerify`)
- [ ] Fix I2 : validation IP DHCP statique
- [ ] Documentation D1/D3 : sécurité opérationnelle README

### Long terme
- [ ] C5 : "Confirmation gate" pour opérations destructives (NAT, reboot, vm_delete, fs_delete)
- [ ] D2 : Dépend de C4

---

*Voir aussi : [QUALITY_AUDIT.md](QUALITY_AUDIT.md) pour les findings non-cyber (tests, architecture, CI).*
