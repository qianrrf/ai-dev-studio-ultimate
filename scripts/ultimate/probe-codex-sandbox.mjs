// Native command/exec only: no thread, model turn, account refresh or login.
import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { createInterface } from 'node:readline';
import { mkdir, mkdtemp, writeFile, access } from 'node:fs/promises';
import { createServer } from 'node:http';
import os from 'node:os';
import path from 'node:path';

assert.equal(process.platform, 'win32', 'This probe tests the native Windows sandbox');
const root = path.resolve(process.env.AIDSU_VALIDATION_DIR || path.join(os.tmpdir(), 'aidsu-sandbox-probe'));
await mkdir(root, { recursive: true });
const workspace = await mkdtemp(path.join(root, 'codex-sandbox-'));
await writeFile(path.join(workspace, 'input.txt'), 'READ-ONLY-PROBE', 'utf8');
const binary = process.env.CODEX_BINARY || 'codex.cmd';
assert.ok(!/["%\r\n]/.test(binary), 'CODEX_BINARY must be a literal executable path');
const sandbox = process.env.AIDSU_CODEX_WINDOWS_SANDBOX || '';
assert.ok(['', 'elevated', 'unelevated', 'mxc'].includes(sandbox), 'Unsupported sandbox');
const override = sandbox ? ` -c windows.sandbox=${sandbox}` : '';
const child = spawn(process.env.ComSpec || 'cmd.exe', ['/d', '/v:off', '/s', '/c', `""${binary}"${override} app-server"`], { windowsVerbatimArguments: true, windowsHide: true, stdio: ['pipe', 'pipe', 'pipe'] });
const report = { startedAt: new Date().toISOString(), modelPromptsSent: 0, sandbox: sandbox || 'user configuration', workspace, steps: [] };
const pending = new Map();
let nextId = 0, stderrBytes = 0, loopbackRequests = 0;
child.stderr.on('data', chunk => { stderrBytes += chunk.length; });
function failPending(error) {
  for (const entry of pending.values()) { clearTimeout(entry.timer); entry.reject(error); }
  pending.clear();
}
child.on('error', () => failPending(new Error('Could not launch native Codex')));
child.stdin.on('error', () => failPending(new Error('Native Codex input closed')));
child.on('exit', code => failPending(new Error(`Native Codex exited (${code})`)));
createInterface({ input: child.stdout }).on('line', line => {
  let message;
  try { message = JSON.parse(line); } catch { return; }
  if (message.method && message.id !== undefined) {
    child.stdin.write(JSON.stringify({ id: message.id, error: { code: -32601, message: 'No client tools in sandbox probe' } }) + '\n');
    return;
  }
  const entry = pending.get(message.id);
  if (!entry) return;
  clearTimeout(entry.timer); pending.delete(message.id);
  if (message.error) entry.reject(new Error(`${entry.method}: RPC ${message.error.code}`));
  else entry.resolve(message.result);
});
function request(method, params) {
  return new Promise((resolve, reject) => {
    const id = ++nextId;
    const timer = setTimeout(() => { pending.delete(id); reject(new Error(`${method}: timed out`)); }, 20000);
    pending.set(id, { resolve, reject, timer, method });
    child.stdin.write(JSON.stringify({ id, method, params }) + '\n');
  });
}
const server = createServer((_req, res) => { loopbackRequests++; res.end('PROBE'); });
await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
try {
  await request('initialize', { clientInfo: { name: 'ultimate-sandbox-probe', version: '0.1.0' }, capabilities: { experimentalApi: true } });
  child.stdin.write(JSON.stringify({ method: 'initialized' }) + '\n');
  const commands = [
    ['read', "Get-Content -LiteralPath input.txt -ErrorAction Stop"],
    ['write', "Set-Content -LiteralPath forbidden.txt -Value probe -ErrorAction Stop"],
    ['loopback', `$ErrorActionPreference='Stop'; Invoke-WebRequest -UseBasicParsing -TimeoutSec 3 http://127.0.0.1:${server.address().port}/`],
  ];
  for (const [name, command] of commands) {
    try {
      const result = await request('command/exec', { command: [path.join(process.env.SystemRoot || 'C:\\Windows', 'System32', 'WindowsPowerShell', 'v1.0', 'powershell.exe'), '-NoProfile', '-NonInteractive', '-Command', command], cwd: workspace, sandboxPolicy: { type: 'readOnly', networkAccess: false }, timeoutMs: 10000 });
      report.steps.push({ name, exitCode: result.exitCode, expectedMarkerRead: name === 'read' && result.stdout.includes('READ-ONLY-PROBE') });
    } catch (error) { report.steps.push({ name, error: error.message }); }
  }
  let fileExists = false;
  try { await access(path.join(workspace, 'forbidden.txt')); fileExists = true; } catch {}
  report.forbiddenFileExists = fileExists;
  report.loopbackRequests = loopbackRequests;
  const [read, write, network] = report.steps;
  report.ok = read.exitCode === 0 && read.expectedMarkerRead && Number.isInteger(write.exitCode) && write.exitCode !== 0 && Number.isInteger(network.exitCode) && network.exitCode !== 0 && !fileExists && loopbackRequests === 0;
} catch (error) { report.ok = false; report.error = error.message; }
finally {
  server.closeAllConnections();
  await new Promise(resolve => server.close(resolve));
  child.stdin.end();
  if (child.exitCode === null) await new Promise(resolve => {
    const timer = setTimeout(() => { const stop = spawn('taskkill.exe', ['/PID', String(child.pid), '/T', '/F'], { windowsHide: true, stdio: 'ignore' }); stop.once('exit', resolve); stop.once('error', resolve); }, 2000);
    child.once('exit', () => { clearTimeout(timer); resolve(); });
  });
  failPending(new Error('Probe ended'));
  report.finishedAt = new Date().toISOString(); report.stderrBytes = stderrBytes;
  await writeFile(path.join(workspace, 'result.json'), JSON.stringify(report, null, 2) + '\n');
  console.log(JSON.stringify(report, null, 2));
  if (!report.ok) process.exitCode = 1;
}
