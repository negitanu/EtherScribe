const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { createRequire } = require('node:module');
const { spawnSync } = require('node:child_process');
const root = path.resolve(__dirname, '..');
const asar = createRequire(path.join(root, 'desktop/package.json'))('@electron/asar');
const version = require('../desktop/package.json').version;
async function fixture(change) {
  const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'etherscribe-payload-test-'));
  const input = path.join(temporary, 'input');
  const app = path.join(temporary, 'Test.app');
  const resources = path.join(app, 'Contents/Resources');
  const write = (base, name, content) => { const file = path.join(base, name); fs.mkdirSync(path.dirname(file), { recursive: true }); fs.writeFileSync(file, content); };
  for (const name of ['dist/index.html', 'dist-electron/main/main.js', 'dist-electron/preload/preload.js']) write(input, name, 'synthetic');
  write(input, 'package.json', JSON.stringify({ name: 'etherscribe', version }));
  write(resources, 'engine/etherscribe', Buffer.alloc(64));
  for (const [name, source] of Object.entries({ 'Go-LICENSE.txt': 'core/third_party/go/LICENSE', 'golang-x-text-LICENSE.txt': 'core/third_party/text/LICENSE', 'golang-x-text-PATENTS.txt': 'core/third_party/text/PATENTS' })) write(resources, 'licenses/' + name, fs.readFileSync(path.join(root, source)));
  try {
    change({ input, resources, write });
    await asar.createPackage(input, path.join(resources, 'app.asar'));
    return spawnSync(process.execPath, [path.join(__dirname, 'verify-release.mjs'), app, 'darwin', 'arm64'], { encoding: 'utf8' });
  } finally { fs.rmSync(temporary, { recursive: true, force: true }); }
}
for (const [name, change, expected] of [
  ['reject raw configuration', ({ input, write }) => write(input, 'dist/sample.conf', 'DO_NOT_SHIP'), /Configuration\/source map in app payload/],
  ['reject source maps', ({ input, write }) => write(input, 'dist/main.js.map', '{}'), /Configuration\/source map in app payload/],
  ['reject environment file', ({ resources, write }) => write(resources, '.env', 'DO_NOT_SHIP'), /Unexpected resources entry/],
  ['reject mismatched version', ({ input, write }) => write(input, 'package.json', JSON.stringify({ name: 'etherscribe', version: '0.0.0' })), /App version mismatch/],
  ['reject wrong architecture', () => {}, /Expected macOS arm64 executable/],
]) test(name, async () => { const result = await fixture(change); assert.equal(result.status, 1); assert.match(result.stderr, expected); });
