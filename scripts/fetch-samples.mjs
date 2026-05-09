// Recupere les payloads JSON complets des endpoints tier 1 pour generer les structs Go.
// Ecrit samples-tier1.json (un map path -> response.result).

import { createHmac } from 'node:crypto';
import { writeFileSync } from 'node:fs';
import { resolve } from 'node:path';

const BASE = process.env.FREEBOX_API_BASE ?? 'http://mafreebox.freebox.fr/api/v15';
const APP_ID = process.env.FREEBOX_APP_ID ?? 'mcp-freebox';
const TOKEN = process.env.FREEBOX_APP_TOKEN;
const OUT = resolve(process.cwd(), 'samples-tier1.json');
if (!TOKEN) { console.error('no token'); process.exit(1); }

async function getJson(url, headers = {}) {
  const r = await fetch(url, { headers });
  return r.json();
}
async function postJson(url, body, headers = {}) {
  const r = await fetch(url, { method: 'POST', headers: { 'content-type': 'application/json', ...headers }, body: JSON.stringify(body) });
  return r.json();
}

const ch = await getJson(`${BASE}/login/`);
const password = createHmac('sha1', TOKEN).update(ch.result.challenge).digest('hex');
const sess = await postJson(`${BASE}/login/session/`, { app_id: APP_ID, password });
const H = { 'X-Fbx-App-Auth': sess.result.session_token };

const PATHS = [
  '/connection/logs/',
  '/connection/ipv6/config/',
  '/lan/browser/interfaces/',
  '/wifi/mac_filter/',
  '/wifi/planning/',
  '/storage/raid/',
  '/fs/info/?path=L0Rpc3F1ZSBkdXIv',  // /Disque dur/ encode base64 — fs/info exige un path
  '/vm/info',
  '/vm/distros/',
];

// /switch/port/{id} — on commence par lister via /switch/status/ pour avoir un id valide
const sw = await getJson(`${BASE}/switch/status/?_dc=${Date.now()}`, H);
const firstPortId = sw?.result?.[0]?.id ?? 1;
PATHS.push(`/switch/port/${firstPortId}/`);
PATHS.push(`/switch/port/${firstPortId}/stats/`);

const out = {};
for (const p of PATHS) {
  const url = `${BASE}${p}${p.includes('?') ? '&' : '?'}_dc=${Date.now()}`;
  const r = await getJson(url, H);
  out[p] = r;
  const ok = r?.success === true ? 'OK' : 'ERR';
  console.error(`[${ok}] ${p}`);
}

writeFileSync(OUT, JSON.stringify(out, null, 2));
console.error(`wrote ${OUT}`);
