// Runs only on a disposable Windows CI account: upstream registration is intentional.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawn, spawnSync } = require('node:child_process');
const { chromium } = require('../../frontend/node_modules/playwright');

assert.equal(process.env.GITHUB_ACTIONS, 'true', 'Use a disposable GitHub runner, never the user account');
assert.equal(process.env.RUNNER_OS, 'Windows');
const [productInstaller, upstreamInstaller, root] = process.argv.slice(2);
assert.ok(productInstaller && upstreamInstaller && root, 'Expected product installer, upstream installer, new scratch root');
assert.ok(!fs.existsSync(root), 'Scratch root must be new');
fs.mkdirSync(root, { recursive: true });
const upstreamState = path.join(os.homedir(), '.ao');
assert.ok(!fs.existsSync(upstreamState), 'Original application state must be absent on this disposable account');
fs.mkdirSync(upstreamState);
fs.writeFileSync(path.join(upstreamState, 'update-settings.json'), JSON.stringify({ enabled: false, channel: 'latest', nightlyAck: false, feature: null }));
const apps = [
  { id: 'product', name: 'AI Dev Studio Ultimate', exe: 'ai-dev-studio-ultimate.exe', protocol: 'ai-dev-studio-ultimate', installer: productInstaller, port: 31338, debug: 31362 },
  { id: 'upstream', name: 'Orchestrator.inc', exe: 'agent-orchestrator.exe', protocol: 'ao-app', installer: upstreamInstaller, port: 3001, debug: 31364 },
].map(app => ({ ...app, install: path.join(root, app.id, 'installed'), data: path.join(root, app.id, 'data') }));
const report = { startedAt: new Date().toISOString(), tests: [], modelsStarted: 0 };
const pause = ms => new Promise(resolve => setTimeout(resolve, ms));
function run(exe, args) {
  const result = spawnSync(exe, args, { encoding: 'utf8', windowsHide: true, timeout: 120000 });
  assert.equal(result.status, 0, result.error?.message || result.stderr || result.stdout);
  return result.stdout;
}
function protocol(app) {
  return run('reg.exe', ['query', `HKCU\\Software\\Classes\\${app.protocol}\\shell\\open\\command`, '/ve']);
}
function protocolAbsent(app) {
  const result = spawnSync('reg.exe', ['query', `HKCU\\Software\\Classes\\${app.protocol}`], { windowsHide: true, timeout: 10000 });
  assert.ok(!result.error, result.error?.message);
  assert.ok(result.status === 0 || result.status === 1, 'Registry query did not complete');
  return result.status === 1;
}
function shortcut(app) {
  return path.join(process.env.APPDATA, 'Microsoft', 'Windows', 'Start Menu', 'Programs', app.name + '.lnk');
}
async function until(fn, description) {
  for (let i = 0; i < 200; i++) {
    try { const value = await fn(); if (value) return value; } catch {}
    await pause(300);
  }
  throw new Error('Timeout: ' + description);
}
async function start(app) {
  fs.mkdirSync(app.data, { recursive: true });
  const env = { ...process.env, AO_DATA_DIR: app.data, AO_RUN_FILE: path.join(root, app.id, 'running.json'), AO_TELEMETRY_REMOTE: 'off', AO_TELEMETRY_EVENTS: '0', ELECTRON_ENABLE_LOGGING: '1' };
  for (const key of ['AO_PORT', 'AO_DEV_DAEMON_BINARY', 'AO_DAEMON_COMMAND', 'AO_KEEP_DAEMON', 'AO_OWNER', 'AO_APP_RUN_ID', 'NODE_OPTIONS', 'ELECTRON_RUN_AS_NODE']) delete env[key];
  app.log = fs.openSync(path.join(root, app.id, 'desktop.log'), 'a');
  app.child = spawn(path.join(app.install, app.exe), [`--remote-debugging-port=${app.debug}`, '--remote-debugging-address=127.0.0.1'], { env, windowsHide: true, stdio: ['ignore', app.log, app.log] });
  await until(async () => (await fetch(`http://127.0.0.1:${app.debug}/json/version`)).ok, app.id + ' Chromium');
  app.browser = await chromium.connectOverCDP(`http://127.0.0.1:${app.debug}`);
  const page = await until(async () => app.browser.contexts().flatMap(context => context.pages()).find(page => page.url().startsWith('app://renderer/')), app.id + ' renderer');
  app.page = page;
  app.status = await until(async () => { const status = await page.evaluate(() => window.ao.daemon.getStatus()); return status.state === 'ready' ? status : null; }, app.id + ' daemon');
  assert.equal(app.status.port, app.port, 'Default ports must remain distinct');
  for (const route of ['projects', 'sessions']) {
    const response = await fetch(`http://127.0.0.1:${app.port}/api/v1/${route}`);
    assert.equal(response.status, 200);
    assert.equal((await response.json())[route].length, 0, 'No model-backed sessions or real projects');
  }
  await page.screenshot({ path: path.join(root, app.id, 'desktop.png') });
  fs.writeFileSync(path.join(app.data, 'preserve.txt'), app.id + '-uninstall-preservation');
  assert.ok(fs.existsSync(path.join(app.data, 'ao.db')));
  assert.ok(fs.existsSync(shortcut(app)));
  assert.ok(protocol(app).includes(path.join(app.install, app.exe)));
}
async function stop(app) {
  if (!app.browser) return;
  // Closing a window can hide it in the tray; invoke the existing Quit action.
  try { await app.page.evaluate(() => window.ao.menu.action('app.quit')); } catch {}
  await until(() => app.child.exitCode !== null, app.id + ' exit');
  app.browser = null;
  fs.closeSync(app.log);
}
async function uninstall(app) {
  const uninstaller = fs.readdirSync(app.install).find(name => /^Uninstall .*\.exe$/i.test(name));
  assert.ok(uninstaller);
  run(path.join(app.install, uninstaller), ['/S']);
  // NSIS relaunches the uninstaller; its launcher exit does not finish cleanup.
  await until(() => !fs.existsSync(path.join(app.install, app.exe)) &&
    !fs.existsSync(shortcut(app)) && (app.id !== 'product' || protocolAbsent(app)), app.id + ' uninstall files, shortcut and protocol');
  assert.ok(!fs.existsSync(shortcut(app)));
  assert.ok(fs.existsSync(path.join(app.data, 'ao.db')));
  assert.equal(fs.readFileSync(path.join(app.data, 'preserve.txt'), 'utf8'), app.id + '-uninstall-preservation');
}

