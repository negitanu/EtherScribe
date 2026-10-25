#!/usr/bin/env node
import { mkdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const os = process.argv[2] || ({ win32: 'windows' }[process.platform] || process.platform);
const arch = process.argv[3] || ({ x64: 'amd64' }[process.arch] || process.arch);
if (!['windows', 'darwin', 'linux'].includes(os) || !['amd64', 'arm64'].includes(arch)) throw new Error('Unsupported engine target');
const output = path.resolve(root, process.argv[4] || `desktop/resources/engine/etherscribe${os === 'windows' ? '.exe' : ''}`);
mkdirSync(path.dirname(output), { recursive: true });
const result = spawnSync('go', ['build', '-trimpath', '-buildvcs=false', '-ldflags=-s -w', '-o', output, './cmd/etherscribe'], {
  cwd: path.join(root, 'core'), stdio: 'inherit',
  env: { ...process.env, CGO_ENABLED: '0', GOOS: os, GOARCH: arch, GOPROXY: 'off' },
});
if (result.error) throw result.error;
if (result.status !== 0) process.exit(result.status ?? 1);
console.log(`Built engine ${os}/${arch}`);
