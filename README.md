# freebox-mcp

Serveur MCP (Model Context Protocol) pour la Freebox Delta via l'API Freebox OS v4.

**65 outils MCP** couvrant : connexion WAN, LAN, DHCP, NAT, pare-feu, WiFi, VPN, VM, stockage, fichiers, système, réseau, contacts, TV, téléchargements, contrôle parental, Wake-on-LAN, AirMedia, journal système, firmware.

## Prérequis

- Go 1.22+
- Une Freebox Delta (ou compatible API v4)
- Accès réseau LAN à la Freebox (mDNS ou `mafreebox.freebox.fr`)

## Installation

```bash
go install github.com/mipsou/mcp-freebox/cmd/freebox-mcp@latest
go install github.com/mipsou/mcp-freebox/cmd/freebox-pair@latest
```

Ou depuis les sources :

```bash
git clone https://github.com/mipsou/mcp-freebox.git
cd mcp-freebox
go build ./cmd/...
```

## Pairing (première fois)

Le pairing génère un `app_token` lié à cette application sur la Freebox.
**freebox-mcp effectue le pairing automatiquement au premier démarrage** si aucun token n'est trouvé.

### Pairing automatique (recommandé)

1. Lancer `freebox-mcp` (ou redémarrer Claude Desktop)
2. Les logs indiquent que la Freebox attend une validation
3. Ouvrir **Freebox OS → Paramètres → Gestion des accès → Applications**
4. Accepter la demande **MCP Freebox**
5. Le token est sauvegardé automatiquement dans le **Windows Credential Manager**

### Pairing manuel (optionnel)

```bash
./freebox-pair
```

1. Le programme contacte la Freebox
2. **Appuyer sur le bouton physique** de la Freebox quand demandé (LED clignote)
3. Le token est sauvegardé dans le Windows Credential Manager

> Le token est une clé d'accès complète à l'API Freebox. Il est stocké de façon sécurisée dans le Windows Credential Manager (DPAPI), jamais en clair.

## Configuration

| Variable             | Défaut                    | Description                                        |
|----------------------|---------------------------|----------------------------------------------------|
| `FREEBOX_APP_ID`     | `mcp-freebox`             | Identifiant applicatif (idem lors du pairing)      |
| `FREEBOX_HOST`       | *(mDNS auto)*             | Hostname ou IP de la Freebox (optionnel)           |
| `FREEBOX_APP_TOKEN`  | *(Credential Manager)*    | Fallback token si le Credential Manager est vide   |

## Intégration Claude Desktop

Dans `claude_desktop_config.json` (token géré par Credential Manager) :

```json
{
  "mcpServers": {
    "freebox-mcp": {
      "command": "C:\\chemin\\vers\\freebox-mcp.exe",
      "env": {
        "FREEBOX_APP_ID": "mcp-freebox"
      }
    }
  }
}
```

## Démarrage

```bash
# Token depuis le Credential Manager (recommandé)
./freebox-mcp

# Token explicite (fallback)
FREEBOX_APP_TOKEN="<token>" ./freebox-mcp
```

Le serveur lit stdin et écrit sur stdout (protocole MCP stdio).

## Outils disponibles (v0.14)

### Connexion WAN

| Outil                        | Description                                              |
|------------------------------|----------------------------------------------------------|
| `freebox_connection_status`  | État ligne, IP publique IPv4/IPv6, débits temps réel     |
| `freebox_connection_ftth`    | SFP présent, signal optique Tx/Rx (dBm×100), liaison     |
| `freebox_connection_xdsl`    | SNR, atténuation, FEC/CRC, uptime xDSL                  |
| `freebox_dyndns_list`        | Entrées DynDNS configurées et leur état                  |

### Découverte

| Outil                  | Description                                                              |
|------------------------|--------------------------------------------------------------------------|
| `freebox_discover`     | Découvre la Freebox sans IP (mDNS) — modèle, version API, port HTTPS    |

### LAN

