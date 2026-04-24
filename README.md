# freebox-mcp

Serveur MCP (Model Context Protocol) pour la Freebox Delta via l'API Freebox OS v4.

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
3. Le token s'affiche sur stdout — le sauvegarder si besoin

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

## Outils disponibles (v0.6)

### Connexion WAN

| Outil                        | Description                                              |
|------------------------------|----------------------------------------------------------|
| `freebox_connection_status`  | État ligne, IP publique IPv4/IPv6, débits temps réel     |
| `freebox_connection_ftth`    | SFP présent, signal optique Tx/Rx (dBm×100), liaison     |
| `freebox_connection_xdsl`    | SNR, atténuation, FEC/CRC, uptime xDSL                  |
| `freebox_dyndns_list`        | Entrées DynDNS configurées et leur état                  |

### Découverte

| Outil                  | Description                                                       |
|------------------------|-------------------------------------------------------------------|
| `freebox_discover`     | Découvre la Freebox sans IP (mDNS) — modèle, version API, port HTTPS |

### LAN

| Outil                  | Description                                                  |
|------------------------|--------------------------------------------------------------|
| `freebox_lan_hosts`    | Équipements du réseau local : nom, MAC, IP, accessibilité    |

### DHCP

| Outil                        | Description                                                  |
|------------------------------|--------------------------------------------------------------|
| `freebox_dhcp_static`        | Liste les réservations DHCP statiques (MAC → IP fixe)        |
| `freebox_dhcp_leases`        | Liste les baux DHCP dynamiques actifs (clients connectés)    |
| `freebox_dhcp_static_create` | Crée une réservation DHCP statique (MAC → IP fixe)           |
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
| `freebox_firewall_incoming`  | Liste les règles pare-feu entrantes personnalisées (lecture seule) |
| `freebox_firewall_dmz`       | Configuration DMZ : état et IP de l'hôte DMZ (lecture seule)    |

### Stockage

| Outil                        | Description                                                    |
|------------------------------|----------------------------------------------------------------|
| `freebox_storage_disks`      | Disques connectés : type, état, taille, connecteur             |
| `freebox_storage_partitions` | Partitions : fstype, état de montage, espace libre/utilisé     |

### VM

| Outil                  | Description                                                          |
|------------------------|----------------------------------------------------------------------|
| `freebox_vm_list`      | Inventaire des VMs : nom, état, mémoire, vCPUs, OS                  |
| `freebox_vm_create`    | Crée une nouvelle VM (nom, mémoire, vCPUs, disque, OS)               |
| `freebox_vm_start`     | Démarre une VM arrêtée (PRA — remontée de service)                   |
| `freebox_vm_stop`      | Arrêt gracieux d'une VM (signal ACPI/shutdown)                       |
| `freebox_vm_kill`      | Force l'arrêt d'une VM bloquée (coupure secteur)                     |
| `freebox_vm_update`    | Modifie la config d'une VM arrêtée (nom, mémoire, vCPUs, écran)     |
| `freebox_vm_delete`    | Supprime définitivement une VM et libère son disque                   |

### WiFi

| Outil                  | Description                                                  |
|------------------------|--------------------------------------------------------------|
| `freebox_wifi_aps`     | Points d'accès WiFi : bande, canal, état, DFS                |
| `freebox_wifi_config`  | Configuration WiFi globale (activé, filtre MAC)              |
| `freebox_wifi_toggle`  | Active ou désactive le WiFi global                           |

### Système

| Outil                  | Description                                                             |
|------------------------|-------------------------------------------------------------------------|
| `freebox_system`       | Uptime, version firmware, températures (CPU, switch), ventilateurs (RPM)|

### Switch LAN

| Outil                    | Description                                                        |
|--------------------------|--------------------------------------------------------------------|
| `freebox_switch_ports`   | État des ports LAN : lien, vitesse (Mbps), duplex, MAC + hostname  |

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
