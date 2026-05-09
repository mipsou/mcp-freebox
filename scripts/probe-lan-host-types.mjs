// Liste les hosts du LAN avec leur primary_name + host_type + vendor
// pour identifier les classifications douteuses.

import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const TOKEN = process.env.FREEBOX_APP_TOKEN;
if (!TOKEN) { console.error('no token'); process.exit(1); }

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

await rpc('initialize', { protocolVersion: '2024-11-05', capabilities: {}, clientInfo: { name: 'probe-lan', version: '1.0' } });
child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized', params: {} }) + '\n');

const r = await rpc('tools/call', { name: 'freebox_lan_hosts', arguments: {} });
const hosts = JSON.parse(r.result.content[0].text);

// Tri par host_type pour repérer les patterns
const byType = {};
for (const h of hosts) {
  const t = h.host_type ?? 'unset';
  if (!byType[t]) byType[t] = [];
  byType[t].push(h);
}

console.log(`Total hosts: ${hosts.length}`);
console.log(`Reachable: ${hosts.filter(h => h.reachable).length}`);
console.log('');
for (const [t, hs] of Object.entries(byType).sort()) {
  console.log(`=== type=${t} (${hs.length}) ===`);
  for (const h of hs) {
    const macs = (h.l2ident ?? []).map(l => l.id).join(',');
    const ips = (h.l3connectivities ?? []).filter(c => c.active).map(c => c.addr).join(',');
    console.log(`  id=${h.id}`);
    console.log(`    name=${h.primary_name}  vendor=${h.vendor_name || '-'}  reachable=${h.reachable}  manual=${h.primary_name_manual}`);
    console.log(`    mac=${macs}  ip=${ips}`);
  }
  console.log('');
}

child.stdin.end();
child.kill();
process.exit(0);
