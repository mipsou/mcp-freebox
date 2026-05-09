// Applique les corrections de host_type validees avec l'utilisateur.
// Match les hosts par MAC (l2ident.id) puisque c'est stable, contrairement
// au primary_name qui peut etre l'auto-detected.

import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const TOKEN = process.env.FREEBOX_APP_TOKEN;
if (!TOKEN) { console.error('no token'); process.exit(1); }

// Changes : { mac (uppercase) | name match : new_type }
// Si MAC commence par 'mac:' on match exact mac; sinon on match par primary_name
const CHANGES = [
  // → nas : serveurs / services 24/7
  { match: { mac: 'D0:50:99:FD:24:31' }, new_type: 'nas',                desc: 'xigmanas (.77)' },
  { match: { mac: 'E8:4E:06:32:38:9F' }, new_type: 'nas',                desc: 'Maison (Home Assistant)' },
  { match: { mac: '20:F8:3B:00:70:D9' }, new_type: 'nas',                desc: 'homeassistant (Nabu Casa)' },
  { match: { mac: 'D6:C0:B8:EC:40:43' }, new_type: 'nas',                desc: 'jeedom (serveur domotique)' },
  // → networking_device : passerelles
  { match: { mac: 'C8:2E:18:52:67:78' }, new_type: 'networking_device', desc: 'SLZB-06M (Zigbee, .121)' },
  { match: { mac: 'C8:2E:18:52:67:7B' }, new_type: 'networking_device', desc: 'SLZB-06M (Zigbee, alt)' },
  { match: { mac: '68:EC:8A:00:CE:91' }, new_type: 'networking_device', desc: 'Ikia gw2 (Trådfri)' },
  { match: { mac: '00:17:88:2B:45:41' }, new_type: 'networking_device', desc: 'Hue Gw (Philips Hue Bridge)' },
  // → appliances : électroménager / smart-home fixe
  { match: { mac: '50:14:79:12:0C:AC' }, new_type: 'appliances',        desc: 'iRobot-Lavo (robot ménager)' },
  { match: { mac: '18:B4:30:35:D8:33' }, new_type: 'appliances',        desc: 'Nest Protect (détecteur fumée)' },
  { match: { mac: '70:EE:50:1D:F5:C0' }, new_type: 'appliances',        desc: 'Netatmo Weather Station' },
  { match: { mac: '00:24:E4:1C:23:C8' }, new_type: 'appliances',        desc: 'Withings Balance' },
  // → multimedia_device : streaming/audio/vidéo connecté
  { match: { mac: '3C:6A:9D:15:48:A2' }, new_type: 'multimedia_device', desc: 'Elgato Key Light Air Droite (streaming)' },
  { match: { mac: '3C:6A:9D:16:12:E1' }, new_type: 'multimedia_device', desc: 'Elgato Key Light Air Gauche (streaming)' },
];

const BIN = resolve(process.cwd(), 'freebox-mcp.exe');
const child = spawn(BIN, [], {
  env: { ...process.env, FREEBOX_APP_TOKEN: TOKEN, FREEBOX_APP_ID: process.env.FREEBOX_APP_ID ?? 'mcp-freebox' },
  stdio: ['pipe', 'pipe', 'pipe'],
});

let buf = '';
const pending = new Map();
let nextId = 1;

child.stdout.on('data', d => {
  buf += d.toString();
  let idx;
  while ((idx = buf.indexOf('\n')) !== -1) {
    const line = buf.slice(0, idx).trim();
    buf = buf.slice(idx + 1);
    if (!line) continue;
    try {
      const msg = JSON.parse(line);
      if (msg.id != null && pending.has(msg.id)) {
        const { resolve: r } = pending.get(msg.id);
        pending.delete(msg.id);
        r(msg);
      }
    } catch {}
  }
});
child.stderr.on('data', d => process.stderr.write('[mcp.stderr] ' + d.toString()));

function rpc(method, params) {
  const id = nextId++;
  return new Promise((resolve, reject) => {
    pending.set(id, { resolve });
    child.stdin.write(JSON.stringify({ jsonrpc: '2.0', id, method, params }) + '\n');
    setTimeout(() => { if (pending.has(id)) { pending.delete(id); reject(new Error('timeout')); } }, 30000);
  });
}

async function call(tool, args = {}) {
  const r = await rpc('tools/call', { name: tool, arguments: args });
  if (r.result?.isError) throw new Error(`${tool}: ${r.result.content?.[0]?.text}`);
  const text = r.result?.content?.[0]?.text ?? '';
  try { return JSON.parse(text); } catch { return text; }
}

await rpc('initialize', { protocolVersion: '2024-11-05', capabilities: {}, clientInfo: { name: 'apply-types', version: '1.0' } });
child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized', params: {} }) + '\n');

console.log('[1] fetch current LAN hosts');
const hosts = await call('freebox_lan_hosts');
console.log(`    ${hosts.length} hosts loaded`);

console.log('\n[2] apply changes');
let applied = 0, skipped = 0, failed = 0;
for (const ch of CHANGES) {
  const target = hosts.find(h => {
    if (ch.match.mac) {
      return (h.l2ident ?? []).some(l => l.id === ch.match.mac);
    }
    return h.primary_name === ch.match.name;
  });
  if (!target) {
    console.log(`  SKIP ${ch.desc}: not found in current LAN list`);
    skipped++;
    continue;
  }
  if (target.host_type === ch.new_type) {
    console.log(`  SKIP ${ch.desc}: already type=${ch.new_type}`);
    skipped++;
    continue;
  }
  try {
    await call('freebox_lan_host_update', { id: target.id, host_type: ch.new_type });
    console.log(`  OK   ${ch.desc} (id=${target.id}): ${target.host_type} → ${ch.new_type}`);
    applied++;
  } catch (e) {
    console.error(`  FAIL ${ch.desc}: ${e.message}`);
    failed++;
  }
}

console.log(`\n=== summary ===`);
console.log(`applied: ${applied}, skipped: ${skipped}, failed: ${failed}`);

child.stdin.end();
child.kill();
process.exit(failed === 0 ? 0 : 1);
