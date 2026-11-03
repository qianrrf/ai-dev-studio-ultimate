import assert from 'node:assert/strict';
import { spawn } from 'node:child_process';
import { createInterface } from 'node:readline';
import { mkdir, writeFile } from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';

const root = path.resolve(process.env.AIDSU_VALIDATION_DIR || path.join(os.tmpdir(), 'aidsu-acp-probe'));
const workspace = path.join(root, 'acp-workspace');
await mkdir(workspace, { recursive: true });
const binary = process.env.DSH_BINARY || 'dsh.cmd';
assert.ok(!/["%\r\n]/.test(binary), 'DSH_BINARY must be a literal executable path');
const child = spawn(process.env.ComSpec || 'cmd.exe', ['/d', '/s', '/c', `""${binary}" --profile acp"`], { windowsVerbatimArguments: true, windowsHide: true, stdio: ['pipe', 'pipe', 'pipe'] });
const report = { startedAt: new Date().toISOString(), modelPromptsSent: 0, workspace, steps: [], nonProtocolLines: 0 };
let nextId = 1;
const pending = new Map();
let stderrBytes = 0;
child.stderr.on('data', chunk => { stderrBytes += chunk.length; });
const lines = createInterface({ input: child.stdout });
lines.on('line', line => {
  let message;
  try { message = JSON.parse(line); } catch { report.nonProtocolLines++; return; }
  if (message.method && message.id !== undefined) {
    child.stdin.write(JSON.stringify({ jsonrpc: '2.0', id: message.id, error: { code: -32601, message: 'No client tools enabled in protocol-only probe' } }) + '\n');
    return;
  }
  const entry = pending.get(message.id);
  if (!entry) return;
  clearTimeout(entry.timer);
  pending.delete(message.id);
  if (message.error) entry.reject(new Error(`${entry.method}: RPC ${message.error.code}: ${String(message.error.message).replace(/sk-[\w-]+/g, '[redacted]')}`));
  else entry.resolve(message.result);
});
child.on('error', error => { for (const entry of pending.values()) entry.reject(error); });
function request(method, params) {
  return new Promise((resolve, reject) => {
    const id = nextId++;
    const timer = setTimeout(() => { pending.delete(id); reject(new Error(`${method}: timed out`)); }, 30000);
    pending.set(id, { resolve, reject, timer, method });
    child.stdin.write(JSON.stringify({ jsonrpc: '2.0', id, method, params }) + '\n');
  });
}
try {
  const initialized = await request('initialize', { protocolVersion: 1, clientCapabilities: { fs: { readTextFile: false, writeTextFile: false }, terminal: false }, clientInfo: { name: 'ai-dev-studio-ultimate-protocol-probe', version: '0.1.0' } });
  assert.equal(initialized.protocolVersion, 1);
  report.steps.push({ method: 'initialize', ok: true, protocolVersion: initialized.protocolVersion, agentInfo: initialized.agentInfo, capabilities: initialized.agentCapabilities, authMethodCount: initialized.authMethods?.length || 0 });
  const session = await request('session/new', { cwd: workspace, mcpServers: [] });
  assert.equal(typeof session.sessionId, 'string');
  assert.ok(session.sessionId);
  report.steps.push({ method: 'session/new', ok: true, sessionId: session.sessionId, resultKeys: Object.keys(session), configOptionCount: session.configOptions?.length || 0 });
  try { await request('session/close', { sessionId: session.sessionId }); report.steps.push({ method: 'session/close', ok: true }); }
  catch (error) { report.steps.push({ method: 'session/close', ok: false, error: error.message }); }
  report.ok = true;
} catch (error) {
  report.ok = false;
  report.error = error.message;
  process.exitCode = 1;
} finally {
  report.finishedAt = new Date().toISOString();
  report.stderrBytes = stderrBytes;
  for (const entry of pending.values()) clearTimeout(entry.timer);
  child.stdin.end();
  if (child.exitCode === null) await new Promise(resolve => {
    const timer = setTimeout(() => { const kill = spawn('taskkill.exe', ['/PID', String(child.pid), '/T', '/F'], { windowsHide: true, stdio: 'ignore' }); kill.on('close', resolve); }, 2000);
    child.once('exit', () => { clearTimeout(timer); resolve(); });
  });
  await writeFile(path.join(root, 'dsh-acp-protocol.json'), JSON.stringify(report, null, 2) + '\n');
  console.log(JSON.stringify(report, null, 2));
}
