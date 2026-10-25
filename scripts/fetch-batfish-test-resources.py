#!/usr/bin/env python3
"""Fetch pinned public parser references, never used by ordinary tests."""
import concurrent.futures
import hashlib
import json
from pathlib import Path
import time
import urllib.request

REVISION = '2f6128b8f113c6406254c1557daad0c2b05fb739'
PREFIX = 'projects/batfish/src/test/resources/org/batfish/grammar/'
VENDORS = {'cisco', 'cisco_asa', 'cisco_xr', 'fortios', 'juniper', 'palo_alto', 'frr'}
ROOT = Path(__file__).resolve().parents[1] / '.test' / 'batfish'


def fetch(url):
    for attempt in range(4):
        try:
            with urllib.request.urlopen(url, timeout=60) as response:
                return response.read()
        except Exception:
            if attempt == 3:
                raise
            time.sleep(attempt + 1)


def store(path, data):
    if path.exists() and path.read_bytes() != data:
        raise RuntimeError(f'Existing content differs: {path.relative_to(ROOT)}')
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(data)


def main():
    tree = json.loads(fetch(f'https://api.github.com/repos/batfish/batfish/git/trees/{REVISION}?recursive=1'))
    if tree.get('truncated'):
        raise RuntimeError('Incomplete tree')
    entries = [e for e in tree['tree'] if e['type'] == 'blob' and
               e['path'].startswith(PREFIX) and e['path'][len(PREFIX):].split('/')[0] in VENDORS]
    if not entries:
        raise RuntimeError('Empty corpus')
    # Also preserve upstream licensing files without rewriting attribution.
    entries += [e for e in tree['tree'] if e['type'] == 'blob' and '/' not in e['path'] and
                e['path'].upper().startswith(('LICENSE', 'NOTICE', 'COPYING'))]

    def download(entry):
        source = entry['path']
        relative = 'grammar/' + source[len(PREFIX):] if source.startswith(PREFIX) else source
        data = fetch(f'https://raw.githubusercontent.com/batfish/batfish/{REVISION}/{source}')
        blob = hashlib.sha1(b'blob ' + str(len(data)).encode() + b'\0' + data).hexdigest()
        if blob != entry['sha']:
            raise RuntimeError(f'Git blob hash mismatch: {source}')
        store(ROOT / relative, data)
        return {'path': relative, 'source': source, 'bytes': len(data),
                'sha256': hashlib.sha256(data).hexdigest(), 'gitBlob': blob}

    with concurrent.futures.ThreadPoolExecutor(max_workers=12) as pool:
        records = list(pool.map(download, entries))
    manifest = {'repository': 'https://github.com/batfish/batfish', 'revision': REVISION,
                'vendors': sorted(VENDORS), 'files': sorted(records, key=lambda r: r['path'])}
    store(ROOT / 'manifest.json', (json.dumps(manifest, indent=2) + '\n').encode())
    print(f'Verified {len(records)} files, {sum(r["bytes"] for r in records)} bytes at {REVISION}')


if __name__ == '__main__':
    main()
