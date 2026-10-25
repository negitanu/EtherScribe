#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const require = createRequire(path.join(root, 'desktop/package.json'));
const asar = require('@electron/asar');
const [directory, platform, arch] = process.argv.slice(2);
if (!directory || !['darwin', 'windows'].includes(platform) || !['arm64', 'amd64'].includes(arch)) throw new Error('Usage: verify-release.mjs directory darwin|windows arm64|amd64');
const app = path.resolve(directory);
const resources = platform === 'darwin' ? path.join(app, 'Contents/Resources') : path.join(app, 'resources');
if (!((platform === 'darwin' && arch === 'arm64') || (platform === 'windows' && arch === 'amd64'))) throw new Error('Unsupported release architecture');
const version = JSON.parse(fs.readFileSync(path.join(root, 'desktop/package.json'))).version;
for (const name of fs.readdirSync(resources)) {
  if (!['app.asar', 'engine', 'licenses', 'electron.icns'].includes(name) && !(platform === 'darwin' && /^[A-Za-z0-9_]+\.lproj$/.test(name))) throw new Error('Unexpected resources entry');
}
const licenseSources = {
  'Go-LICENSE.txt': 'core/third_party/go/LICENSE',
  'golang-x-text-LICENSE.txt': 'core/third_party/text/LICENSE',
  'golang-x-text-PATENTS.txt': 'core/third_party/text/PATENTS',
};
if (JSON.stringify(fs.readdirSync(path.join(resources, 'licenses')).sort()) !== JSON.stringify(Object.keys(licenseSources).sort())) throw new Error('Unexpected license payload');
for (const [name, source] of Object.entries(licenseSources)) {
  if (!fs.readFileSync(path.join(resources, 'licenses', name)).equals(fs.readFileSync(path.join(root, source)))) throw new Error('Missing/modified dependency license');
}
const archive = path.join(resources, 'app.asar');
const files = asar.listPackage(archive).map(p => p.replaceAll('\\', '/'));
for (const name of files) {
  if (!/^\/(?:package\.json|dist(?:\/|$)|dist-electron(?:\/|$))/.test(name)) throw new Error('Unapproved app payload entry');
  if (/\.(?:map|conf|xml|set|pem|key|env)$/.test(name)) throw new Error('Configuration/source map in app payload');
  if (/\.(?:js|html|css|json)$/.test(name)) {
    const content = asar.extractFile(archive, name.slice(1)).toString();
    if (/(?:\/Users\/[^/\s]+\/|[A-Z]:\\Users\\[^\\\s]+\\|\/home\/runner\/work\/)/.test(content)) throw new Error('Local build path in app payload');
  }
}
const metadata = JSON.parse(asar.extractFile(archive, 'package.json').toString());
if (metadata.name !== 'etherscribe' || metadata.version !== version) throw new Error('App version mismatch');
for (const entry of ['dist/index.html', 'dist-electron/main/main.js', 'dist-electron/preload/preload.js']) {
  if (!files.includes('/' + entry)) throw new Error('Missing application entry');
}
const forbidden = /^(?:\.data|data|\.aws|\.ssh|\.env(?:\..*)?|\.test|testdata|node_modules|\.git)$/;
function walk(directory) {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    if (forbidden.test(entry.name)) throw new Error('Private/development directory in distribution');
    const filename = path.join(directory, entry.name);
    if (entry.isSymbolicLink()) {
      const resolved = fs.realpathSync(filename);
      if (!resolved.startsWith(app + path.sep)) throw new Error('External symlink in distribution');
    } else if (entry.isDirectory()) walk(filename);
    else if (/\.(?:conf|set|pem|key|map)$/.test(entry.name)) throw new Error('Unexpected input/source map in distribution');
  }
}
walk(app);
function verifyArchitecture(filename) {
  const data = fs.readFileSync(filename);
  if (platform === 'windows') {
    if (data.toString('ascii', 0, 2) !== 'MZ' || data.readUInt16LE(data.readUInt32LE(0x3c) + 4) !== 0x8664) throw new Error('Expected Windows x64 executable');
  } else if (data.readUInt32LE(0) !== 0xfeedfacf || data.readUInt32LE(4) !== 0x100000c) throw new Error('Expected macOS arm64 executable');
}
const engineDirectory = path.join(resources, 'engine');
const engineName = platform === 'windows' ? 'etherscribe.exe' : 'etherscribe';
if (JSON.stringify(fs.readdirSync(engineDirectory)) !== JSON.stringify([engineName])) throw new Error('Unexpected engine files');
const engine = path.join(engineDirectory, engineName);
verifyArchitecture(engine);
verifyArchitecture(platform === 'windows' ? path.join(app, 'EtherScribe.exe') : path.join(app, 'Contents/MacOS/EtherScribe'));
const run = (args, input) => {
  const result = spawnSync(engine, args, { input, encoding: 'utf8', timeout: 20000 });
  if (result.error || result.status !== 0) throw new Error('Packaged engine failed to execute');
  return result.stdout;
};
if (run(['version']).trim() !== `EtherScribe ${version}`) throw new Error('Engine version mismatch');
const output = run(['rpc'], JSON.stringify({ jsonrpc: '2.0', id: 1, method: 'engine.version', params: {} }) + '\n');
const response = output.split('\n').filter(Boolean).map(line => JSON.parse(line)).find(message => message.id === 1);
if (response?.result?.version !== version) throw new Error('Packaged RPC handshake failed');


if (platform === 'darwin') {
 const check = spawnSync('/usr/bin/codesign', ['--verify', '--deep', '--strict', app], { encoding: 'utf8' });
 if (check.status !== 0) throw new Error('macOS ad-hoc signature verification failed');
 console.log('macOS ad-hoc signature verified');
}

console.log(`Distribution audit passed: ${platform}/${arch}, app and engine ${version}, allowlisted files, preserved licenses, correct architectures and RPC handshake`);
