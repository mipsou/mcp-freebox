// Probing direct des endpoints Freebox via le token MCP existant.
// Auth flow : GET /api/v15/login/ → challenge → HMAC-SHA1(token, challenge) → POST /login/session/ → session_token
// Puis GET sur chaque path candidat avec header X-Fbx-App-Auth.
//
// Prerequis : variable d'env FREEBOX_APP_TOKEN (cf. scripts/load-token.ps1)

import { createHmac } from 'node:crypto';
import { writeFileSync } from 'node:fs';
import { resolve } from 'node:path';

const BASE = process.env.FREEBOX_API_BASE ?? 'http://mafreebox.freebox.fr/api/v15';
const APP_ID = process.env.FREEBOX_APP_ID ?? 'mcp-freebox';
const TOKEN = process.env.FREEBOX_APP_TOKEN;
const OUT = resolve(process.cwd(), 'probe-output.json');
const CONCURRENCY = parseInt(process.env.PROBE_CONCURRENCY ?? '4', 10);

if (!TOKEN) {
  console.error('FREEBOX_APP_TOKEN manquant. Source scripts/load-token.ps1 d\'abord.');
  process.exit(1);
}

async function getJson(url, headers = {}) {
  const r = await fetch(url, { headers });
  const text = await r.text();
  let json;
  try { json = JSON.parse(text); } catch { json = null; }
  return { status: r.status, json, text };
}

async function postJson(url, body, headers = {}) {
  const r = await fetch(url, {
    method: 'POST',
    headers: { 'content-type': 'application/json', ...headers },
    body: typeof body === 'string' ? body : JSON.stringify(body),
  });
  const text = await r.text();
  let json;
  try { json = JSON.parse(text); } catch { json = null; }
  return { status: r.status, json, text };
}

console.error(`[probe] auth flow on ${BASE}`);
const ch = await getJson(`${BASE}/login/`);
if (!ch.json?.success) { console.error('login fail', ch.text); process.exit(2); }
const challenge = ch.json.result.challenge;

const password = createHmac('sha1', TOKEN).update(challenge).digest('hex');
const sess = await postJson(`${BASE}/login/session/`, { app_id: APP_ID, password });
if (!sess.json?.success) { console.error('session fail', sess.text); process.exit(3); }
const sessionToken = sess.json.result.session_token;
const perms = sess.json.result.permissions;
console.error(`[probe] session ok, perms = ${JSON.stringify(perms)}`);

const H = { 'X-Fbx-App-Auth': sessionToken };

// Paths candidats : SDK + extrapolations. Format = '/path/' (trailing slash quand collection).
// On groupe par theme pour lisibilite du rapport.
const CANDIDATES = {
  system: [
    '/system/', '/system/sensors/', '/system/temperature/', '/system/uptime/',
    '/system/reboot/', '/system/update/', '/system/info/', '/system/config/',
    '/freeplug/', '/freeplug/config/', '/freeplug/networks/',
  ],
  network: [
    '/network/', '/network/interfaces/', '/network/lan/',
    '/lan/config/', '/lan/browser/pub/', '/lan/browser/interfaces/', '/lan/wol/',
    '/connection/', '/connection/config/', '/connection/xdsl/', '/connection/ftth/',
    '/connection/lte/', '/connection/dsl/', '/connection/logs/', '/connection/ipv6/config/',
    '/connection/ipv6/delegations/', '/connection/dyndns/', '/connection/dyndns/types/',
    '/lldp/list/', '/lldp/config/', '/neighbors/', '/neighbour/',
  ],
  dhcp: [
    '/dhcp/config/', '/dhcp/dynamic_lease/', '/dhcp/static_lease/',
    '/dhcp/option/', '/dhcp/options/',
    '/dhcpv6/config/', '/dhcpv6/lease/',
    '/tftp/config/',
  ],
  wifi: [
    '/wifi/config/', '/wifi/ap/', '/wifi/bss/', '/wifi/stations/', '/wifi/mac_filter/',
    '/wifi/diag/', '/wifi/diagnostic/', '/wifi/planning/', '/wifi/wps/', '/wifi/wps/sessions/',
    '/wifi/custom_key/', '/wifi/dect/',
  ],
  firewall: [
    '/firewall/', '/firewall/config/',
    '/fw/redir/', '/fw/dmz/', '/fw/incoming/',
    '/upnp/config/', '/upnp/igd/rules/', '/upnpav/config/',
  ],
  vm: [
    '/vm/', '/vm/info', '/vm/info/', '/vm/distros/', '/vm/distros',
    '/vm/usb_ports/', '/vm/disks/', '/vm/installer_log/',
  ],
  storage: [
    '/storage/disk/', '/storage/partition/', '/storage/volume/', '/storage/raid/',
    '/storage/lvm/', '/storage/btrfs/',
  ],
  fs: [
    '/fs/ls/', '/fs/tasks/', '/fs/info/', '/fs/share_link/',
  ],
  share: [
    '/share/', '/share/samba/', '/share/afp/', '/share/auto/',
    '/netshare/samba/', '/netshare/samba/share/', '/netshare/afp/',
  ],
  ftp: [
    '/ftp/config/',
  ],
  downloads: [
    '/downloads/', '/downloads/config/', '/downloads/feeds/',
    '/download_feeds/', '/download_feeds/items/',
  ],
  airmedia: [
    '/airmedia/config/', '/airmedia/receivers/',
  ],
  contacts: [
    '/contact/', '/contact/groups/', '/contact/types/',
  ],
  calls: [
    '/call/log/', '/call/types/',
  ],
  parental: [
    '/parental/config/', '/parental/filter/', '/parental/planning/',
  ],
  vpn: [
    '/vpn/', '/vpn/connection/', '/vpn/config/', '/vpn/server/', '/vpn/user/',
    '/vpn_client/config/', '/vpn_client/connection/',
    '/wireguard/', '/wireguard/server/', '/wireguard/client/',
  ],
  hotspot: [
    '/hotspot/', '/hotspot/config/', '/hotspot/clients/', '/hotspot/quotas/',
  ],
  switch: [
    '/switch/status/', '/switch/port/', '/switch/port/1/', '/switch/port/1/stats/',
    '/switch/config/', '/switch/port/1/info/',
  ],
  lcd: [
    '/lcd/config/', '/screen/config/',
  ],
  pvr: [
    '/pvr/programmed/', '/pvr/finished/', '/pvr/media/', '/pvr/quota/',
  ],
  tv: [
    '/tv/channels/', '/tv/bouquets/', '/tv/airwave/',
  ],
  notif: [
    '/notif/', '/notification/', '/notification/targets/',
  ],
  rrd: [
    '/rrd/', // body POST normalement, GET retourne souvent une liste
  ],
  voip: [
    '/sip/', '/sip/config/', '/sip/lines/', '/sip/log/',
    '/dect/', '/dect/config/', '/dect/handsets/',
  ],
  misc: [
    '/profile/', '/sysaction/', '/sysaction/types/',
    '/login/authorize/', '/login/sessions/',
    '/igd/', '/igd/config/',
  ],
};

