#!/usr/bin/env python3
"""Bundle common model data and exact-source archives, with stable ZIP metadata."""
import argparse,hashlib,json,shutil,urllib.request,zipfile
from pathlib import Path
parser=argparse.ArgumentParser();parser.add_argument('--source-only',action='store_true');args=parser.parse_args()
ROOT=Path(__file__).resolve().parents[1];dist=ROOT/'dist';dist.mkdir(exist_ok=True);source_dir=ROOT/'.cache/corresponding-source';source_dir.mkdir(exist_ok=True)
sources=[('offline-speech-9fdb771e-source.tar.gz','https://codeload.github.com/dimacurrentai/offline-speech/tar.gz/9fdb771e5bfeec518ed7e9f7afc0f27985623a97'),('espeak-ng-browser-890071ff.tar.gz','https://codeload.github.com/espeak-ng/espeak-ng/tar.gz/890071ff30231883d761e4d59edabd6d69faf159'),('sherpa-onnx-v1.12.35.tar.gz','https://codeload.github.com/k2-fsa/sherpa-onnx/tar.gz/refs/tags/v1.12.35'),('piper-phonemize-78a788e0.zip','https://codeload.github.com/csukuangfj/piper-phonemize/zip/78a788e0b719013401572d70fef372e77bff8e43'),('espeak-ng-7251208e.zip','https://codeload.github.com/csukuangfj/espeak-ng/zip/7251208ee427ef87e94f39d0e99b15d287224ed6'),('llama.cpp-b12681.tar.gz','https://codeload.github.com/ggml-org/llama.cpp/tar.gz/refs/tags/b12681')]
def add(z,p,name):
 info=zipfile.ZipInfo(name,date_time=(1980,1,1,0,0,0));info.compress_type=zipfile.ZIP_STORED;info.external_attr=0o100644<<16
 with p.open('rb')as src,z.open(info,'w',force_zip64=True)as dst:shutil.copyfileobj(src,dst,4*1024*1024)
media=json.loads((ROOT/'native/media-lock.json').read_text())
for row in media.values():sources.append((row['name'],row['url']))
inventory=[]
locked={row['file']:row for row in json.loads((ROOT/'SOURCE_ARCHIVES.json').read_text())} if (ROOT/'SOURCE_ARCHIVES.json').exists() else {}
for name,url in sources:
 p=source_dir/name
 if not p.exists()and(ROOT/'.cache'/name).exists():shutil.copy2(ROOT/'.cache'/name,p)
 if not p.exists():
  with urllib.request.urlopen(url,timeout=90)as r,p.open('wb')as f:shutil.copyfileobj(r,f)
 row={'file':name,'source':url,'sha256':hashlib.file_digest(p.open('rb'),'sha256').hexdigest(),'bytes':p.stat().st_size}
 if name in locked and row!=locked[name]:raise ValueError('Corresponding-source identity changed: '+name)
 inventory.append(row);print('Source retained:',name,flush=True)
(ROOT/'SOURCE_ARCHIVES.json').write_text(json.dumps(inventory,indent=2)+'\n')
with zipfile.ZipFile(dist/'pagespindle-speech-and-engine-source.zip','w')as z:
 for row in inventory:add(z,source_dir/row['file'],'PageSpindle-source/'+row['file'])
 for name in ['SOURCE_ARCHIVES.json','THIRD_PARTY.md','LICENSE','native/speech-worker.c','scripts/build.py','dependencies.json','native/media-lock.json','scripts/build_media.py']:add(z,ROOT/name,'PageSpindle-source/'+name)
if args.source_only:
 print('Packaged corresponding source only',flush=True);raise SystemExit(0)
files=sorted(p for p in (ROOT/'models').rglob('*') if p.is_file())+sorted(p for p in (ROOT/'licenses').rglob('*') if p.is_file())+[ROOT/'THIRD_PARTY.md']
rows=[]
with zipfile.ZipFile(dist/'pagespindle-common-models.zip','w',allowZip64=True)as z:
 for p in files:
  name=p.relative_to(ROOT).as_posix()
  with p.open('rb')as f:digest=hashlib.file_digest(f,'sha256').hexdigest()
  rows.append({'path':name,'bytes':p.stat().st_size,'sha256':digest})
  add(z,p,'PageSpindle/'+name)
 info=zipfile.ZipInfo('PageSpindle/MODELS_INVENTORY.json',date_time=(1980,1,1,0,0,0));info.external_attr=0o100644<<16;z.writestr(info,json.dumps({'schema':1,'files':rows},indent=2)+'\n')
with zipfile.ZipFile(dist/'pagespindle-common-models.zip')as z:
 expected={'PageSpindle/'+r['path']for r in rows}|{'PageSpindle/MODELS_INVENTORY.json'}
 if len(z.namelist())!=len(expected)or set(z.namelist())!=expected:raise ValueError('Common model archive inventory mismatch')
 for r in rows:
  with z.open('PageSpindle/'+r['path'])as f:digest=hashlib.file_digest(f,'sha256').hexdigest()
  if digest!=r['sha256']or z.getinfo('PageSpindle/'+r['path']).file_size!=r['bytes']:raise ValueError('Stale common component '+r['path'])
print('Packaged and verified common models and corresponding-source archives',flush=True)