(async () => {
  for (const app of apps) {
    run(app.installer, ['/S', '/D=' + app.install]);
    assert.ok(fs.existsSync(path.join(app.install, app.exe)));
  }
  for (const app of apps) await start(app);
  assert.notEqual(apps[0].status.pid, apps[1].status.pid);
  assert.ok(fs.existsSync(path.join(apps[0].data, 'electron')));
  assert.ok(fs.existsSync(path.join(os.homedir(), '.ao', 'electron')));
  report.tests.push('both original NSIS installers install and run concurrently', 'distinct default ports, daemon PIDs, protocols, shortcuts and Electron profiles');
  const upstreamProtocol = protocol(apps[1]);
  await stop(apps[0]);
  await uninstall(apps[0]);
  assert.ok(protocolAbsent(apps[0]));
  assert.ok(fs.existsSync(path.join(apps[1].install, apps[1].exe)));
  assert.equal(protocol(apps[1]), upstreamProtocol);
  assert.ok(fs.existsSync(shortcut(apps[1])));
  assert.equal((await fetch(`http://127.0.0.1:${apps[1].port}/healthz`)).status, 200);
  report.tests.push('product uninstall preserves upstream installation, protocol, shortcut and running daemon');
  run(apps[0].installer, ['/S', '/D=' + apps[0].install]);
  await start(apps[0]);
  const productProtocol = protocol(apps[0]);
  await stop(apps[1]);
  await uninstall(apps[1]);
  report.upstreamProtocolRemoved = protocolAbsent(apps[1]);
  assert.equal(protocol(apps[0]), productProtocol);
  assert.ok(fs.existsSync(shortcut(apps[0])));
  assert.equal((await fetch(`http://127.0.0.1:${apps[0].port}/healthz`)).status, 200);
  report.tests.push('upstream uninstall preserves product installation, protocol, shortcut and running daemon');
  await stop(apps[0]);
  await uninstall(apps[0]);
  report.tests.push('both final uninstalls retain synthetic databases and preservation files');
  report.passed = true;
})().catch(error => { report.passed = false; report.error = error.stack; process.exitCode = 1; }).finally(async () => {
  for (const app of apps) { try { await stop(app); } catch {} }
  report.completedAt = new Date().toISOString();
  fs.writeFileSync(path.join(root, 'coexistence-result.json'), JSON.stringify(report, null, 2));
  console.log(JSON.stringify(report, null, 2));
});
