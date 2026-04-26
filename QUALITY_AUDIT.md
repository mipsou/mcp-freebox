# Quality Audit — freebox-mcp

**Date** : 2026-04-26  
**Périmètre** : v0.16.0 + P14 (feature/v0.17-fs-firmware)  
**Outils** : 65 MCP tools, 34 fichiers, 9 packages Go  
**Comité** : 5 rôles (sécurité · tests · architecture · doc/UX · API/release)

---

## Résumé exécutif

| Sévérité | Findings | Fixés |
|----------|----------|-------|
| HIGH     | 7        | 5 ✅  |
| MEDIUM   | 6        | 0     |
| LOW      | 2        | 0     |

---

## Sécurité

### [FIXED ✅] S1 — `go vet`: misuse de `unsafe.Pointer` dans `wincred_windows.go`
- **Avant** : `CredentialBlob uintptr` + arithmétique de pointeur → 2 warnings go vet
- **Fix** : `CredentialBlob unsafe.Pointer`, `unsafe.Slice` pour la lecture, `*credentialW` output parameter (pattern golang.org/x/sys)
- **Résultat** : `go vet ./...` → PASS, aucun warning

### [MEDIUM] S2 — HMAC-SHA1 pour la signature de session
- **Finding** : L'API Freebox impose HMAC-SHA1. Cryptographiquement faible pour de nouvelles conceptions.
- **Contrainte** : Impossible à changer côté client (contrainte Freebox OS)
- **Mitigation** : Les sessions expirent ; le risque est limité au réseau local
- **Action** : Surveiller l'évolution de l'API Freebox (v8+)

### [MEDIUM] S3 — Session token en mémoire non zérisé à la fin de vie
- **Finding** : `Session.Token` (string in-memory) jamais effacé après usage
- **Impact** : Dump mémoire du processus MCP expose le token actif
- **Action future** : Utiliser `[]byte` + `runtime.KeepAlive` + memset sur expiration

### [MEDIUM] S4 — Permissions maximales au pairing
- **Finding** : Les 11 permissions sont toutes demandées par défaut
- **Mitigation** : L'utilisateur peut les restreindre dans l'interface Freebox OS post-pairing
- **Action future** : Envisager un profil "lecture seule" vs "lecture/écriture"

### [MEDIUM] S5 — Pas de redaction des credentials dans les erreurs API
- **Finding** : Les messages d'erreur Freebox sont retournés bruts (jamais de secrets dedans en pratique, mais non garanti)
- **Action future** : Filtrer les messages d'erreur contenant `token`, `password`, `secret`

---

## Qualité des tests

### [FIXED ✅] T1 — Interfaces `writer` sur des fonctions read-only
- **Avant** : 11 fonctions `registerXxx(c writer)` sans jamais appeler Post/Put/Delete
- **Fix** : Dégradées à `getter` — principle of least privilege appliqué
- **Fichiers** : calls, contacts, dhcpconfig, firewall, netshare, parental, syslog, switchconfig, tv, upnp, vpn

### [HIGH] T2 — `mockGetter.Post/Put/Delete` n'enregistrent pas les paths
- **Finding** : Retournent `nil` sans validation → un mauvais path POST passe en vert
- **Contexte** : Les bugs d'endpoints GET sont détectés (le path doit être dans le map) ; seuls POST/PUT/DELETE sont aveugles
- **Action future** : Ajouter un `validatingMock` optionnel enregistrant les calls POST/PUT/DELETE

### [HIGH] T3 — 0% coverage sur 6 packages critiques
- **Packages non testés** : `cmd/freebox-mcp`, `cmd/freebox-pair`, `internal/config`, `internal/mdns`, `internal/pair`, `internal/wincred`
- **Risque** : Le flow de pairing et l'authentification HMAC ne sont pas couverts par des tests automatisés
- **Action future** : Tests d'intégration end-to-end sur environnement Freebox de test

### [HIGH] T4 — Tests NAT sans couverture des opérations mutantes
- **Finding** : `freebox_nat_create`, `freebox_nat_toggle`, `freebox_nat_delete` exposés mais non testés au niveau POST/PUT/DELETE path
- **Action future** : Ajouter tests `TestNatCreate_NoError`, `TestNatToggle_NoError`, `TestNatDelete_NoError`

---

## Architecture

### [FIXED ✅] A1 — getter/writer : principe de moindre privilège
- Voir T1 ci-dessus — 11 fonctions downgraded

### [HIGH] A2 — Croissance "flat file" : 34 fichiers tools sans hiérarchie
- **Finding** : À 65+ outils, la structure plate devient difficile à naviguer
- **Impact** : PRs difficiles à reviewer, recherches lentes, risque de conflicts
- **Action future** : Réorganiser en sous-packages `tools/network/`, `tools/media/`, `tools/vm/`, etc.
- **Note** : Non urgent avant 100 outils — coût de migration élevé

### [MEDIUM] A3 — `RegisterAll` avec 33 appels explicites
- **Finding** : Chaque nouveau tool nécessite une ligne dans RegisterAll
- **Action future** : Auto-registry via `init()` ou slice globale de `RegisterFn`

---

## Documentation & UX

### [HIGH] D1 — README non à jour (affiche 49 outils, réel = 65)
- **Statut** : À corriger dans ce PR (v0.17.0)

### [MEDIUM] D2 — Nommage légèrement incohérent
- **Exemples** : `freebox_nat_rules` vs `freebox_parental_config` (rules vs config pour des listes similaires)
- **Action future** : Établir une convention documentée (`_list` pour les collections, `_config` pour les singletons)

### [OK] D3 — Descriptions MCP
- Toutes en français, détaillées, niveau info correct pour Claude
- `Required()` bien appliqué partout
- `⚠️` sur les opérations destructives (reboot, fs_delete)

---

## API & Release

### [FIXED ✅] CI1 — Absence de pipeline CI/CD
- **Fix** : `.github/workflows/ci.yml` ajouté (go vet + tests + build + golangci-lint)
- **Fix** : `.golangci.yml` avec govet, errcheck, staticcheck, ineffassign

### [FIXED ✅] CI2 — Absence de golangci-lint config
- Voir CI1

### [HIGH] CI3 — Absence de tests d'intégration réels
- **Finding** : Aucun test ne valide les endpoints contre une vraie Freebox
- **Conséquence** : Les bugs d'endpoints passent jusqu'en production (cf. P6 netshare)
- **Action future** : Créer `internal/integration/` avec tests marqués `//go:build integration`, exécutés manuellement contre une vraie Freebox

### [MEDIUM] CI4 — Versions CHANGELOG toutes à la même date
- **Finding** : v0.14.0, v0.15.0, v0.16.0 datées du même jour (développement intensif)
- **Non-critique** : Le format est correct, les entrées sont distinctes

---

## Actions immédiates (v0.17.0)

- [x] Fix `go vet` wincred unsafe.Pointer
- [x] Fix interfaces getter/writer (11 fichiers)
- [x] Ajout CI/CD (.github/workflows/ci.yml + .golangci.yml)
- [ ] Mettre à jour README (65 outils)
- [ ] Tester P14 contre vraie Freebox

## Backlog qualité

- [ ] `validatingMock` pour POST/PUT/DELETE path checking
- [ ] Tests d'intégration `//go:build integration`
- [ ] Réorganisation en sous-packages (après 100 outils)
- [ ] Auto-registry pour RegisterAll
- [ ] Profil permissions minimales au pairing
