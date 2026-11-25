#!/usr/bin/env python3
"""Fixed synthetic pilot in the real native webview; no claim of OS picker gestures."""
import argparse,json,os,shutil,subprocess,time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1];ap=argparse.ArgumentParser();ap.add_argument('--target',required=True);ap.add_argument('--executable');a=ap.parse_args();target=a.target
stage=ROOT/'dist'/target/'PageSpindle';exe=Path(a.executable).resolve()if a.executable else stage/('pagespindle.exe'if os.name=='nt'else'pagespindle')
output=ROOT/'.cache'/('native-pilot-'+target)
if output.exists():shutil.rmtree(output)
output.mkdir();log=(ROOT/'.cache'/('native-pilot-'+target+'.log')).open('w')
args=[str(exe),'--models',str(ROOT/'models'),'--native-self-test',str(output)]
if not a.executable:args+=['--assets',str(stage)]
p=subprocess.Popen(args,stdout=log,stderr=subprocess.STDOUT)
try:
 report=output/'native-window-report.json';end=time.monotonic()+900
 while not report.exists()and p.poll()is None and time.monotonic()<end:time.sleep(.5)
 if not report.exists():raise RuntimeError('Native window pilot produced no report; see native-pilot log')
 data=json.loads(report.read_text());print(json.dumps(data),flush=True)
 if not data['ok']:raise RuntimeError(data['message'])
 p.wait(timeout=30)
 if p.returncode:raise RuntimeError('Native app did not close cleanly')
 for name in ['PageSpindle-project.json','My Listening Edition - Audiobook.zip','My Listening Edition.m4b']:
  if not(output/name).is_file()or(output/name).stat().st_size<100:raise RuntimeError('Missing actual native output '+name)
finally:
 if p.poll()is None:p.kill();p.wait(timeout=10)
 log.close()
