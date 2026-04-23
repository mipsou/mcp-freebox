# Changelog

Toutes les modifications notables de ce projet sont documentées ici.

Format : [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/)
Versionnage : [Semantic Versioning](https://semver.org/lang/fr/)

---

## [Unreleased]

---

## [0.2.0] - 2026-04-23

### Ajouté
- **LAN** : outil `freebox_lan_hosts` — liste les équipements du réseau local (nom, MAC, IP, type d'hôte, accessibilité)
- **DHCP** : outil `freebox_dhcp_static` — réservations DHCP statiques (MAC → IP fixe)
- **DHCP** : outil `freebox_dhcp_leases` — baux DHCP dynamiques actifs
- **NAT** : outil `freebox_nat_rules` — règles de redirection de ports (port forwarding)
- **WiFi** : outil `freebox_wifi_aps` — points d'accès WiFi (bande, canal, état, DFS)
- **WiFi** : outil `freebox_wifi_config` — configuration WiFi globale (activé, filtre MAC)
- Types Go structurés pour tous les nouveaux endpoints (pas de `json.RawMessage`)
- Tests unitaires pour chaque outil (pattern `mockGetter`)
- Section v0.2 dans le README

---

## [0.1.0] - 2026-04-23

### Ajouté
- Scaffold initial : module `github.com/mipsou/mcp-freebox`, dépendance `mark3labs/mcp-go`
- **Auth** : `internal/auth` — session HMAC-SHA1, mutex sur `Manager`, `Invalidate()` pour retry
- **Client** : `internal/client` — retry automatique sur `auth_required`, replay du body
- **Config** : `internal/config` — chargement depuis variables d'environnement (`FREEBOX_APP_TOKEN`, `FREEBOX_HOST`, `FREEBOX_APP_ID`)
- **Connexion WAN** : outils `freebox_connection_status`, `freebox_connection_xdsl`, `freebox_connection_ftth`, `freebox_dyndns_list`
- **freebox-pair** : CLI d'appairage interactif — affichage des permissions (nécessaires vs extras), double confirmation, token sur stdout
- Tests unitaires : auth (vecteur RFC 2202), client, outils connection (pattern `mockGetter`)
- README : prérequis, installation, pairing, configuration, intégration Claude Desktop
- Licence EUPL-1.2

---

[Unreleased]: https://github.com/mipsou/mcp-freebox/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/mipsou/mcp-freebox/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/mipsou/mcp-freebox/releases/tag/v0.1.0
