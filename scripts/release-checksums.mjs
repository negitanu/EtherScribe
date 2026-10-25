#!/usr/bin/env node
import fs from 'node:fs';
import path from 'node:path';
import { createHash } from 'node:crypto';
const directory = process.argv[2];
if (!directory) throw new Error('Release asset directory required');
const files = fs.readdirSync(directory).filter(file => /\.(?:zip|dmg)$/.test(file)).sort();
if (files.length !== 3) throw new Error('Expected Windows ZIP and macOS ZIP/DMG');
const version = JSON.parse(fs.readFileSync('desktop/package.json')).version;
const expected = [`EtherScribe-${version}-win-x64.zip`, `EtherScribe-${version}-mac-arm64.zip`, `EtherScribe-${version}-mac-arm64.dmg`].sort();
if (JSON.stringify(files) !== JSON.stringify(expected)) throw new Error('Unexpected release asset names');
const sums = files.map(file => `${createHash('sha256').update(fs.readFileSync(path.join(directory, file))).digest('hex')}  ${file}`).join('\n') + '\n';
fs.writeFileSync(path.join(directory, 'SHA256SUMS.txt'), sums);
console.log('Verified three expected assets and generated SHA256SUMS.txt');
