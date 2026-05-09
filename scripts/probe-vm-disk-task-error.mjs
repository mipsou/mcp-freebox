// Sonde la shape REELLE de VMDiskTask en provoquant une erreur volontaire :
// resize d'un disque inexistant. Imprime la reponse brute pour confirmer
// les noms de champs (error_message? error_msg? autre?).

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

await rpc('initialize', { protocolVersion: '2024-11-05', capabilities: {}, clientInfo: { name: 'probe', version: '1.0' } });
child.stdin.write(JSON.stringify({ jsonrpc: '2.0', method: 'notifications/initialized', params: {} }) + '\n');

// Créer une tâche disque sur un chemin invalide, puis lire la task pour voir
// la shape reelle (error_message? autre?).
console.log('[1] disk_create avec disk_dir invalide → erreur attendue');
const create = await rpc('tools/call', { name: 'freebox_vm_disk_create', arguments: {
  disk_name: 'probe-error.qcow2',
  disk_dir: '/Inexistant/Path/',
  size_gb: 1,
  disk_type: 'qcow2',
}});
console.log('  raw response:', JSON.stringify(create.result, null, 2));

// Si la task est créée malgré le path invalide, on poll pour voir la shape erreur
const text = create.result?.content?.[0]?.text ?? '';
let task;
try { task = JSON.parse(text); } catch { task = null; }
if (task?.id) {
  console.log(`[2] poll task ${task.id} pour voir la shape error finale`);
  for (let i = 0; i < 10; i++) {
    await new Promise(r => setTimeout(r, 1500));
    const st = await rpc('tools/call', { name: 'freebox_vm_disk_task', arguments: { task_id: task.id } });
    const tt = JSON.parse(st.result?.content?.[0]?.text ?? '{}');
    console.log(`  poll ${i}: ${JSON.stringify(tt)}`);
    if (tt.done || tt.error) {
      console.log('\n  === RAW TASK FIELDS ===');
      console.log(JSON.stringify(tt, null, 2));
      // Cleanup
      try { await rpc('tools/call', { name: 'freebox_vm_disk_task_delete', arguments: { task_id: task.id } }); } catch {}
      break;
    }
  }
}

child.stdin.end();
child.kill();
process.exit(0);
