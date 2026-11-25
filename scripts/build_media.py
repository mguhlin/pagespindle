#!/usr/bin/env python3
"""Build the two required encoders from pinned sources, without network protocols."""
import argparse,hashlib,json,os,platform,shlex,shutil,subprocess,tarfile,urllib.request
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1];p=argparse.ArgumentParser();p.add_argument('--target',required=True);a=p.parse_args();target=a.target;lock=json.loads((ROOT/'native/media-lock.json').read_text());cache=ROOT/'.cache';cache.mkdir(exist_ok=True)
for item in lock.values():
 path=cache/item['name']
 if not path.exists():
  with urllib.request.urlopen(item['url'],timeout=90)as r,path.open('wb')as f:shutil.copyfileobj(r,f)
 with path.open('rb')as f:d=hashlib.file_digest(f,'sha256').hexdigest()
 if d!=item['sha256']or path.stat().st_size!=item['size']:raise ValueError('Media source integrity failed')
work=cache/('media-'+target);work.mkdir(exist_ok=True);prefix=work/'prefix';dest=ROOT/'runtime'/target/'media';dest.mkdir(parents=True,exist_ok=True)
for name in lock:
 folder=work/('ffmpeg-8.0.1'if name=='ffmpeg'else'lame-3.100')
 if not folder.exists():
  with tarfile.open(cache/lock[name]['name'])as archive:archive.extractall(work,filter='data')
def run(args,cwd):
 env={k:v for k,v in os.environ.items()if k.upper()in {'PATH','SYSTEMROOT','WINDIR','COMSPEC','TMP','TEMP','TMPDIR','LANG','LC_ALL','MACOSX_DEPLOYMENT_TARGET','HOME'}}
 if target.startswith('windows'):
  bash=Path('C:/msys64/usr/bin/bash.exe');env['MSYSTEM']='MINGW64';env['CHERE_INVOKING']='1';cmd='export PATH=/mingw64/bin:/usr/bin; cd '+shlex.quote(cwd.as_posix())+'; '+shlex.join([str(x).replace('\\','/')for x in args]);subprocess.run([str(bash),'--login','-c',cmd],env=env,check=True)
 else:subprocess.run([str(x)for x in args],cwd=cwd,env=env,check=True)
lame=work/'lame-3.100';ff=work/'ffmpeg-8.0.1'
if not (prefix/'lib/libmp3lame.a').exists():
 opts=['sh','./configure','--disable-shared','--enable-static','--disable-frontend','--prefix='+prefix.as_posix()]
 if target.startswith('windows'):opts+=['--host=x86_64-w64-mingw32']
 run(opts,lame);run(['make','-j2'],lame);run(['make','install'],lame)
args=['sh','./configure','--disable-everything','--disable-autodetect','--disable-doc','--disable-debug','--disable-network','--disable-ffplay','--disable-ffprobe','--enable-ffmpeg','--enable-small','--disable-shared','--enable-static','--disable-x86asm','--disable-iconv','--disable-zlib','--disable-bzlib','--disable-lzma','--enable-libmp3lame','--enable-protocol=file,pipe','--enable-demuxer=wav,concat,ffmetadata','--enable-muxer=mp3,mp4,ipod,ffmetadata','--enable-decoder=pcm_s16le','--enable-encoder=libmp3lame,aac','--enable-filter=aresample,anull,abuffer,abuffersink','--extra-cflags=-I'+(prefix/'include').as_posix(),'--extra-ldflags=-L'+(prefix/'lib').as_posix()]
if target.startswith('windows'):args+=['--target-os=mingw32','--arch=x86_64','--extra-libs=-static-libgcc']
run(args,ff);run(['make','-j2'],ff);name='ffmpeg.exe'if target.startswith('windows')else'ffmpeg';shutil.copy2(ff/name,dest/name)
for name,source in [('FFmpeg-LGPL-2.1.txt',ff/'COPYING.LGPLv2.1'),('LAME-LGPL.txt',lame/'COPYING')]:shutil.copy2(source,ROOT/'licenses'/name)
(dest/'BUILD.json').write_text(json.dumps({'target':target,'sources':lock,'configure':args,'sha256':hashlib.sha256((dest/('ffmpeg.exe'if target.startswith('windows')else'ffmpeg')).read_bytes()).hexdigest()},indent=2)+'\n')
print('Built offline AAC/MP3 media converter for',target)
