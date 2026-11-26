#!/usr/bin/env python3
"""Package a native build into an OS installer; models are installed by first launch."""
import argparse,hashlib,json,os,platform,shutil,subprocess,urllib.request
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
a=argparse.ArgumentParser();a.add_argument('--target',required=True);target=a.parse_args().target
stage=ROOT/'dist'/target/'PageSpindle';dist=ROOT/'dist'
def run(args,**kw):subprocess.run([str(x)for x in args],check=True,cwd=ROOT,**kw)
if target.startswith('windows'):
 lock=json.loads((ROOT/'native/windows-runtime-lock.json').read_text())
 for row in lock.values():
  p=ROOT/'.cache'/row['name']
  if not p.exists():
   with urllib.request.urlopen(row['url'],timeout=120)as r,p.open('wb')as f:shutil.copyfileobj(r,f)
  with p.open('rb')as f:d=hashlib.file_digest(f,'sha256').hexdigest()
  if d!=row['sha256']or p.stat().st_size!=row['size']:raise ValueError('Runtime checksum mismatch')
  command="$s=Get-AuthenticodeSignature -LiteralPath '"+str(p).replace("'","''")+"'; if($s.Status -ne 'Valid' -or $s.SignerCertificate.Subject -notmatch 'O=Microsoft Corporation'){throw 'Microsoft signature validation failed'}; $s.SignerCertificate.Subject"
  run(['pwsh','-NoProfile','-NonInteractive','-Command',command])
 iscc=Path(os.environ.get('ProgramFiles(x86)',r'C:\Program Files (x86)'))/'Inno Setup 6/ISCC.exe'
 run([iscc,'/DSourceRoot='+str(ROOT),ROOT/'packaging/windows/pagespindle.iss'])
elif target.startswith('darwin'):
 payload=dist/target/'pkgroot';app=payload/'Applications/PageSpindle.app';contents=app/'Contents';mac=contents/'MacOS';resources=contents/'Resources'
 if payload.exists():shutil.rmtree(payload)
 mac.mkdir(parents=True);resources.mkdir(parents=True)
 shutil.copy2(stage/'pagespindle',mac/'pagespindle')
 for item in stage.iterdir():
  if item.name=='pagespindle':continue
  if item.is_dir():shutil.copytree(item,resources/item.name,symlinks=True)
  else:shutil.copy2(item,resources/item.name)
 iconset=dist/target/'PageSpindle.iconset';iconset.mkdir(exist_ok=True)
 for size in [16,32,128,256,512]:
  for scale in [1,2]:run(['sips','-z',str(size*scale),str(size*scale),ROOT/'artwork/save.png','--out',iconset/('icon_'+str(size)+'x'+str(size)+('@2x'if scale==2 else'')+'.png')])
 run(['iconutil','-c','icns',iconset,'-o',resources/'PageSpindle.icns'])
 import plistlib
 with (contents/'Info.plist').open('wb')as f:plistlib.dump({'CFBundleIconFile':'PageSpindle.icns','CFBundleIdentifier':'org.mguhlin.pagespindle','CFBundleName':'PageSpindle','CFBundleDisplayName':'PageSpindle','CFBundleExecutable':'pagespindle','CFBundlePackageType':'APPL','CFBundleShortVersionString':'0.3.0','CFBundleVersion':'3','LSMinimumSystemVersion':'15.0','NSHighResolutionCapable':True},f)
 # Ad-hoc integrity only: no Developer ID, no notarization claim.
 run(['codesign','--force','--deep','--sign','-',app]);run(['codesign','--verify','--deep','--strict',app])
 rows=[]
 for p in sorted(app.rglob('*')):
  if p.is_file()and not p.is_symlink():
   with p.open('rb')as f:d=hashlib.file_digest(f,'sha256').hexdigest()
   rows.append({'path':p.relative_to(app).as_posix(),'sha256':d,'bytes':p.stat().st_size})
 # This auxiliary receipt is outside the signed app; adding resources would invalidate its seal.
 (dist/('pagespindle-'+target+'-signed-inventory.json')).write_text(json.dumps(rows,indent=2)+'\n')
 run(['pkgbuild','--root',payload,'--identifier','org.mguhlin.pagespindle','--version','0.3.0','--install-location','/',dist/('pagespindle-'+target+'.pkg')])
elif target=='linux-amd64':
 payload=dist/target/'debroot'
 if payload.exists():shutil.rmtree(payload)
 app=payload/'opt/pagespindle';app.parent.mkdir(parents=True);shutil.copytree(stage,app,symlinks=True)
 desktop=payload/'usr/share/applications';desktop.mkdir(parents=True);(desktop/'pagespindle.desktop').write_text('[Desktop Entry]\nType=Application\nName=PageSpindle\nComment=Spin text into listening\nExec=/opt/pagespindle/pagespindle\nIcon=pagespindle\nTerminal=false\nCategories=AudioVideo;Audio;Education;\n')
 icon=payload/'usr/share/icons/hicolor/256x256/apps';icon.mkdir(parents=True);shutil.copy2(ROOT/'artwork/save.png',icon/'pagespindle.png')
 control=payload/'DEBIAN';control.mkdir();(control/'control').write_text('Package: pagespindle\nVersion: 0.3.0\nArchitecture: amd64\nMaintainer: Miguel Guhlin <mguhlin@users.noreply.github.com>\nDepends: libc6 (>= 2.39), libgtk-4-1, libwebkitgtk-6.0-4, libstdc++6, libgomp1, libvulkan1, libssl3t64, libasound2t64, ca-certificates, gstreamer1.0-plugins-base, gstreamer1.0-plugins-good\nSection: sound\nPriority: optional\nHomepage: https://mguhlin.github.io/pagespindle/\nDescription: Local text preparation and audiobook maker\n Native window, local AI, speech and bundled media converter.\n First launch downloads verified models once; then works offline.\n')
 run(['dpkg-deb','--root-owner-group','--build',payload,dist/'pagespindle-linux-amd64.deb'])
else:raise ValueError('Unsupported installer target')
print('Packaged native installer:',target)
