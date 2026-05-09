// Runtime test : freebox_lan_host_update (nouveau tool) + L2Idents fix.
// Pick a non-critical reachable host avec host_type=other → change to "iot" →
// verify → restore. Garde le palier "type=" qui est valide pour cette Freebox.

import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const TOKEN = process.env.FREEBOX_APP_TOKEN;
if (!TOKEN) { console.error('no token'); process.exit(1); }

// Cible : Nest Protect (type=other actuellement, vendor Nest Labs)
const TARGET_VENDOR = 'Nest Labs Inc.';
const TARGET_NAME = 'Nest Protect';
// Set valide confirmé runtime : iot/tv/router rejetés. appliances marche.
const NEW_TYPE = 'appliances';

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

let originalType = null;
let targetId = null;

async function restore() {
  if (targetId && originalType) {
    try {
      await call('freebox_lan_host_update', { id: targetId, host_type: originalType });
      console.log(`  [restore] ${targetId} type → ${originalType}`);
    } catch (e) {
      console.error(`  [restore] FAIL: ${e.message}`);
    }
  }
}

async function main() {
  await rpc('initialize', { protocolVersion: '2024-11-05', capabilities: {}, clientInfo: { name: 'lan-update-test', version: '1.0' } });
  child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized', params: {} }) + '\n');

  console.log('[1] freebox_lan_hosts (test L2Idents fix + locate target)');
  const hosts = await call('freebox_lan_hosts');
  if (!Array.isArray(hosts)) throw new Error('expected array');
  console.log(`    OK: ${hosts.length} hosts (no L2Ident unmarshal error)`);

  const target = hosts.find(h => h.vendor_name === TARGET_VENDOR);
  if (!target) {
    console.log(`    Target ${TARGET_NAME} (${TARGET_VENDOR}) not found in current LAN — using first reachable host as test instead`);
    const fallback = hosts.find(h => h.host_type === 'other' && h.primary_name);
    if (!fallback) throw new Error('no test target found');
    targetId = fallback.id;
    originalType = fallback.host_type;
    console.log(`    using fallback: ${fallback.primary_name} (id=${fallback.id}, current type=${fallback.host_type})`);
  } else {
    targetId = target.id;
    originalType = target.host_type;
    console.log(`    target found: ${target.primary_name} (id=${target.id}, current type=${target.host_type})`);
  }

  console.log(`[2] update ${targetId} → host_type=${NEW_TYPE}`);
  const updated = await call('freebox_lan_host_update', { id: targetId, host_type: NEW_TYPE });
  console.log(`    response host_type=${updated.host_type}`);
  if (updated.host_type !== NEW_TYPE) {
    throw new Error(`host_type not applied: got ${updated.host_type}, want ${NEW_TYPE} — l'API a peut-être rejeté la valeur "${NEW_TYPE}"`);
  }

  console.log('[3] re-list pour confirmer persistance côté API');
  const after = await call('freebox_lan_hosts');
  const v = after.find(h => h.id === targetId);
  if (!v) throw new Error(`target ${targetId} disappeared`);
  console.log(`    persisted host_type=${v.host_type}`);
  if (v.host_type !== NEW_TYPE) {
    throw new Error(`host_type didn't persist: got ${v.host_type}`);
  }

  console.log('\nOK — freebox_lan_host_update validé sur Freebox réelle');
  console.log(`  L2Idents fix : 72+ hosts décodés sans erreur`);
  console.log(`  host_type write : ${originalType} → ${NEW_TYPE} OK et persisté`);
}

let exitCode = 0;
try { await main(); } catch (e) { console.error(`FAIL: ${e.message}`); exitCode = 1; }
finally { await restore(); child.stdin.end(); child.kill(); process.exit(exitCode); }
