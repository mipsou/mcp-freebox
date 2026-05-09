// Runtime test des fixes #93 (fs_list) et #94 (fs_rename) sur Freebox reelle.
// Cycle :
//   1. fs_list /Disque 1/VMs → ne doit plus erreur "unmarshal object"
//   2. disk_create test-rename-old.qcow2
//   3. fs_rename → test-rename-new.qcow2
//   4. fs_list confirme la presence du nouveau nom et l'absence de l'ancien
//   5. cleanup

import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const TOKEN = process.env.FREEBOX_APP_TOKEN;
if (!TOKEN) { console.error('no token'); process.exit(1); }

const DISK_DIR = '/Disque 1/VMs/';
const OLD_NAME = `test-rename-old-${Date.now()}.qcow2`;
const NEW_NAME = `test-rename-new-${Date.now()}.qcow2`;

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

async function pollDiskTask(taskId, maxSec = 60) {
  const start = Date.now();
  while (Date.now() - start < maxSec * 1000) {
    const t = await call('freebox_vm_disk_task', { task_id: taskId });
    if (t.error === true) throw new Error('task error');
    if (t.done === true) return t;
    await new Promise(r => setTimeout(r, 1500));
  }
  throw new Error(`task timeout`);
}

const taskIds = [];
let cleanupDone = false;

async function cleanup() {
  if (cleanupDone) return;
  cleanupDone = true;
  for (const tid of taskIds) {
    try { await call('freebox_vm_disk_task_delete', { task_id: tid }); } catch {}
  }
  // Try delete both names (whichever exists)
  for (const name of [OLD_NAME, NEW_NAME]) {
    try {
      await call('freebox_fs_delete', { path: `${DISK_DIR}${name}` });
      console.log(`  [cleanup] disk ${name} deleted`);
    } catch {}
  }
}

async function main() {
  await rpc('initialize', { protocolVersion: '2024-11-05', capabilities: {}, clientInfo: { name: 'fs-test', version: '1.0' } });
  child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized', params: {} }) + '\n');

  console.log('=== #93 : fs_list /Disque 1/VMs ===');
  const entries = await call('freebox_fs_list', { path: '/Disque 1/VMs' });
  if (!Array.isArray(entries)) throw new Error(`fs_list returned non-array: ${typeof entries}`);
  console.log(`  OK fs_list → ${entries.length} entries (no unmarshal error)`);

  console.log('\n=== #94 : create disk → rename → verify ===');
  console.log(`[1] disk_create ${OLD_NAME}`);
  const ct = await call('freebox_vm_disk_create', { disk_name: OLD_NAME, disk_dir: DISK_DIR, size_gb: 1, disk_type: 'qcow2' });
  taskIds.push(ct.id);
  await pollDiskTask(ct.id);
  console.log(`    OK created`);

  console.log(`[2] fs_rename ${OLD_NAME} → ${NEW_NAME}`);
  const newPath = await call('freebox_fs_rename', { src_path: `${DISK_DIR}${OLD_NAME}`, new_name: NEW_NAME });
  console.log(`    new path returned: ${newPath}`);
  if (typeof newPath !== 'string') throw new Error(`fs_rename returned non-string: ${JSON.stringify(newPath)}`);

  console.log(`[3] fs_list pour vérifier rename effectif`);
  const after = await call('freebox_fs_list', { path: '/Disque 1/VMs' });
  const hasNew = after.some(e => e.name === NEW_NAME);
  const hasOld = after.some(e => e.name === OLD_NAME);
  console.log(`    found new=${hasNew} old=${hasOld}`);
  if (!hasNew) throw new Error(`new name ${NEW_NAME} not found after rename`);
  if (hasOld) throw new Error(`old name ${OLD_NAME} still present after rename`);

  console.log('\nOK — #93 (fs_list shape) + #94 (fs_rename) validated on real Freebox');
}

let exitCode = 0;
try { await main(); } catch (e) { console.error(`FAIL: ${e.message}`); exitCode = 1; }
finally { await cleanup(); child.stdin.end(); child.kill(); process.exit(exitCode); }
