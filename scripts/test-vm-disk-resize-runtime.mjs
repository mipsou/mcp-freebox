// Test runtime de freebox_vm_disk_create + freebox_vm_disk_resize (#85).
// Cycle: disk_create (1 GiB) -> poll -> resize (2 GiB) -> poll -> fs_delete.
// Pas besoin de creer une VM : on teste les disk operations directement.
// Prerequis : FREEBOX_APP_TOKEN dans l'env, /Disque 1/VMs/ ecrivable.

import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const TOKEN = process.env.FREEBOX_APP_TOKEN;
if (!TOKEN) { console.error('no token'); process.exit(1); }

const DISK_DIR = process.env.FREEBOX_TEST_DISK_DIR ?? '/Disque 1/VMs/';
const DISK_NAME = `test-vm-disk-resize-${Date.now()}.qcow2`;
const VM_NAME = 'mcp-test-vm-disk-resize';
const INITIAL_GB = 1;
const TARGET_GB = 2;

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
    }, 60000);
  });
}

function notify(method, params) {
  child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method, params }) + '\n');
}

async function call(tool, args = {}) {
  const r = await rpc('tools/call', { name: tool, arguments: args });
  if (r.result?.isError) throw new Error(`${tool}: ${r.result.content?.[0]?.text}`);
  const text = r.result?.content?.[0]?.text ?? '';
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

async function pollTask(taskId, maxSec = 120) {
  const start = Date.now();
  while (Date.now() - start < maxSec * 1000) {
    const t = await call('freebox_vm_disk_task', { task_id: taskId });
    console.log(`    task ${taskId} state=${t.state} progress=${t.progress}% done=${t.done} error=${t.error}`);
    if (t.error === true) throw new Error(`task failed: ${t.error_message ?? 'no message'}`);
    if (t.done === true || t.state === 'done') return t;
    if (t.state === 'failed') throw new Error(`task failed: ${t.error_message ?? t.error}`);
    await new Promise(r => setTimeout(r, 2000));
  }
  throw new Error(`task ${taskId} did not complete within ${maxSec}s`);
}

async function main() {
  await rpc('initialize', {
    protocolVersion: '2024-11-05',
    capabilities: {},
    clientInfo: { name: 'vm-disk-resize-test', version: '1.0' },
  });
  notify('notifications/initialized', {});

  console.log(`[1] disk_create blank qcow2 (${INITIAL_GB} GiB) at ${DISK_DIR}${DISK_NAME}`);
  const createTask = await call('freebox_vm_disk_create', {
    disk_name: DISK_NAME,
    disk_dir: DISK_DIR,
    size_gb: INITIAL_GB,
    disk_type: 'qcow2',
  });
  console.log(`    create task: id=${createTask.id} type=${createTask.type} state=${createTask.state}`);
  if (typeof createTask.id !== 'number') throw new Error(`createTask.id not a number: ${JSON.stringify(createTask)}`);

  console.log(`[2] poll create task ${createTask.id} until done`);
  const createFinal = await pollTask(createTask.id, 60);
  console.log(`    final: done=${createFinal.done} error=${createFinal.error}`);
  if (createFinal.error) throw new Error(`disk_create failed: ${createFinal.error_message}`);

  console.log('[3] verify disk file exists');
  const info = await call('freebox_fs_info', { path: `${DISK_DIR}${DISK_NAME}` });
  console.log(`    fs_info: type=${info.type} size=${info.size} mimetype=${info.mimetype}`);

  console.log(`[4] vm_create attache au disque pre-existant`);
  const vm = await call('freebox_vm_create', {
    name: VM_NAME,
    memory: 256,
    vcpus: 1,
    disk_name: DISK_NAME,
    disk_dir: DISK_DIR,
    disk_type: 'qcow2',
    os: 'debian',
  });
  vmId = vm.id;
  console.log(`    VM created id=${vmId} disk_path=${vm.disk_path}`);

  console.log(`[5] disk_resize → ${TARGET_GB} GiB`);
  const resizeTask = await call('freebox_vm_disk_resize', {
    id: vmId, size_gb: TARGET_GB, allow_shrink: false,
  });
  console.log(`    resize task: id=${resizeTask.id} type=${resizeTask.type} state=${resizeTask.state}`);

  console.log(`[6] poll resize task ${resizeTask.id} until done`);
  const resizeFinal = await pollTask(resizeTask.id, 120);
  console.log(`    final: done=${resizeFinal.done} error=${resizeFinal.error}`);
  if (resizeFinal.error) throw new Error(`disk_resize failed: ${resizeFinal.error_message}`);

  console.log('[7] verify disk size grew');
  const infoAfter = await call('freebox_fs_info', { path: `${DISK_DIR}${DISK_NAME}` });
  console.log(`    fs_info after resize: size=${infoAfter.size}`);
  if (infoAfter.size <= info.size) {
    console.warn(`    WARN: size_before=${info.size}, size_after=${infoAfter.size} — qcow2 sparse, peut ne pas refléter la taille virtuelle`);
  }

  console.log('\nOK — vm_disk_create + vm_disk_resize + vm_disk_task work on real Freebox');
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
