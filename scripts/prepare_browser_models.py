#!/usr/bin/env python3
"""Record immutable browser model locations and SRI from upstream artifacts."""
import base64,hashlib,json,urllib.request
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
def fetch(u):return urllib.request.urlopen(u,timeout=90).read()
def sri(b):return 'sha256-'+base64.b64encode(hashlib.sha256(b).digest()).decode()
previous=json.loads((ROOT/'browser/browser-models.json').read_text());retired=previous.get('retired_text',[])
records=[];librev='5b7251f3bdce957be2a7bf73a6f6c560d9874df4'
for size,precision,rev in [('1.7B','q4f32_1','bddd4d584cabe19113f7e4ff46fd9e73b4d3dc89'),('1.7B','q4f16_1','80b3abcec6c3b3f5355dc0cc99cc4fb578f192bc')]:
 name='Qwen3-'+size+'-'+precision+'-MLC';url='https://huggingface.co/mlc-ai/'+name+'/resolve/'+rev+'/'
 config=fetch(url+'mlc-chat-config.json');d=json.loads(config);lib='https://raw.githubusercontent.com/mlc-ai/binary-mlc-llm-libs/'+librev+'/web-llm-models/v0_2_87/base/Qwen3-'+size+'-'+precision+'_cs1k-webgpu.wasm'
 integrity={'config':sri(config),'model_lib':sri(fetch(lib)),'tokenizer':{},'onFailure':'error'}
 for filename in d.get('tokenizer_files',[]):integrity['tokenizer'][filename]=sri(fetch(url+filename))
 records.append({'id':name,'url':url,'lib':lib,'integrity':integrity});print('Pinned',name,flush=True)
model=ROOT/'models/vits-piper-en_US-ljspeech-high-int8/en_US-ljspeech-high.onnx';config=model.with_suffix('.onnx.json')
voice={'model':'voice/ljspeech-'+hashlib.sha256(model.read_bytes()).hexdigest()[:16]+'.onnx','sha256':hashlib.sha256(model.read_bytes()).hexdigest(),'bytes':model.stat().st_size,'config':'voice/ljspeech-config.json','config_sha256':hashlib.sha256(config.read_bytes()).hexdigest(),'config_bytes':config.stat().st_size}
(ROOT/'browser/browser-models.json').write_text(json.dumps({'schema':1,'text':records,'voice':voice,'retired_text':retired},indent=2)+'\n')