| Outil                      | Description                                                  |
|----------------------------|--------------------------------------------------------------|
| `freebox_lan_hosts`        | Équipements du réseau local : nom, MAC, IP, accessibilité    |
| `freebox_lan_config`       | Configuration LAN : IP, masque, mode router/bridge, DNS      |
| `freebox_lan_host_rename`  | Renomme un équipement du LAN par son ID                      |

### DHCP

| Outil                        | Description                                                  |
|------------------------------|--------------------------------------------------------------|
| `freebox_dhcp_static`        | Liste les réservations DHCP statiques (MAC → IP fixe)        |
| `freebox_dhcp_leases`        | Liste les baux DHCP dynamiques actifs                        |
| `freebox_dhcp_static_create` | Crée une réservation DHCP statique                           |
| `freebox_dhcp_static_delete` | Supprime une réservation DHCP statique par son ID            |

### NAT / Port forwarding

| Outil                  | Description                                                  |
|------------------------|--------------------------------------------------------------|
| `freebox_nat_rules`    | Liste les règles de redirection de ports NAT                 |
| `freebox_nat_create`   | Crée une règle de port forwarding                            |
| `freebox_nat_toggle`   | Active ou désactive une règle NAT existante                  |
| `freebox_nat_delete`   | Supprime définitivement une règle NAT                        |

### Pare-feu

| Outil                        | Description                                                      |
|------------------------------|------------------------------------------------------------------|
| `freebox_firewall_incoming`  | Règles pare-feu entrantes personnalisées (lecture seule)         |
| `freebox_firewall_dmz`       | Configuration DMZ : état et IP de l'hôte DMZ (lecture seule)    |

### Réseau

| Outil                    | Description                                                        |
|--------------------------|--------------------------------------------------------------------|
| `freebox_routes_ipv4`    | Routes statiques IPv4 (destination, masque, gateway, état)         |
| `freebox_routes_ipv6`    | Routes statiques IPv6 (destination, préfixe, gateway, état)        |
| `freebox_route_add`      | Crée une route statique IPv4                                       |
| `freebox_route_delete`   | Supprime une route statique IPv4                                   |

### Stockage & Fichiers

| Outil                        | Description                                                    |
|------------------------------|----------------------------------------------------------------|
| `freebox_storage_disks`      | Disques connectés : type, état, taille, connecteur             |
| `freebox_storage_partitions` | Partitions : fstype, état de montage, espace libre/utilisé     |
| `freebox_fs_list`            | Browse le stockage Freebox (path base64url) — vérif qcow2 PRA  |

### Partages réseau (Netshare)

| Outil                    | Description                                                      |
|--------------------------|------------------------------------------------------------------|
| `freebox_samba_config`   | État et configuration du serveur Samba (workgroup, logon)        |
| `freebox_samba_shares`   | Liste des partages Samba (nom, chemin, lecture seule)            |
| `freebox_afp_config`     | État du serveur AFP (Apple Filing Protocol)                      |

### VM

| Outil                  | Description                                                          |
|------------------------|----------------------------------------------------------------------|
| `freebox_vm_list`      | Inventaire des VMs : nom, état, mémoire, vCPUs, OS                  |
| `freebox_vm_create`    | Crée une nouvelle VM (nom, mémoire, vCPUs, disque, OS)               |
| `freebox_vm_start`     | Démarre une VM arrêtée (PRA — remontée de service)                   |
| `freebox_vm_stop`      | Arrêt gracieux d'une VM (signal ACPI)                                |
| `freebox_vm_kill`      | Force l'arrêt d'une VM bloquée                                       |
| `freebox_vm_update`    | Modifie la config d'une VM arrêtée                                   |
| `freebox_vm_delete`    | Supprime définitivement une VM                                        |

### WiFi

