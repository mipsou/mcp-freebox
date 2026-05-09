import { chromium } from 'playwright';
import { writeFileSync } from 'node:fs';
import { resolve } from 'node:path';

const EDGE_USER_DATA = `${process.env.LOCALAPPDATA}\\Microsoft\\Edge\\User Data`;
const FREEBOX_URL = 'http://mafreebox.freebox.fr/';
const OUTPUT = resolve(process.cwd(), 'sniff-output.json');
const DEBUG = process.env.SNIFF_DEBUG === '1';
const HEADLESS = process.env.SNIFF_HEADLESS !== '0';
const LOGIN_TIMEOUT_MS = parseInt(process.env.SNIFF_LOGIN_TIMEOUT_MS ?? '300000', 10);

const endpoints = new Map();
const allRequests = [];
const apiRegex = /\/api\/(?:latest|v\d+)\/(.+?)(?:\?|$)/;

function record(req) {
  const url = req.url();
  const method = req.method();
  if (DEBUG) allRequests.push({ method, url });
  const m = url.match(apiRegex);
  if (!m) return;
  const path = '/' + m[1];
  const key = `${method} ${path}`;
  if (!endpoints.has(key)) {
    endpoints.set(key, { method, path, full: url, count: 1 });
  } else {
    endpoints.get(key).count++;
  }
}

const ctx = await chromium.launchPersistentContext(EDGE_USER_DATA, {
  channel: 'msedge',
  headless: HEADLESS,
  args: ['--profile-directory=Default', '--disable-blink-features=AutomationControlled'],
  viewport: { width: 1920, height: 1080 },
  ignoreHTTPSErrors: true,
});

ctx.on('request', record);
ctx.on('response', (resp) => {
  if (!DEBUG) return;
  const url = resp.url();
  if (apiRegex.test(url)) {
    console.error(`[resp] ${resp.status()} ${resp.request().method()} ${url}`);
  }
});

const page = await ctx.newPage();
console.error(`[sniff] navigating to ${FREEBOX_URL} (headless=${HEADLESS})`);
await page.goto(FREEBOX_URL, { waitUntil: 'networkidle', timeout: 30000 }).catch(e => console.error('[goto]', e.message));

// Verifier la session reelle via /api/latest/login/
async function checkLogin() {
  try {
    const resp = await ctx.request.get(`${FREEBOX_URL}api/latest/login/?_dc=${Date.now()}`);
    if (!resp.ok()) return false;
    const body = await resp.json();
    return body?.success === true && body?.result?.logged_in === true;
  } catch (e) {
    if (DEBUG) console.error('[checkLogin] err', e.message);
    return false;
  }
}

let isLoggedIn = await checkLogin();
console.error(`[sniff] initial login state = ${isLoggedIn}`);

if (!isLoggedIn) {
  console.error(`[sniff] please login dans la fenetre Edge ouverte (${FREEBOX_URL}) — j'attends jusqu'a ${LOGIN_TIMEOUT_MS / 1000}s...`);
  const start = Date.now();
  while (Date.now() - start < LOGIN_TIMEOUT_MS) {
    await page.waitForTimeout(3000);
    isLoggedIn = await checkLogin();
    if (isLoggedIn) {
      console.error(`[sniff] login detecte apres ${Math.round((Date.now() - start) / 1000)}s`);
      break;
    }
  }
}

if (!isLoggedIn) {
  console.error(`[sniff] login non detecte dans le delai imparti — abort`);
  await ctx.close();
  writeFileSync(OUTPUT, JSON.stringify({ count: 0, isLoggedIn: false, endpoints: [] }, null, 2));
  process.exit(2);
}

// On garde tout ce qui a ete capture jusqu'ici (rafale post-login = mine d'or)
console.error(`[sniff] post-login: ${endpoints.size} endpoints deja captures, traversee des sections...`);

const sections = [
  '#Fbx.os.app.system',
  '#Fbx.os.app.network',
  '#Fbx.os.app.dhcp',
  '#Fbx.os.app.wifi',
  '#Fbx.os.app.firewall',
  '#Fbx.os.app.portforwarding',
  '#Fbx.os.app.upnp',
  '#Fbx.os.app.dyndns',
  '#Fbx.os.app.vpn',
  '#Fbx.os.app.lan',
  '#Fbx.os.app.parental',
  '#Fbx.os.app.downloader',
  '#Fbx.os.app.airmedia',
  '#Fbx.os.app.fs',
  '#Fbx.os.app.share',
  '#Fbx.os.app.vm',
  '#Fbx.os.app.contacts',
  '#Fbx.os.app.phone',
  '#Fbx.os.app.diskmgr',
  '#Fbx.os.app.netshare',
  '#Fbx.os.app.connection',
  '#Fbx.os.app.update',
];

// page.reload() force le SPA a re-boot et fetcher la section ciblee par le hash
for (const s of sections) {
  const before = endpoints.size;
  await page.evaluate(h => { location.hash = h; }, s).catch(() => {});
  await page.waitForTimeout(100);
  await page.reload({ waitUntil: 'networkidle', timeout: 20000 }).catch(e => console.error(`[reload ${s}]`, e.message));
  await page.waitForTimeout(1500);
  const delta = endpoints.size - before;
  console.error(`[sniff] visit ${s} +${delta} api calls (total=${endpoints.size})`);
}

await ctx.close();

const sorted = [...endpoints.values()].sort((a, b) => a.path.localeCompare(b.path));
const out = { count: sorted.length, isLoggedIn, endpoints: sorted };
if (DEBUG) out.allRequestsSample = allRequests.slice(0, 50);
writeFileSync(OUTPUT, JSON.stringify(out, null, 2));
console.error(`[sniff] wrote ${sorted.length} unique endpoints (totalReq=${allRequests.length}) → ${OUTPUT}`);
