# freebox-mcp

Serveur MCP (Model Context Protocol) pour la Freebox Delta via l'API Freebox OS v4.

## Prérequis

- Go 1.22+
- Une Freebox Delta (ou compatible API v4)
- Accès réseau à `mafreebox.freebox.fr` (ou IP LAN de la box)

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

Le pairing génère un `app_token` lié à cette application sur la Freebox :

```bash
./freebox-pair
```

1. Le programme contacte la Freebox
2. **Appuyer sur le bouton physique** de la Freebox quand demandé (LED clignote)
3. Le token s'affiche sur stdout — le sauvegarder immédiatement

```
freebox-pair: requesting app token...
freebox-pair: *** PRESS THE BUTTON ON YOUR FREEBOX NOW ***
freebox-pair: waiting for authorization (60s timeout)...
freebox-pair: SUCCESS
<app_token>
```

> Le token est une clé d'accès complète à l'API Freebox. Le stocker dans un gestionnaire de secrets (Windows Credential Manager, pass, 1Password...).

## Configuration

| Variable             | Défaut                    | Description                        |
|----------------------|---------------------------|------------------------------------|
| `FREEBOX_APP_TOKEN`  | *(obligatoire)*           | Token obtenu via `freebox-pair`    |
| `FREEBOX_HOST`       | `mafreebox.freebox.fr`    | Hostname ou IP de la Freebox       |
| `FREEBOX_APP_ID`     | `mcp-freebox`             | Identifiant applicatif             |

## Démarrage

```bash
export FREEBOX_APP_TOKEN="<token>"
./freebox-mcp
```

Le serveur lit stdin et écrit sur stdout (protocole MCP stdio).

## Intégration Claude Desktop

Dans `claude_desktop_config.json` :

```json
{
  "mcpServers": {
    "freebox": {
      "command": "/chemin/vers/freebox-mcp",
      "env": {
        "FREEBOX_APP_TOKEN": "<token>"
      }
    }
  }
}
```

## Outils disponibles

### Connexion WAN

| Outil                        | Description                                              |
|------------------------------|----------------------------------------------------------|
| `freebox_connection_status`  | État ligne, IP publique IPv4/IPv6, débits temps réel     |
| `freebox_connection_ftth`    | SFP présent, signal optique Tx/Rx (dBm×100), liaison     |
| `freebox_connection_xdsl`    | SNR, atténuation, FEC/CRC, uptime xDSL                  |
| `freebox_dyndns_list`        | Entrées DynDNS configurées et leur état                  |

### LAN

| Outil                  | Description                                                  |
|------------------------|--------------------------------------------------------------|
| `freebox_lan_hosts`    | Équipements du réseau local : nom, MAC, IP, accessibilité    |

### DHCP

| Outil                  | Description                                                  |
|------------------------|--------------------------------------------------------------|
| `freebox_dhcp_static`  | Réservations DHCP statiques (MAC → IP fixe)                  |
| `freebox_dhcp_leases`  | Baux DHCP dynamiques actifs (clients connectés)              |

### NAT / Port forwarding

| Outil                  | Description                                                  |
|------------------------|--------------------------------------------------------------|
| `freebox_nat_rules`    | Règles de redirection de ports NAT                           |

### WiFi

| Outil                  | Description                                                  |
|------------------------|--------------------------------------------------------------|
| `freebox_wifi_aps`     | Points d'accès WiFi : bande, canal, état, DFS                |
| `freebox_wifi_config`  | Configuration WiFi globale (activé, filtre MAC)              |

## Licence

[EUPL-1.2](LICENSE) — European Union Public Licence v1.2
