// Test runtime de freebox_vm_update sur Freebox reelle (#80).
// Cycle: create test VM -> update -> verify via list -> delete VM + disk.
// Prerequis : FREEBOX_APP_TOKEN dans l'env, /Disque 1/VMs/ ecrivable.

import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const TOKEN = process.env.FREEBOX_APP_TOKEN;
if (!TOKEN) { console.error('no token'); process.exit(1); }

const DISK_DIR = process.env.FREEBOX_TEST_DISK_DIR ?? '/Disque 1/VMs/';
const DISK_NAME = `test-vm-update-${Date.now()}.qcow2`;
const VM_NAME_INITIAL = 'mcp-test-vm-update';
const VM_NAME_RENAMED = 'mcp-test-vm-update-renamed';

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
    setTimeout(() => {
      if (pending.has(id)) { pending.delete(id); reject(new Error('timeout')); }
    }, 30000);
  });
}

function notify(method, params) {
  child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method, params }) + '\n');
}

async function call(tool, args = {}) {
  const r = await rpc('tools/call', { name: tool, arguments: args });
  if (r.result?.isError) throw new Error(`${tool}: ${r.result.content?.[0]?.text}`);
  const text = r.result?.content?.[0]?.text ?? '';
  // Tools like freebox_vm_delete return plain text (not JSON) on success.
  try { return JSON.parse(text); } catch { return text; }
}

let vmId = null;
let cleanupDone = false;

async function cleanup() {
  if (cleanupDone) return;
  cleanupDone = true;
  if (vmId !== null) {
    try {
      await call('freebox_vm_delete', { id: vmId });
      console.log(`  [cleanup] VM ${vmId} deleted`);
      const vms = await call('freebox_vm_list');
      const stillThere = Array.isArray(vms) && vms.some(v => v.id === vmId);
      if (stillThere) console.error(`  [cleanup] WARN VM ${vmId} still in list`);
      else console.log(`  [cleanup] VM ${vmId} confirmed gone from list`);
    } catch (e) {
      console.error(`  [cleanup] vm_delete failed: ${e.message}`);
    }
  }
  try {
    const diskPath = `${DISK_DIR}${DISK_NAME}`;
    await call('freebox_fs_delete', { path: diskPath });
    console.log(`  [cleanup] disk ${diskPath} deleted`);
  } catch (e) {
    console.error(`  [cleanup] fs_delete failed: ${e.message}`);
  }
}

async function main() {
  await rpc('initialize', {
    protocolVersion: '2024-11-05',
    capabilities: {},
    clientInfo: { name: 'vm-update-test', version: '1.0' },
  });
  notify('notifications/initialized', {});

  console.log('[1] create test VM');
  const created = await call('freebox_vm_create', {
    name: VM_NAME_INITIAL,
    memory: 256,
    vcpus: 1,
    disk_name: DISK_NAME,
    disk_dir: DISK_DIR,
    disk_type: 'qcow2',
    os: 'debian',
  });
  vmId = created.id;
  console.log(`    VM created id=${vmId} name=${created.name} memory=${created.memory}`);

  console.log('[2] update VM (name + memory)');
  const updated = await call('freebox_vm_update', {
    id: vmId,
    name: VM_NAME_RENAMED,
    memory: 512,
  });
  console.log(`    update returned: name=${updated.name} memory=${updated.memory}`);

  console.log('[3] verify via list');
  const vms = await call('freebox_vm_list');
  const v = Array.isArray(vms) ? vms.find(x => x.id === vmId) : null;
  if (!v) throw new Error(`VM ${vmId} not found in list after update`);
  console.log(`    list shows: name=${v.name} memory=${v.memory}`);
  if (v.name !== VM_NAME_RENAMED) throw new Error(`name not updated: got ${v.name}, want ${VM_NAME_RENAMED}`);
  if (v.memory !== 512) throw new Error(`memory not updated: got ${v.memory}, want 512`);

  console.log('[4] update only enable_screen=false (regression test for *bool)');
  await call('freebox_vm_update', { id: vmId, enable_screen: false });

  console.log('\nOK — vm_update works on real Freebox');
}

let exitCode = 0;
try {
  await main();
} catch (e) {
  console.error(`FAIL: ${e.message}`);
  exitCode = 1;
} finally {
  await cleanup();
  child.stdin.end();
  child.kill();
  process.exit(exitCode);
}
