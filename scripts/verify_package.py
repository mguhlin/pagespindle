#!/usr/bin/env python3
"""Verify complete archive inventory and byte identity against the build stage."""
import hashlib,json,sys,zipfile
from pathlib import Path
root=Path(__file__).resolve().parents[1];target=sys.argv[1];stage=root/'dist'/target/'PageSpindle';manifest=json.loads((stage/'INVENTORY.json').read_text());expected={'PageSpindle/INVENTORY.json':(stage/'INVENTORY.json').read_bytes()}
for x in manifest['files']:
 p=stage/x['path'];data=x['symlink'].encode() if 'symlink' in x else p.read_bytes()
 if 'sha256'in x and hashlib.sha256(data).hexdigest()!=x['sha256']:raise SystemExit('Stale component '+x['path'])
 expected['PageSpindle/'+x['path']]=data
with zipfile.ZipFile(root/'dist'/('pagespindle-'+target+'.zip')) as z:
 if len(z.namelist())!=len(expected) or set(z.namelist())!=set(expected):raise SystemExit('Package inventory mismatch')
 for name,data in expected.items():
  if z.read(name)!=data:raise SystemExit('Stale packaged file '+name)
print('Verified',target,len(expected),'components and hashes')
