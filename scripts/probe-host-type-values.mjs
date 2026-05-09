// Sonde les valeurs valides de host_type en tentant plusieurs sur un host non-critique.
// Restauration garantie en cas d'échec.

import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const TOKEN = process.env.FREEBOX_APP_TOKEN;
if (!TOKEN) { console.error('no token'); process.exit(1); }

const TARGET_VENDOR = 'Nest Labs Inc.';
// Brute-force plus large : doc v4 + extensions hacf-fr + candidats domotique
// usuels pour découvrir les ajouts non documentés de l'API v15.
const CANDIDATES = [
  // doc v4 dev.freebox.fr (15)
  'workstation', 'laptop', 'smartphone', 'tablet', 'printer',
  'vg_console', 'television', 'nas', 'ip_camera', 'ip_phone',
  'freebox_player', 'freebox_hd', 'networking_device',
  'multimedia_device', 'other',
  // déjà sondé OK
  'freebox_pop', 'appliances',
  // hacf-fr extras
  'freebox_delta', 'freebox_mini', 'freebox_revolution', 'freebox_one',
  'freebox_server',
  // candidates domotiques smart-home
  'voice_assistant', 'smart_speaker', 'google_home', 'amazon_echo',
  'thermostat', 'light', 'lighting', 'smart_light',
  'door_bell', 'doorbell', 'lock', 'smart_lock',
  'robot', 'robot_vacuum', 'vacuum_cleaner', 'vacuum',
  'camera', 'security_camera', 'webcam',
  'weather_station', 'sensor', 'iot_sensor', 'smart_sensor',
  'wearable', 'watch', 'smartwatch', 'health_device',
  'hub', 'smart_hub', 'gateway', 'bridge',
  'shutter', 'shutters', 'switch', 'plug', 'smart_plug', 'outlet',
  'wallbox', 'ev_charger',
  'solar', 'inverter',
  'audio', 'speaker', 'video', 'streaming_device',
  'fbx', 'home', 'connected', 'iot',
  // alias possibles
  'pc', 'phone', 'mobile', 'console', 'tv', 'router', 'access_point',
  'firebox',
  // user push : 3 types absents — sonder leurs noms exacts
  'freebox_crystal', 'freebox_wifi_pop', 'freebox_pop_wifi',
  'connected_vehicle', 'vehicle', 'car', 'connected_car',
  // Wi-Fi Pop extender variants
  'wifi_pop', 'pop_extender', 'freebox_extender', 'extender',
  'repeater', 'wifi_repeater', 'mesh', 'mesh_node', 'fbx_wifi_pop',
  'fbx_pop_wifi', 'freebox_repeater', 'freebox_mesh',
  'pop_wifi', 'wifipop', 'popwifi', 'freebox_companion',
  'pop_companion', 'freebox_pop_companion', 'freebox_pop_extender',
  'freebox_pop_v2', 'freebox_pop_2', 'freebox_pop_wifi_extender',
  'pop_amplifier', 'freebox_amplifier', 'wifi_amplifier',
  'freebox_satellite', 'satellite', 'companion',
  // domotique sonore éventuelle
  'speaker_voice',
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
  return new Promise((resolve) => {
    pending.set(id, { resolve });
    child.stdin.write(JSON.stringify({ jsonrpc: '2.0', id, method, params }) + '\n');
    setTimeout(() => { if (pending.has(id)) { pending.delete(id); resolve({ result: { isError: true, content: [{ text: 'timeout' }] } }); } }, 15000);
  });
}

async function call(tool, args = {}) {
  const r = await rpc('tools/call', { name: tool, arguments: args });
  if (r.result?.isError) {
    return { error: r.result.content?.[0]?.text };
  }
  const text = r.result?.content?.[0]?.text ?? '';
  try { return { value: JSON.parse(text) }; } catch { return { value: text }; }
}

await rpc('initialize', { protocolVersion: '2024-11-05', capabilities: {}, clientInfo: { name: 'probe-types', version: '1.0' } });
child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized', params: {} }) + '\n');

console.log('[1] locate target host');
const hostsResp = await call('freebox_lan_hosts');
if (hostsResp.error) { console.error(hostsResp.error); process.exit(1); }
const target = hostsResp.value.find(h => h.vendor_name === TARGET_VENDOR);
if (!target) { console.error('target not found'); process.exit(1); }
const TID = target.id;
const ORIG = target.host_type;
console.log(`    target: ${target.primary_name} (id=${TID}, original type=${ORIG})`);

console.log('\n[2] probe candidates');
const valid = [];
const invalid = [];
for (const t of CANDIDATES) {
  if (t === ORIG) continue;
  const r = await call('freebox_lan_host_update', { id: TID, host_type: t });
  if (r.error) {
    invalid.push({ t, err: r.error.split('\n')[0].slice(0, 80) });
    process.stdout.write(`  ${t}: REJECTED (${r.error.slice(0, 60)})\n`);
  } else {
    valid.push(t);
    process.stdout.write(`  ${t}: OK\n`);
  }
}

console.log('\n[3] restore original type');
await call('freebox_lan_host_update', { id: TID, host_type: ORIG });

console.log('\n=== summary ===');
console.log(`valid host_type values (${valid.length}):`, valid.join(', '));
console.log(`invalid (${invalid.length}):`, invalid.map(i => i.t).join(', '));

child.stdin.end();
child.kill();
process.exit(0);
