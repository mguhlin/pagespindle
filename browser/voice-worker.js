import * as ort from 'onnxruntime-web/wasm';
import {phonemize} from '../third_party/espeak-ng/phonemizer.js';
ort.env.wasm.numThreads=1;ort.env.wasm.proxy=false;ort.env.wasm.wasmPaths=new URL('./',self.location.href).href;
let session=null,config=null;
self.onmessage=async({data})=>{const{id,kind}=data;try{if(kind==='load'){config=data.config;session=await ort.InferenceSession.create(data.model,{executionProviders:['wasm'],graphOptimizationLevel:'all',logSeverityLevel:4});self.postMessage({id,result:{ready:true}});return}
 if(kind!=='speak'||!session||typeof data.text!=='string'||data.text.length>1800)throw Error('Voice is not ready or passage is too long');
 const sentences=await phonemize(data.text,'en-us');const outputs=[];let total=0;
 for(const sentence of sentences){const ids=[1,0];for(const ch of sentence){const values=config.phoneme_id_map[ch];if(!values)throw Error('This passage contains a phoneme the English voice cannot represent');for(const value of values)ids.push(value);ids.push(0)}ids.push(2);
 const feeds={input:new ort.Tensor('int64',BigInt64Array.from(ids,BigInt),[1,ids.length]),input_lengths:new ort.Tensor('int64',new BigInt64Array([BigInt(ids.length)]),[1]),scales:new ort.Tensor('float32',new Float32Array([config.inference.noise_scale,1/data.speed,config.inference.noise_w]),[3])};
 const result=await session.run(feeds);const audio=result[session.outputNames[0]].data;if(!(audio instanceof Float32Array)||audio.length>8000000)throw Error('Voice output exceeded its limit');const pcm=new Uint8Array(audio.length*2),view=new DataView(pcm.buffer);for(let i=0;i<audio.length;i++)view.setInt16(i*2,Math.round(Math.max(-1,Math.min(1,audio[i]))*32767),true);outputs.push(pcm,new Uint8Array(11024));total+=pcm.length+11024;for(const t of Object.values(feeds))t.dispose();for(const t of Object.values(result))t.dispose();if(total>24*1024*1024)throw Error('Voice chunk exceeded its memory limit')}
 const pcm=new Uint8Array(total);let offset=0;for(const part of outputs){pcm.set(part,offset);offset+=part.length}self.postMessage({id,result:{pcm:pcm.buffer,rate:22050}},[pcm.buffer]);
 }catch(e){self.postMessage({id,error:e.message||'Local voice generation failed'})}};