const allPaths = Object.entries(CANDIDATES).flatMap(([cat, ps]) => ps.map(p => ({ cat, p })));
console.error(`[probe] testing ${allPaths.length} paths (concurrency=${CONCURRENCY})...`);

async function probe({ cat, p }) {
  const url = `${BASE}${p}?_dc=${Date.now()}`;
  try {
    const { status, json } = await getJson(url, H);
    let resultKind = 'unknown';
    let topFields = null;
    if (json?.success === true) {
      resultKind = 'success';
      const r = json.result;
      if (r && typeof r === 'object' && !Array.isArray(r)) {
        topFields = Object.keys(r).sort();
      } else if (Array.isArray(r)) {
        topFields = r.length > 0 && typeof r[0] === 'object' ? ['[]', ...Object.keys(r[0]).sort()] : ['[]'];
      } else {
        topFields = [typeof r];
      }
    } else if (json?.success === false) {
      resultKind = json.error_code ?? 'error';
    } else {
      resultKind = `http_${status}`;
    }
    return { cat, path: p, status, resultKind, topFields, msg: json?.msg ?? null };
  } catch (e) {
    return { cat, path: p, status: 0, resultKind: 'fetch_error', error: e.message };
  }
}

// Pool simple
const results = [];
let idx = 0;
async function worker() {
  while (idx < allPaths.length) {
    const i = idx++;
    const r = await probe(allPaths[i]);
    results.push(r);
    const tag = r.resultKind === 'success' ? '✓' : (r.resultKind === 'http_404' ? '·' : '✗');
    console.error(`[${tag}] ${r.path} → ${r.resultKind}${r.topFields ? ' [' + r.topFields.slice(0, 6).join(',') + (r.topFields.length > 6 ? ',...' : '') + ']' : ''}`);
  }
}

await Promise.all(Array.from({ length: CONCURRENCY }, worker));

// Tri par categorie puis path pour lisibilite
results.sort((a, b) => a.cat.localeCompare(b.cat) || a.path.localeCompare(b.path));

const summary = {
  total: results.length,
  success: results.filter(r => r.resultKind === 'success').length,
  not_found: results.filter(r => r.resultKind === 'http_404').length,
  forbidden: results.filter(r => r.resultKind === 'insufficient_rights' || r.resultKind === 'auth_required').length,
  other_errors: results.filter(r => !['success', 'http_404', 'insufficient_rights', 'auth_required'].includes(r.resultKind)).length,
};
console.error(`\n[summary] ${JSON.stringify(summary)}`);

writeFileSync(OUT, JSON.stringify({ summary, base: BASE, results }, null, 2));
console.error(`[probe] wrote → ${OUT}`);
