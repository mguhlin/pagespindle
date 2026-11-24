#!/usr/bin/env python3
"""Fetch pinned, hashed runtime/data assets; never execute downloaded scripts."""
import argparse, concurrent.futures, hashlib, json, os, shutil, tarfile, urllib.request, zipfile
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]
LOCK=json.loads((ROOT/'dependencies.json').read_text())

def download(item):
    cache=ROOT/'.cache'/item['name'];cache.parent.mkdir(exist_ok=True)
    def digest(p):
        h=hashlib.sha256()
        with p.open('rb') as f:
            for b in iter(lambda:f.read(4*1024*1024),b''):h.update(b)
        return h.hexdigest()
    if cache.exists() and digest(cache)==item['sha256']:return cache
    part=cache.with_suffix(cache.suffix+'.part')
    with urllib.request.urlopen(item['url'],timeout=60) as r, part.open('wb') as f:shutil.copyfileobj(r,f)
    if part.stat().st_size!=item['size'] or digest(part)!=item['sha256']:
        part.unlink(missing_ok=True);raise ValueError('Integrity check failed: '+item['name'])
    part.replace(cache);print('Verified',item['name'],flush=True);return cache

def extract(archive,dest):
    dest.mkdir(parents=True,exist_ok=True)
    if archive.suffix=='.zip':
        with zipfile.ZipFile(archive) as z:
            for i in z.infolist():
                if i.filename.startswith('/') or '..' in Path(i.filename).parts:raise ValueError('Unsafe archive path')
            z.extractall(dest)
    else:
        with tarfile.open(archive) as t:t.extractall(dest,filter='data')
    if os.name!='nt':
        for p in dest.rglob('*'):
            if p.is_file() and (p.name.startswith(('llama-','sherpa-onnx-')) or '.so' in p.name or p.suffix=='.dylib'):p.chmod(0o755)

def main():
    ap=argparse.ArgumentParser();ap.add_argument('--target',choices=list(LOCK['targets']),required=True);ap.add_argument('--common',action='store_true');args=ap.parse_args()
    items=list(LOCK['targets'][args.target].values())
    if args.common:items+=list(LOCK['common'].values())
    with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:files=list(pool.map(download,items))
    runtime=ROOT/'runtime'/args.target
    for role,item in LOCK['targets'][args.target].items():
        if (runtime/role).exists():shutil.rmtree(runtime/role)
        extract(ROOT/'.cache'/item['name'],runtime/role)
    if args.common:
        model=LOCK['common']['model'];voice=LOCK['common']['voice'];common=ROOT/'models';common.mkdir(exist_ok=True)
        shutil.copy2(ROOT/'.cache'/model['name'],common/model['name']);extract(ROOT/'.cache'/voice['name'],common)
    print('Prepared',args.target,flush=True)

if __name__=='__main__':main()