| Outil                      | Description                                                    |
|----------------------------|----------------------------------------------------------------|
| `freebox_wifi_aps`         | Points d'accès WiFi : bande, canal, état, DFS                  |
| `freebox_wifi_config`      | Configuration WiFi globale (activé, filtre MAC)                |
| `freebox_wifi_toggle`      | Active ou désactive le WiFi global                             |
| `freebox_wifi_ssids`       | Liste des SSIDs (BSS) : nom, bande, chiffrement                |
| `freebox_wifi_stations`    | Clients WiFi connectés : signal (dBm), débits Rx/Tx            |
| `freebox_wifi_ssid_toggle` | Active ou désactive un SSID spécifique                         |

### VPN

| Outil                         | Description                                                         |
|-------------------------------|---------------------------------------------------------------------|
| `freebox_vpn_server_status`   | État des serveurs VPN (PPTP, OpenVPN ×2, IPsec IKEv2, WireGuard)   |
| `freebox_vpn_connections`     | Clients VPN connectés : login, IP source, IP tunnel, routes         |
| `freebox_vpn_client_configs`  | Configurations VPN sortantes (WireGuard/OpenVPN)                    |

### Téléchargements

| Outil                      | Description                                                    |
|----------------------------|----------------------------------------------------------------|
| `freebox_downloads`        | Liste les téléchargements (HTTP/BitTorrent/NZB, statut)        |
| `freebox_download_add`     | Ajoute un téléchargement par URL ou lien magnet                |
| `freebox_download_toggle`  | Pause ou reprise d'un téléchargement                           |
| `freebox_download_delete`  | Supprime un téléchargement de la liste                         |

### Contacts & Appels

| Outil                   | Description                                                      |
|-------------------------|------------------------------------------------------------------|
| `freebox_contacts`      | Répertoire téléphonique : nom, numéros, emails                   |
| `freebox_contact_get`   | Détail complet d'un contact par ID                               |
| `freebox_call_log`      | Journal des appels : entrants, manqués, sortants                 |

### TV

| Outil                    | Description                                                      |
|--------------------------|------------------------------------------------------------------|
| `freebox_tv_channels`    | Chaînes TV disponibles (nom, numéro, qualité SD/HD/UHD)          |
| `freebox_tv_records`     | Enregistrements TV programmés et leur état                       |

### Contrôle parental

| Outil                         | Description                                                  |
|-------------------------------|--------------------------------------------------------------|
| `freebox_parental_config`     | Configuration globale (activé, politique par défaut)         |
| `freebox_parental_planning`   | Plages horaires de restriction                               |
| `freebox_parental_filters`    | Appareils filtrés par adresse MAC                            |

### Wake-on-LAN

| Outil              | Description                                                              |
|--------------------|--------------------------------------------------------------------------|
| `freebox_wol`      | Envoie un magic packet (réveil NAS/PC à distance, SecureOn supporté)     |

### Système

| Outil                    | Description                                                             |
|--------------------------|-------------------------------------------------------------------------|
| `freebox_system`         | Uptime, firmware, températures (CPU, switch), ventilateurs (RPM)        |
| `freebox_reboot`         | Redémarre la Freebox (⚠️ confirmation requise)                          |
| `freebox_switch_ports`   | État des ports LAN : lien, vitesse (Mbps), duplex, MAC + hostname        |

## Architecture interne

```
freebox-mcp/
├── cmd/
│   ├── freebox-mcp/      # Serveur MCP principal (stdio)
│   └── freebox-pair/     # CLI de pairing interactif
├── internal/
│   ├── auth/             # Session HMAC-SHA1, détection de révocation
│   ├── client/           # Client HTTP Freebox avec retry auto
│   ├── config/           # Chargement depuis env vars
│   ├── mdns/             # Découverte mDNS (QU bit, api_domain priority)
│   ├── pair/             # Logique d'appairage réutilisable
│   ├── tools/            # Outils MCP (un fichier par domaine)
│   └── wincred/          # Windows Credential Manager (advapi32.dll)
```

## Licence

[EUPL-1.2](LICENSE) — European Union Public Licence v1.2
