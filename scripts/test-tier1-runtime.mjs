// Test runtime des outils MCP tier 1 contre la Freebox reelle.
// Lance freebox-mcp.exe en stdio, envoie initialize + tools/call pour
// chaque outil tier 1, valide que la reponse n'est pas une erreur.
//
// Prerequis : FREEBOX_APP_TOKEN dans l'env.

import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const TOKEN = process.env.FREEBOX_APP_TOKEN;
const APP_ID = process.env.FREEBOX_APP_ID ?? 'mcp-freebox';
if (!TOKEN) { console.error('no token'); process.exit(1); }

const BIN = resolve(process.cwd(), 'freebox-mcp.exe');
const child = spawn(BIN, [], {
  env: { ...process.env, FREEBOX_APP_TOKEN: TOKEN, FREEBOX_APP_ID: APP_ID },
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
    } catch {
      // ignore non-JSON
    }
  }
});

child.stderr.on('data', d => process.stderr.write('[mcp.stderr] ' + d.toString()));
child.on('error', e => { console.error('spawn error:', e); process.exit(2); });

function rpc(method, params) {
  const id = nextId++;
  return new Promise((resolve, reject) => {
    pending.set(id, { resolve });
    const msg = JSON.stringify({ jsonrpc: '2.0', id, method, params }) + '\n';
    child.stdin.write(msg);
    setTimeout(() => {
      if (pending.has(id)) { pending.delete(id); reject(new Error('timeout')); }
    }, 15000);
  });
}

function notify(method, params) {
  const msg = JSON.stringify({ jsonrpc: '2.0', method, params }) + '\n';
  child.stdin.write(msg);
}

// Initialize
const init = await rpc('initialize', {
  protocolVersion: '2024-11-05',
  capabilities: {},
  clientInfo: { name: 'tier1-test', version: '1.0' },
});
console.log('[init]', init.result?.serverInfo?.name, init.result?.serverInfo?.version);
notify('notifications/initialized', {});

// Cas de test : nom outil + args eventuels + extrait attendu dans la sortie
const cases = [
  { tool: 'freebox_connection_logs', args: {}, expect: 'date' },
  { tool: 'freebox_connection_ipv6_config', args: {}, expect: 'ipv6_enabled' },
  { tool: 'freebox_lan_interfaces', args: {}, expect: 'host_count' },
  { tool: 'freebox_wifi_mac_filter', args: {}, expect: 'mac' },
  { tool: 'freebox_wifi_planning', args: {}, expect: 'use_planning' },
  { tool: 'freebox_storage_raid', args: {}, expect: '[' }, // [] ou liste
  { tool: 'freebox_fs_info', args: { path: '/Disque dur' }, expect: 'mimetype' },
  { tool: 'freebox_vm_info', args: {}, expect: 'total_cpus' },
  { tool: 'freebox_vm_distros', args: {}, expect: 'os' },
  { tool: 'freebox_switch_port_config', args: { id: 1 }, expect: 'duplex' },
  { tool: 'freebox_switch_port_stats', args: { id: 1 }, expect: 'rx_good_bytes' }, // nouveau field
];

let pass = 0, fail = 0;
for (const c of cases) {
  try {
    const r = await rpc('tools/call', { name: c.tool, arguments: c.args });
    const content = r.result?.content?.[0]?.text ?? '';
    const isError = r.result?.isError === true;
    const ok = !isError && content.includes(c.expect);
    console.log(`[${ok ? 'OK' : 'FAIL'}] ${c.tool} → ${isError ? 'ERROR' : 'success'}, contains "${c.expect}"=${content.includes(c.expect)}`);
    if (!ok) {
      console.log('  preview:', content.slice(0, 200));
      fail++;
    } else {
      pass++;
    }
  } catch (e) {
    console.log(`[ERR ] ${c.tool} → ${e.message}`);
    fail++;
  }
}

console.log(`\n=== ${pass} passed, ${fail} failed ===`);
child.stdin.end();
child.kill();
process.exit(fail === 0 ? 0 : 1);
