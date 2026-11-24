#!/usr/bin/env python3
"""Test the extracted release executable and its pruned runtimes, not the source tree."""
import hashlib,io,json,os,posixpath,re,shutil,signal,stat,subprocess,sys,time,urllib.request,wave,zipfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1];target=sys.argv[1]
work=ROOT/'.cache'/('installed-smoke-'+target)
if work.exists():shutil.rmtree(work)
work.mkdir(parents=True);temporary=work/'tmp';temporary.mkdir()
with zipfile.ZipFile(ROOT/'dist'/('pagespindle-'+target+'.zip'))as z:
 inventory=json.loads(z.read('PageSpindle/INVENTORY.json'));expected={'PageSpindle/INVENTORY.json'}
 for row in inventory['files']:
  name='PageSpindle/'+row['path'];expected.add(name);data=z.read(name)
  if 'sha256'in row:
   if hashlib.sha256(data).hexdigest()!=row['sha256']:raise ValueError('Packaged component hash mismatch')
  elif data.decode()!=row['symlink']:raise ValueError('Packaged link mismatch')
 if len(z.namelist())!=len(expected)or set(z.namelist())!=expected:raise ValueError('Package inventory mismatch')
 for info in z.infolist():
  if info.filename.startswith('/')or'..'in Path(info.filename).parts:raise ValueError('Unsafe archive member')
  path=work/info.filename;path.parent.mkdir(parents=True,exist_ok=True);mode=info.external_attr>>16
  if stat.S_ISLNK(mode):
   link=z.read(info).decode()
   if link.startswith('/')or not posixpath.normpath(posixpath.join(posixpath.dirname(info.filename),link)).startswith('PageSpindle/'):raise ValueError('Unsafe package link')
   path.symlink_to(link)
  else:
   path.write_bytes(z.read(info))
   if os.name!='nt':path.chmod(mode&0o777)
installed=work/'PageSpindle';shutil.copytree(ROOT/'models',installed/'models')
exe=installed/('pagespindle.exe'if os.name=='nt'else'pagespindle')
env=dict(os.environ,TMPDIR=str(temporary),TMP=str(temporary),TEMP=str(temporary));proc=subprocess.Popen([str(exe),'--no-browser'],cwd=installed,env=env,stdout=subprocess.PIPE,stderr=subprocess.DEVNULL,text=True,encoding='utf-8')
api=None
try:
 for _ in range(5):
  line=proc.stdout.readline();match=re.search(r'(http://127\.0\.0\.1:\d+)/#([a-f0-9]{64})',line)
  if match:base,key=match.groups();break
 else:raise RuntimeError('Extracted launcher did not start')
 client=urllib.request.build_opener(urllib.request.ProxyHandler({}))
 def call(path,value=None,binary=False):
  data=None if value is None else json.dumps(value).encode();request=urllib.request.Request(base+'/api/'+path,data=data,headers={'X-PageSpindle-Key':key,'Origin':base,'Content-Type':'application/json'})
  with client.open(request,timeout=15)as response:body=response.read(8*1024*1024)
  return body if binary else json.loads(body)
 api=call
 def finish():
  deadline=time.monotonic()+420
  while time.monotonic()<deadline:
   job=call('job')
   if job['state']=='complete':return job
   if job['state']in ['failed','cancelled']:raise RuntimeError(job.get('message','Package job failed'))
   time.sleep(.3)
  raise TimeoutError('Extracted package operation exceeded seven minutes')
 source='The fictional library opens at 9:00 and closes at 5:00. Keep the well-known path.'
 document=call('new',{'text':source,'title':'Package smoke'})
 call('prepare',document);prepared=finish()['document']
 assert prepared['source']==source and prepared['blocks'][0]['prepared'].strip()==source
 assert not any(block['reviewed']for block in prepared['blocks'])
 for block in prepared['blocks']:block['reviewed']=True
 call('render',{'document':prepared,'speed':1,'format':'wav','preview':True});job=finish()
 audio=call('download',{'id':job['id'],'name':job['files'][0]},binary=True)
 with wave.open(io.BytesIO(audio),'rb')as reader:assert reader.getframerate()==22050 and reader.getnchannels()==1 and reader.getnframes()>22050
 call('clear',{})
 print('PASS: extracted',target,'release executable, actual local AI, unreviewed proposal, real narration, allowed download and clear; source',inventory['source_revision'][:7],flush=True)
finally:
 if api:
  try:api('clear',{})
  except Exception:pass
 if proc.poll()is None:
  if os.name=='nt':proc.terminate()
  else:proc.send_signal(signal.SIGINT)
  try:proc.wait(timeout=15)
  except subprocess.TimeoutExpired:proc.kill();proc.wait(timeout=5)
