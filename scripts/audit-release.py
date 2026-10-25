#!/usr/bin/env python3
"""Audit public source/history without printing credential values or private input."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess as sp

ROOT = Path(__file__).resolve().parents[1]
PATTERNS = [
    rb'-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----',
    rb'AKIA[0-9A-Z]{16}', rb'gh[pousr]_[A-Za-z0-9]{30,}',
    rb'github_pat_[A-Za-z0-9_]{40,}', rb'sk-(?:proj-|svcacct-)[A-Za-z0-9_-]{40,}',
    rb'https?://[^\s/:]+:[^\s/@]+@',
]

def git(*args):
    return sp.check_output(['git', *args], cwd=ROOT)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--history', action='store_true')
    parser.add_argument('--tag')
    args = parser.parse_args()
    pkg = json.loads((ROOT / 'desktop/package.json').read_text())
    lock = json.loads((ROOT / 'desktop/package-lock.json').read_text())
    version = pkg['version']
    go = re.search(r'Version\s*=\s*"([^"]+)"', (ROOT / 'core/internal/version/version.go').read_text())[1]
    assert version == go == lock['version'] == lock['packages']['']['version'], 'Version mismatch'
    if args.tag:
        assert args.tag == 'v' + version, 'Tag/version mismatch'
    manifest = json.loads((ROOT / '.test/batfish/manifest.json').read_text())
    public_blobs = set()
    reference_paths = {".test/batfish/" + entry["path"] for entry in manifest["files"]}
    for entry in manifest['files']:
        data = (ROOT / '.test/batfish' / entry['path']).read_bytes()
        assert hashlib.sha256(data).hexdigest() == entry['sha256'], 'Public reference hash mismatch'
        public_blobs.add(entry['gitBlob'])
    paths = git('ls-files', '-z').decode().split('\0')
    for name in filter(None, paths):
        if name.startswith('.test/batfish/'):
            assert name in reference_paths or name == '.test/batfish/manifest.json', 'Unregistered reference payload tracked'
            continue
        parts = Path(name).parts
        assert not any(p in {'.data', 'data', '.aws', '.ssh', 'node_modules', 'exports', 'backups'} for p in parts), 'Private/generated path tracked'
        assert not any(p == '.env' or p.startswith('.env.') and p != '.env.example' for p in parts), 'Environment file tracked'
        assert not name.startswith(('desktop/dist/', 'desktop/release/', 'desktop/resources/engine/')), 'Build output tracked'
        assert Path(name).suffix.lower() not in {'.pem', '.key', '.p12', '.pfx', '.db', '.sqlite', '.xlsx'}, 'Sensitive/output file tracked'
        if Path(name).suffix in {'.conf', '.xml', '.set'}:
            assert name.startswith(('core/testdata/', 'docs/reference/')), 'Unexpected configuration file tracked'
    objects = git('rev-list', '--objects', '--all').decode().splitlines() if args.history else git('ls-files', '-s').decode().splitlines()
    ids = sorted({line.split()[0] if args.history else line.split()[1] for line in objects})
    batch = sp.run(['git', 'cat-file', '--batch'], cwd=ROOT, input=('\n'.join(ids) + '\n').encode(), stdout=sp.PIPE, check=True).stdout
    offset = 0
    checked = 0
    for _ in ids:
        end = batch.index(b'\n', offset)
        sha, kind, size = batch[offset:end].split()
        start = end + 1
        data = batch[start:start + int(size)]
        offset = start + int(size) + 1
        if kind != b'blob' or sha.decode() in public_blobs:
            continue
        assert not any(re.search(p, data) for p in PATTERNS), 'Credential pattern detected in tracked content/history (values suppressed)'
        checked += 1
    print(f'Public-source audit passed: {len(list(filter(None, paths)))} paths, {checked} scanned blobs, {len(public_blobs)} verified upstream blobs; version {version}')
    print('Owner Git identity and third-party license attribution are intentionally retained; test fixtures are synthetic/public references.')


if __name__ == '__main__':
    main()
