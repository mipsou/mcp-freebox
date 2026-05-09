# Changelog

Toutes les modifications notables de ce projet sont documentées ici.

Format : [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/)
Versionnage : [Semantic Versioning](https://semver.org/lang/fr/)

---

## [Unreleased]

### Corrigé
- `freebox_vm_create` : `disk_path` désormais encodé en base64 standard avant envoi à l'API (cohérence avec `cd_path`, `fs/`, `download_dir`). Corrige `invalid_request` persistant sur création VM (#73).
- `freebox_vm_list` / `freebox_vm_create` : custom `UnmarshalJSON` sur `BindUSBPorts` pour tolérer le quirk de l'API Freebox qui renvoie `""` (string vide) au lieu de `[]` quand aucun port USB n'est lié. Débloque l'usage de `freebox_vm_list` dès qu'au moins une VM existe (#76).

---

## [0.19.0] - 2026-04-27

### Sécurité — Security by Design (contraintes dans le contrat API, pas dans des couches de hardening)

#### Architecture : `validate.go` — module de validation centralisé
- Source de vérité unique pour toutes les validations d'entrée du package `tools`
- Consolidation des validators dispersés (MAC depuis `wol.go`, SSRF depuis `downloads.go`)
- Constantes de patterns exportées (`MACAddrPattern`, `RFC1918Pattern`, `DiskNamePattern`, `IPv4Pattern`) utilisables comme contraintes schema `mcp.Pattern()`

#### `freebox_vm_create` — refonte structurelle `disk_path` → `disk_name`
- **Avant** : l'appelant fournissait un chemin complet → validation par patch runtime
- **Après** : l'appelant fournit uniquement le nom de fichier (`fedora.qcow2`) → le code construit `/Freebox/VMs/<nom>` en interne
- Injection de chemin arbitraire structurellement impossible sans validator ad-hoc
- Contrainte schema : `mcp.Pattern(DiskNamePattern)`, `mcp.Enum("qcow2", "raw")` sur `disk_type`, `mcp.Enum(...)` sur `os`
- Contraintes numériques : `mcp.Min(64)/Max(16384)` sur `memory`, `mcp.Min(1)/Max(8)` sur `vcpus`

#### `freebox_nat_create` — NAT : validation RFC1918 + ports
- `lan_ip` : `mcp.Pattern(RFC1918Pattern)` + `validateRFC1918()` — IP publique rejetée par design
- Ports : `mcp.Min(1)/Max(65535)` + `validatePort()` — port 0 et overflow rejetés
- `ip_proto` : `mcp.Enum("tcp", "udp")` — protocole contraint à l'espace valide

#### `freebox_dhcp_static_create` — DHCP : validation MAC + IP réservée
- `mac` : `mcp.Pattern(MACAddrPattern)` + `validateMAC()` — format strict
- `ip` : `mcp.Pattern(IPv4Pattern)` + `validateDHCPIP()` — derniers octets .0/.1/.254/.255 rejetés (gateway, réseau, broadcast, Freebox)

#### `freebox_wol` — contraintes schema
- `mac` : `mcp.Pattern(MACAddrPattern)` — le schéma guide le client vers le bon format
- `password` : `mcp.MaxLength(17)` — limite déclarée dans le contrat

### Tests ajoutés (26 nouveaux tests de sécurité)
- `TestValidateRFC1918_Valid/Invalid`, `TestValidatePort_Valid/Invalid`
- `TestNATCreate_InvalidIP`, `TestNATCreate_InvalidPort`, `TestNATCreate_OK`
- `TestValidateDHCPIP_Valid/ReservedRejected`, `TestDHCPStaticCreate_InvalidMAC/ReservedIP/OK`
- `TestValidateDiskName_Valid/Invalid`, `TestVMCreate_InvalidDiskName`, `TestVMCreate_DiskPathConstructed`

---

## [0.18.0] - 2026-04-26

### Sécurité — Audit comité cyber (3 rôles : red team · credentials · OWASP)
- **CRITICAL** : `freebox_fs_list/mkdir/delete` — `sanitizeFSPath()` bloque les path traversal (`..`, `/`)
- **CRITICAL** : `freebox_download_add` — `validateDownloadURL()` : whitelist `http/https/magnet/nzb`, blocage loopback et `169.254.x.x` (SSRF)
- **HIGH** : `freebox_wol` — validation regex adresse MAC, limite taille SecureOn (17 chars max)
- **CI** : `govulncheck ./...` ajouté au pipeline GitHub Actions

### Ajouté
- `CYBER_AUDIT.md` — rapport complet (16 findings, plan d'action priorisé)

---

## [0.17.0] - 2026-04-26

### Ajouté
- **Firmware** : `freebox_firmware_update_status` — vérification disponibilité mise à jour
- **AirMedia** : `freebox_airmedia_config` — configuration AirMedia (activé, mot de passe)
- **AirMedia** : `freebox_airmedia_receivers` — liste des récepteurs (capacités photo/vidéo/audio)
- **Connexion** : `freebox_connection_config` — config WAN (ping, accès distant, Wake-on-LAN port)
- **Téléchargements** : `freebox_download_config` — config gestionnaire (vitesse, dossier, ratio seeding)
- **Fichiers** : `freebox_fs_mkdir` — création répertoire (tâche asynchrone)
- **Fichiers** : `freebox_fs_delete` — suppression fichier/répertoire ⚠️ (tâche asynchrone)
- **CI** : `.github/workflows/ci.yml` — pipeline go vet + tests + build + golangci-lint
- **CI** : `.golangci.yml` — configuration linter (govet, errcheck, staticcheck, ineffassign)
- **Qualité** : `QUALITY_AUDIT.md` — rapport comité multidisciplinaire (5 rôles)

### Modifié
- **Sécurité** : `wincred` — fix `unsafe.Pointer` via `unsafe.Slice` + `*credentialW` output pattern ; `go vet` PASS
- **Architecture** : 11 fonctions `registerXxx` dégradées `writer→getter` (principle of least privilege)
- **README** : mis à jour (49→65 outils)

---

## [0.16.0] - 2026-04-27

### Ajouté
- **Switch** : `freebox_switch_port_config` — configuration d'un port LAN (activé, duplex, vitesse)
- **Switch** : `freebox_switch_port_stats` — statistiques temps réel d'un port (débits Rx/Tx, erreurs)
- **Système** : `freebox_system_log` — journal système Freebox (événements, niveaux, catégories)

---

## [0.15.0] - 2026-04-27

### Ajouté
- **DHCP** : `freebox_dhcp_config` — configuration du serveur DHCP (plage, passerelle, masque, DNS)
- **UPnP** : `freebox_upnp_config` — état UPnP/IGD (activé/désactivé)
- **UPnP** : `freebox_upnp_rules` — règles IGD créées automatiquement par UPnP (Plex, jeux, etc.)
- **LCD** : `freebox_lcd_config` — état de l'écran LCD (luminosité, orientation)
- **LCD** : `freebox_lcd_brightness` — règle la luminosité de l'écran LCD (0-100)

### Modifié
- **freebox-pair** : contacts, guide TV, enregistrements marqués `usedByMCP: true`
- **README** : mis à jour avec les 55 outils (v0.6 → v0.14)

---

## [0.14.0] - 2026-04-27

### Ajouté
- **TV** : `freebox_tv_channels` — liste des chaînes TV (nom, numéro, qualité SD/HD/UHD)
- **TV** : `freebox_tv_records` — enregistrements TV programmés et leur état

---

## [0.13.0] - 2026-04-27

### Ajouté
- **WiFi** : `freebox_wifi_ssids` — liste des SSIDs (BSS) : nom, bande, chiffrement, état caché/visible
- **WiFi** : `freebox_wifi_stations` — clients WiFi actuellement connectés (MAC, bande, signal dBm, débit Rx/Tx)
- **WiFi** : `freebox_wifi_ssid_toggle` — activer ou désactiver un SSID spécifique par son ID

---

## [0.12.0] - 2026-04-27

### Ajouté
- **LAN** : `freebox_lan_config` — configuration réseau LAN (IP, masque, mode router/bridge, noms DNS/mDNS/NetBIOS)
- **LAN** : `freebox_lan_host_rename` — renommer un équipement du réseau local par son ID

---

## [0.11.0] - 2026-04-27

### Ajouté
- **Réseau** : `freebox_routes_ipv4` — liste des routes statiques IPv4 (destination, masque, gateway, état)
- **Réseau** : `freebox_routes_ipv6` — liste des routes statiques IPv6 (destination, préfixe, gateway, état)
- **Réseau** : `freebox_route_add` — création d'une route statique IPv4
- **Réseau** : `freebox_route_delete` — suppression d'une route statique IPv4
- **Système** : `freebox_reboot` — redémarrage de la Freebox (confirmation explicite recommandée)

---

## [0.10.0] - 2026-04-25

### Ajouté
- **Contacts** : `freebox_contacts` — liste le répertoire téléphonique (nom, numéros, emails)
- **Contacts** : `freebox_contact_get` — détail complet d'un contact par ID
- **Wake-on-LAN** : `freebox_wol` — envoie un magic packet à une adresse MAC pour démarrer un équipement à distance (NAS, PC) ; supporte SecureOn password et choix d'interface

---

## [0.9.1] - 2026-04-25

### Corrigé
- **Netshare** : chemins API `/share/samba/` → `/netshare/samba/`, `/share/afp/` → `/netshare/afp/` (endpoint réel Freebox)
- **freebox-pair** : `usedByMCP: true` sur fichiers, journal d'appels, téléchargements ; permission **Contrôle parental** ajoutée (re-pairing requis pour activer les outils parental)

---

## [0.9.0] - 2026-04-25

### Ajouté

- **Netshare** : `freebox_samba_config` — état et configuration du serveur Samba (workgroup, logon, imprimante)
- **Netshare** : `freebox_samba_shares` — liste des partages Samba (nom, chemin, lecture seule)
- **Netshare** : `freebox_afp_config` — état du serveur AFP (Apple Filing Protocol)
- **Downloads** : `freebox_downloads` — liste des téléchargements (HTTP, BitTorrent, NZB) avec statut et progression
- **Downloads** : `freebox_download_add` — ajout d'un téléchargement par URL ou lien magnet
- **Downloads** : `freebox_download_toggle` — pause ou reprise d'un téléchargement
- **Downloads** : `freebox_download_delete` — suppression d'un téléchargement de la liste
- **Calls** : `freebox_call_log` — journal des appels téléphoniques (entrants, manqués, sortants)
- **Parental** : `freebox_parental_config` — configuration globale du contrôle parental (activé, politique par défaut)
- **Parental** : `freebox_parental_planning` — plages horaires de restriction par jour
- **Parental** : `freebox_parental_filters` — appareils soumis au contrôle parental (filtre MAC)

---

## [0.8.0] - 2026-04-25

### Ajouté
- **Système de fichiers** : `freebox_fs_list` — browse le stockage Freebox via `GET /fs/ls/{path base64url}` ; utile en PRA pour vérifier les images qcow2 dans `/Freebox/VMs/`

### Corrigé
- `StorageDisk.connector` : `string` → `int` (enum connecteur API réel)
- `StoragePartition.id` : `string` → `int` (ID numérique API réel)

### Interne
- `callTool()` refactorisé sur `callToolWithArgs()` — suppression du doublon

---

## [0.7.1] - 2026-04-24

### Correctif

- **Démarrage non-bloquant** : le serveur MCP démarre immédiatement même si le token est absent ou révoqué — l'auto-pairing tourne en goroutine, les outils retournent un message clair et deviennent opérationnels dès l'approbation Freebox OS, sans redémarrage

---

## [0.7.0] - 2026-04-24

### Ajouté

- **VPN Serveur** : `freebox_vpn_server_status` — état des 5 protocoles (PPTP, OpenVPN Routé/Bridgé, IPsec IKEv2, WireGuard) avec nombre de connexions actives
- **VPN Serveur** : `freebox_vpn_connections` — clients VPN actuellement connectés (login, IP source, IP tunnel, routes poussées)
- **VPN Client** : `freebox_vpn_client_configs` — configurations VPN sortant vers un serveur externe
- Permission **"Gestion du VPN"** dans `freebox-pair` — re-pairing requis pour activer les outils VPN

---

## [0.6.0] - 2026-04-24

### Ajouté
- **VM CRUD complet** : `freebox_vm_create` (POST /vm/), `freebox_vm_stop` (ACPI gracieux), `freebox_vm_update` (PATCH config), `freebox_vm_delete` (DELETE /vm/{id})
- **Pare-feu** (lecture seule) : `freebox_firewall_incoming` (règles entrantes), `freebox_firewall_dmz` (config DMZ)
- **NAT CRUD** : `freebox_nat_create`, `freebox_nat_toggle`, `freebox_nat_delete` (en complément de `freebox_nat_rules` existant)
- **DHCP CRUD** : `freebox_dhcp_static_create`, `freebox_dhcp_static_delete`
- **WiFi toggle** : `freebox_wifi_toggle` (PUT /wifi/config/ `{"enabled": bool}`)

### Modifié
- `cmd/freebox-pair/main.go` : refactorisé pour utiliser `internal/pair.Start()` et `internal/pair.WaitForGrant()` — suppression du code dupliqué `requestToken`/`waitForGrant`
- `cmd/freebox-pair/main.go` : permissions mises à jour (description "Modification des réglages" couvre WiFi, NAT, DHCP, pare-feu ; "Contrôle de la VM" couvre CRUD complet) ; version bump v0.4 → v0.6
- `internal/tools/vm.go` : `freebox_vm_start` et `freebox_vm_kill` migrent vers `toFloat()` pour cohérence avec les autres outils
- README : documentation complète de tous les outils (23 outils total), section architecture interne, pairing automatique

### Corrigé
- `internal/tools/wifi.go` : correction copyright header (doublon `chpujol@mitjeu`)

---

## [0.5.1] - 2026-04-24

### Corrigé
- **mDNS** : ajout du bit QU (unicast-response) dans la query PTR — la Freebox répond en unicast, contournant le pare-feu Windows sur 224.0.0.251:5353
- **`parseResponse()`** : priorité hôte `api_domain` (TXT) > A/AAAA > SRV target ; port SRV (HTTP 80) ignoré — `https_port` vient exclusivement du TXT record (ex : 42460)

### Ajouté
- `cmd/mdns-debug/` : outil de diagnostic mDNS (build ignore) — 4 tests : écoute passive, QU, multicast, fallback DNS

---

## [0.5.0] - 2026-04-24

### Ajouté
- **Auto-pairing au premier démarrage** : `freebox-mcp` initie automatiquement l'appairage si aucun token n'est trouvé — message clair dans les logs (`Freebox OS → Gestion des accès → Applications`)
- **`internal/wincred`** : lecture/écriture Windows Credential Manager via `advapi32.dll` (native, sans module PowerShell externe)
- **`internal/pair`** : logique d'appairage extraite en package réutilisable (`Start`, `WaitForGrant`)
- **Détection de révocation** : `error_code: invalid_token / pending_token` → wincred effacé → re-pair automatique au prochain démarrage
- **Validation eagère du token** au démarrage — vérifie la validité avant d'accepter le token wincred

### Modifié
- `cmd/freebox-mcp/main.go` : `loadAppToken()` remplacé par `acquireToken()` — cascade wincred → env var → auto-pair
- `internal/auth` : `ErrTokenRevoked` sentinelle sur `error_code: invalid_token / pending_token`

---

## [0.4.1] - 2026-04-24

### Ajouté
- **Scripts PowerShell** : `pair-and-save.ps1` — build + pairing interactif + sauvegarde token via `cmdkey`
- **Scripts PowerShell** : `load-token.ps1` — chargement du token depuis Credential Manager

### Modifié
- **mDNS** : découverte automatique de la Freebox sur le LAN — zéro config, zéro IP hardcodée ; fallback `mafreebox.freebox.fr` si timeout
- `freebox-pair` : correction label version, ajout droit "Modification des réglages" nécessaire pour `system`/`switch`

### Corrigé
- Scripts : suppression IP LAN privée hardcodée dans `pair-and-save.ps1`
- Scripts : `ErrorActionPreference Continue` autour de `freebox-pair.exe` (faux positifs stderr Go)
- mDNS : remplacement `grandcat/zeroconf` par implémentation directe (audit sécurité Dependabot)

---

## [0.4.0] - 2026-04-24

### Ajouté
- **Système** : outil `freebox_system` — uptime, version firmware, board name, températures (CPU, switch), vitesse ventilateurs
- **Switch** : outil `freebox_switch_ports` — état de chaque port LAN (lien, vitesse Mbps, duplex, équipements connectés avec MAC + hostname)
- Types Go structurés : `SystemInfo`, `SystemSensor`, `SystemFan`, `SwitchPortStatus`, `SwitchMacEntry`
- Tests unitaires OK/APIError pour les deux nouveaux outils

---

## [0.3.0] - 2026-04-23

### Ajouté
- **Découverte** : outil `freebox_discover` — identifie la Freebox sur le réseau local sans IP connue (mDNS via `mafreebox.freebox.fr`, non authentifié)
- **Stockage** : outil `freebox_storage_disks` — liste les disques connectés (état, taille, connecteur)
- **Stockage** : outil `freebox_storage_partitions` — liste les partitions (fstype, montage, espace libre)
- **VM** : outil `freebox_vm_list` — inventaire des VMs (nom, état, mémoire, vCPUs, OS)
- **VM** : outil `freebox_vm_start` — démarre une VM arrêtée (PRA : remontée de service)
- **VM** : outil `freebox_vm_kill` — force l'arrêt d'une VM bloquée
- `client.DiscoverAPI()` — requête HTTP non authentifiée vers `/api_version` (pas d'enveloppe Freebox)
- `config.DiscoveryURL()` — URL de découverte `http://{host}/api_version`
- Permission **"Contrôle de la VM"** marquée `usedByMCP: true` dans `freebox-pair` (re-pairing requis)

### Modifié
- `client.New()` — nouveau paramètre `discoverURL`
- `tools.RegisterAll()` — nouveau paramètre `discoverer`

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

[Unreleased]: https://github.com/mipsou/mcp-freebox/compare/v0.18.0...HEAD
[0.18.0]: https://github.com/mipsou/mcp-freebox/compare/v0.17.0...v0.18.0
[0.17.0]: https://github.com/mipsou/mcp-freebox/compare/v0.16.0...v0.17.0
[0.16.0]: https://github.com/mipsou/mcp-freebox/compare/v0.15.0...v0.16.0
[0.15.0]: https://github.com/mipsou/mcp-freebox/compare/v0.14.0...v0.15.0
[0.14.0]: https://github.com/mipsou/mcp-freebox/compare/v0.13.0...v0.14.0
[0.13.0]: https://github.com/mipsou/mcp-freebox/compare/v0.12.0...v0.13.0
[0.12.0]: https://github.com/mipsou/mcp-freebox/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/mipsou/mcp-freebox/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/mipsou/mcp-freebox/compare/v0.9.1...v0.10.0
[0.9.1]: https://github.com/mipsou/mcp-freebox/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/mipsou/mcp-freebox/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/mipsou/mcp-freebox/compare/v0.7.1...v0.8.0
[0.7.1]: https://github.com/mipsou/mcp-freebox/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/mipsou/mcp-freebox/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/mipsou/mcp-freebox/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/mipsou/mcp-freebox/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/mipsou/mcp-freebox/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/mipsou/mcp-freebox/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/mipsou/mcp-freebox/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/mipsou/mcp-freebox/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/mipsou/mcp-freebox/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/mipsou/mcp-freebox/releases/tag/v0.1.0
