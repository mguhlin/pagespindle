#!/usr/bin/env python3
"""Install/upgrade on an isolated hosted runner, exercise installed window, uninstall.
Runner OS runtimes may already exist: this is not a clean physical OS qualification.
"""
import argparse,os,shutil,subprocess
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1];ap=argparse.ArgumentParser();ap.add_argument('--target',required=True);target=ap.parse_args().target
work=ROOT/'.cache'/('installer-check-'+target);work.mkdir(parents=True,exist_ok=True)
sentinel=work/'preserved-user-file.txt';sentinel.write_text('preserve this user file')
def run(args):subprocess.run([str(x)for x in args],check=True,cwd=ROOT)
if target.startswith('windows'):
 installer=ROOT/'dist/pagespindle-windows-amd64-setup.exe';dest=work/'PageSpindle'
 install=lambda:run([installer,'/VERYSILENT','/SUPPRESSMSGBOXES','/NORESTART','/SP-','/DIR='+str(dest),'/LOG='+str(work/'setup.log')])
 exe=dest/'pagespindle.exe';uninstall=lambda:run([dest/'unins000.exe','/VERYSILENT','/SUPPRESSMSGBOXES','/NORESTART'])
elif target.startswith('darwin'):
 installer=ROOT/'dist'/('pagespindle-'+target+'.pkg');dest=Path('/Applications/PageSpindle.app')
 if dest.exists():raise RuntimeError('Refusing to replace an existing app before isolated installer test')
 install=lambda:run(['sudo','installer','-pkg',installer,'-target','/']);exe=dest/'Contents/MacOS/pagespindle';uninstall=lambda:run(['sudo','rm','-rf',str(dest)])
else:
 installer=ROOT/'dist/pagespindle-linux-amd64.deb';dest=Path('/opt/pagespindle')
 if dest.exists():raise RuntimeError('Refusing to replace an existing app before isolated installer test')
 install=lambda:run(['sudo','apt-get','install','-y',installer]);exe=dest/'pagespindle';uninstall=lambda:run(['sudo','apt-get','remove','-y','pagespindle'])
try:
 install();install() # upgrade/reinstall must preserve user data outside installation.
 run([os.sys.executable,ROOT/'scripts/native_window_check.py','--target',target,'--executable',exe])
finally:
 if exe.exists():uninstall()
if sentinel.read_text()!='preserve this user file':raise RuntimeError('User recovery sentinel changed')
print('PASS installed/upgrade/native-window/uninstall:',target,'; shared OS runtimes and user models intentionally retained')
