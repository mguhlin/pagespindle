#!/usr/bin/env python3
"""Build optional Mac Intel Metal engine from the retained, pinned source.
Windows/Linux Vulkan binaries are hash-locked in dependencies.json. CPU engines
remain separately packaged and usable without a GPU or a GPU software toolkit.
"""
import argparse,hashlib,json,os,shutil,subprocess,tarfile,urllib.request
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1];ap=argparse.ArgumentParser();ap.add_argument('--target',required=True);target=ap.parse_args().target
if not target.startswith('darwin'):raise SystemExit(0)
def run(args,**kw):subprocess.run([str(x)for x in args],check=True,cwd=ROOT,**kw)
probe=ROOT/'runtime'/target/'gpu-support';probe.mkdir(parents=True,exist_ok=True);run(['cc','-O2',ROOT/'native/metal-check.m','-framework','Foundation','-framework','Metal','-o',probe/'pagespindle-metal-check'])
if target=='darwin-arm64':raise SystemExit(0) # pinned upstream already includes Metal
row=next(r for r in json.loads((ROOT/'SOURCE_ARCHIVES.json').read_text())if r['file']=='llama.cpp-b12681.tar.gz');cache=ROOT/'.cache';archive=cache/row['file'];retained=cache/'corresponding-source'/row['file']
if not archive.exists():
 if retained.exists():shutil.copy2(retained,archive)
 else:
  with urllib.request.urlopen(row['source'],timeout=90)as r,archive.open('wb')as f:shutil.copyfileobj(r,f)
with archive.open('rb')as f:d=hashlib.file_digest(f,'sha256').hexdigest()
if d!=row['sha256']or archive.stat().st_size!=row['bytes']:raise ValueError('Metal engine source identity mismatch')
work=cache/'metal-source';work.mkdir(exist_ok=True)
if not(work/'llama.cpp-b12681').exists():
 with tarfile.open(archive)as t:t.extractall(work,filter='data')
source=work/'llama.cpp-b12681';build=cache/'metal-build'
# New Xcode images may deliver the compiler as a separate Apple SDK component.
if subprocess.run(['xcrun','--find','metal'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL).returncode:run(['xcodebuild','-downloadComponent','MetalToolchain'])
flags=['-DCMAKE_BUILD_TYPE=Release','-DCMAKE_OSX_DEPLOYMENT_TARGET=15.0','-DGGML_METAL=ON','-DGGML_METAL_EMBED_LIBRARY=ON','-DCMAKE_BUILD_WITH_INSTALL_RPATH=ON','-DCMAKE_INSTALL_RPATH=@loader_path','-DCMAKE_INSTALL_NAME_DIR=@rpath','-DGGML_NATIVE=OFF','-DGGML_OPENMP=OFF','-DLLAMA_BUILD_TESTS=OFF','-DLLAMA_BUILD_SERVER=ON']
run(['cmake','-S',source,'-B',build,*flags]);run(['cmake','--build',build,'--config','Release','--target','llama-server','-j','3'])
dest=ROOT/'runtime'/target/'llama-gpu'
if dest.exists():shutil.rmtree(dest)
shutil.copytree(build/'bin',dest,symlinks=True)
for p in dest.iterdir():
 if not p.is_file()or p.is_symlink()or p.suffix not in ['','.dylib']:continue
 lines=subprocess.check_output(['otool','-L',p],text=True).splitlines()[1:]
 for line in lines:
  dependency=line.strip().split(' (')[0]
  if not dependency.startswith(('@rpath/','@loader_path/','/System/Library/','/usr/lib/')):raise RuntimeError('Nonportable Metal runtime dependency: '+dependency)
shutil.copy2(source/'LICENSE',dest/'LICENSE')
(dest/'BUILD.json').write_text(json.dumps({'source':row,'configure':flags,'backend':'Metal'},indent=2)+'\n')
print('Built optional Metal backend and safe hardware probe for',target)
