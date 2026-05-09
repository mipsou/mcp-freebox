// Test runtime du fix #89 (JSON tag) + #90 (vm_update enable_cloudinit).
// Cycle :
//   1. disk_create
//   2. vm_create avec cloudinit_userdata → vérifie enable_cloudinit=true dans réponse
//   3. vm_list → vérifie enable_cloudinit=true persisté côté API
//   4. vm_update enable_cloudinit=false → vérifie désactivation
//   5. vm_list → vérifie enable_cloudinit=false persisté
//   6. cleanup VM + disk + tasks

import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const TOKEN = process.env.FREEBOX_APP_TOKEN;
if (!TOKEN) { console.error('no token'); process.exit(1); }

const DISK_DIR = process.env.FREEBOX_TEST_DISK_DIR ?? '/Disque 1/VMs/';
const DISK_NAME = `test-vm-cloudinit-${Date.now()}.qcow2`;
const VM_NAME = 'mcp-test-vm-cloudinit';
const USERDATA = '#cloud-config\nhostname: test-cloudinit\n';

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

async function pollTask(taskId, maxSec = 60) {
  const start = Date.now();
  while (Date.now() - start < maxSec * 1000) {
    const t = await call('freebox_vm_disk_task', { task_id: taskId });
    if (t.error === true) throw new Error('task error=true');
    if (t.done === true) return t;
    await new Promise(r => setTimeout(r, 1500));
  }
  throw new Error(`task ${taskId} timeout`);
}

let vmId = null;
const taskIds = [];
let cleanupDone = false;

async function cleanup() {
  if (cleanupDone) return;
  cleanupDone = true;
  for (const tid of taskIds) {
    try { await call('freebox_vm_disk_task_delete', { task_id: tid }); } catch {}
  }
  if (vmId !== null) {
    try { await call('freebox_vm_delete', { id: vmId }); console.log(`  [cleanup] VM ${vmId} deleted`); } catch (e) { console.error(`  [cleanup] vm_delete: ${e.message}`); }
  }
  try {
    await call('freebox_fs_delete', { path: `${DISK_DIR}${DISK_NAME}` });
    console.log(`  [cleanup] disk deleted`);
  } catch (e) { console.error(`  [cleanup] fs_delete: ${e.message}`); }
}

async function main() {
  await rpc('initialize', { protocolVersion: '2024-11-05', capabilities: {}, clientInfo: { name: 'cloudinit-test', version: '1.0' } });
  child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized', params: {} }) + '\n');

  console.log('[1] disk_create 1 GiB');
  const ct = await call('freebox_vm_disk_create', { disk_name: DISK_NAME, disk_dir: DISK_DIR, size_gb: 1, disk_type: 'qcow2' });
  taskIds.push(ct.id);
  await pollTask(ct.id);
  console.log(`    OK task ${ct.id}`);

  console.log('[2] vm_create avec cloudinit_userdata');
  const vm = await call('freebox_vm_create', {
    name: VM_NAME, memory: 256, vcpus: 1,
    disk_name: DISK_NAME, disk_dir: DISK_DIR, disk_type: 'qcow2', os: 'debian',
    cloudinit_userdata: USERDATA,
  });
  vmId = vm.id;
  console.log(`    create response : id=${vmId} enable_cloudinit=${vm.enable_cloudinit} (raw response keys: ${Object.keys(vm).join(',')})`);
  if (vm.enable_cloudinit !== true) {
    throw new Error(`#89 not fixed: vm_create response shows enable_cloudinit=${vm.enable_cloudinit}, want true`);
  }

  console.log('[3] vm_list pour confirmer persistance côté API');
  const vms = await call('freebox_vm_list');
  const v = Array.isArray(vms) ? vms.find(x => x.id === vmId) : null;
  if (!v) throw new Error(`VM ${vmId} not in list`);
  console.log(`    list response : enable_cloudinit=${v.enable_cloudinit}`);
  if (v.enable_cloudinit !== true) {
    throw new Error(`#89 not fixed: vm_list shows enable_cloudinit=${v.enable_cloudinit}, want true`);
  }

  console.log('[4] vm_update enable_cloudinit=false');
  const upd = await call('freebox_vm_update', { id: vmId, enable_cloudinit: false });
  console.log(`    update response : enable_cloudinit=${upd.enable_cloudinit}`);
  if (upd.enable_cloudinit !== false) {
    throw new Error(`#90 not working: vm_update returned enable_cloudinit=${upd.enable_cloudinit}, want false`);
  }

  console.log('[5] vm_list pour confirmer désactivation');
  const vms2 = await call('freebox_vm_list');
  const v2 = vms2.find(x => x.id === vmId);
  console.log(`    list response : enable_cloudinit=${v2.enable_cloudinit}`);
  if (v2.enable_cloudinit !== false) {
    throw new Error(`#90 update did not persist: vm_list shows enable_cloudinit=${v2.enable_cloudinit}, want false`);
  }

  console.log('\nOK — #89 + #90 fix validated on real Freebox');
}

let exitCode = 0;
try { await main(); } catch (e) { console.error(`FAIL: ${e.message}`); exitCode = 1; }
finally { await cleanup(); child.stdin.end(); child.kill(); process.exit(exitCode); }
