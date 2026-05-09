// Sonde runtime pour les issues #93 (fs_list bug) et #94 (fs_rename feature).
// 1. Tente fs_list sur un chemin valide → reproduit ou non l'erreur unmarshal ?
// 2. Liste les outils MCP : fs_rename existe-t-il déjà ?
// 3. Tente un appel direct /fs/rename/ via une création/rename pour valider la shape

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

// --- Probe #93 : fs_list ---
console.log('=== #93 probe: freebox_fs_list /Disque 1/VMs ===');
const listResp = await rpc('tools/call', { name: 'freebox_fs_list', arguments: { path: '/Disque 1/VMs' } });
const isErr = listResp.result?.isError === true;
const text = listResp.result?.content?.[0]?.text ?? '';
console.log(`  isError=${isErr}`);
console.log(`  content (first 400 chars):`, text.slice(0, 400));

// --- Probe #94 : list MCP tools to check if fs_rename exists ---
console.log('\n=== #94 probe: tools list (chercher fs_rename) ===');
const toolsResp = await rpc('tools/list', {});
const tools = toolsResp.result?.tools ?? [];
const fsTools = tools.filter(t => t.name.startsWith('freebox_fs_'));
console.log(`  outils fs_*: ${fsTools.map(t => t.name).join(', ')}`);
console.log(`  fs_rename existe ? ${fsTools.some(t => t.name === 'freebox_fs_rename')}`);

child.stdin.end();
child.kill();
process.exit(0);
