#!/usr/bin/env python3
"""Native build and deterministic platform ZIP. Run on the declared target OS/CPU."""
import argparse, hashlib, json, os, platform, shutil, stat, subprocess, zipfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
def run(args,**kwargs):subprocess.run([str(x) for x in args],check=True,cwd=ROOT,**kwargs)
def main():
 ap=argparse.ArgumentParser();ap.add_argument('--target',required=True);a=ap.parse_args();target=a.target
 native={'Windows':'windows','Darwin':'darwin','Linux':'linux'}[platform.system()]+'-'+{'x86_64':'amd64','AMD64':'amd64','arm64':'arm64','aarch64':'arm64'}[platform.machine()]
 if target!=native:raise SystemExit('Native speech build requires '+target+'; this runner is '+native)
 lock=json.loads((ROOT/'dependencies.json').read_text());speech=ROOT/'runtime'/target/'speech';source=speech/lock['targets'][target]['speech']['name'].removesuffix('.tar.bz2');worker=speech/'worker';worker.mkdir(exist_ok=True);exe='.exe' if target.startswith('windows') else '';out=worker/('pagespindle-speech'+exe)
 if target.startswith('windows'):
  vswhere=Path(os.environ.get('ProgramFiles(x86)',r'C:\Program Files (x86)'))/'Microsoft Visual Studio/Installer/vswhere.exe'
  install=subprocess.check_output([str(vswhere),'-latest','-products','*','-requires','Microsoft.VisualStudio.Component.VC.Tools.x86.x64','-property','installationPath'],text=True).strip();vcvars=Path(install)/'VC/Auxiliary/Build/vcvars64.bat'
  cmd=ROOT/'.cache/build-speech.cmd';cmd.write_text('@echo off\ncall "'+str(vcvars)+'"\nif errorlevel 1 exit /b %errorlevel%\ncl /nologo /O2 /I"'+str(source/'include')+'" "'+str(ROOT/'native/speech-worker.c')+'" /Fe:"'+str(out)+'" /Fo:"'+str(worker/'speech.obj')+'" /link /LIBPATH:"'+str(source/'lib')+'" sherpa-onnx-c-api.lib\n')
  run(['cmd.exe','/d','/c',cmd])
  for dll in source.rglob('*.dll'):shutil.copy2(dll,worker/dll.name)
 else:
  origin='@loader_path' if target.startswith('darwin') else '$ORIGIN';rpath=origin+'/../'+source.name+'/lib'
  run(['cc','-O2',ROOT/'native/speech-worker.c','-I',source/'include','-L',source/'lib','-Wl,-rpath,'+rpath,'-lsherpa-onnx-c-api','-o',out])
 stage=ROOT/'dist'/target/'PageSpindle'
 if stage.exists():shutil.rmtree(stage)
 stage.mkdir(parents=True);env=dict(os.environ,CGO_ENABLED='0');run(['go','build','-trimpath','-ldflags=-s -w','-o',stage/('pagespindle-worker'+exe),'./cmd/pagespindle'],env=env)
 env=dict(os.environ,CGO_ENABLED='1');flags='-s -w'+(' -H windowsgui'if target.startswith('windows')else'');run(['go','build','-tags','desktop,production','-trimpath','-ldflags='+flags,'-o',stage/('pagespindle'+exe),'./cmd/pagespindle'],env=env)
 runtime_dest=stage/'runtime'/target
 # Retain the pinned distributions, their loaders/resources and native library modes.
 shutil.copytree(ROOT/'runtime'/target,runtime_dest,symlinks=True)
 for role in ['llama','llama-gpu','speech']:
  for p in (runtime_dest/role).rglob('*'):
   if p.is_file() and not p.is_symlink():
    if role.startswith('llama') and ((p.suffix.lower()=='.exe' and p.name!='llama-server.exe') or (not target.startswith('windows') and p.name.startswith('llama-') and '.' not in p.name and p.name!='llama-server')):p.unlink()
    elif role=='speech' and p.parent.name=='bin' and p.suffix.lower()!='.dll':p.unlink()
    elif p.suffix in {'.obj','.lib','.a','.h','.hpp'}:p.unlink()
 for name in ['LICENSE','README.md','PROTECTIONS.md','THIRD_PARTY.md','VALIDATION.md','dependencies.json']:shutil.copy2(ROOT/name,stage/name)
 shutil.copytree(ROOT/'licenses',stage/'licenses')
 if target.startswith('windows'):pass
 else:
  launcher=stage/('Start PageSpindle.command' if target.startswith('darwin') else 'Start PageSpindle.sh');launcher.write_text('#!/bin/sh\ncd -- "$(dirname -- "$0")" || exit 1\nexec ./pagespindle\n');launcher.chmod(0o755)
 manifest=[]
 for p in sorted(stage.rglob('*')):
  if p.is_symlink():manifest.append({'path':p.relative_to(stage).as_posix(),'symlink':os.readlink(p)})
  elif p.is_file():manifest.append({'path':p.relative_to(stage).as_posix(),'sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'bytes':p.stat().st_size})
 (stage/'INVENTORY.json').write_text(json.dumps({'schema':1,'target':target,'source_revision':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),'go_toolchain':subprocess.check_output(['go','version'],text=True).strip(),'files':manifest},indent=2)+'\n')
 archive=ROOT/'dist'/('pagespindle-'+target+'.zip')
 with zipfile.ZipFile(archive,'w',zipfile.ZIP_DEFLATED,compresslevel=6) as z:
  for p in sorted(stage.rglob('*')):
   if p.is_dir() and not p.is_symlink():continue
   info=zipfile.ZipInfo('PageSpindle/'+p.relative_to(stage).as_posix(),date_time=(1980,1,1,0,0,0));info.create_system=3;info.external_attr=p.lstat().st_mode<<16;info.compress_type=zipfile.ZIP_DEFLATED
   z.writestr(info,os.readlink(p).encode() if p.is_symlink() else p.read_bytes())
 print('Packaged',target,archive.stat().st_size,'bytes')
if __name__=='__main__':main()
