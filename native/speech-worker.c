/* PageSpindle native speech adapter. GPL-3.0-or-later.
 * Reads bounded UTF-8 narration from stdin so private text is not in argv.
 * Arguments are owner-issued model, tokens, data, output paths and speed. */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "sherpa-onnx/c-api/c-api.h"
int main(int argc,char **argv) {
 if(argc!=6)return 2;
 const size_t limit=32768;char *text=(char*)calloc(limit+2,1);if(!text)return 3;
 size_t n=fread(text,1,limit+1,stdin);if(n==0||n>limit||ferror(stdin)){free(text);return 4;}
 float speed=(float)atof(argv[5]);if(speed<0.6f||speed>1.6f){free(text);return 5;}
 SherpaOnnxOfflineTtsConfig config;memset(&config,0,sizeof(config));
 config.model.vits.model=argv[1];config.model.vits.tokens=argv[2];config.model.vits.data_dir=argv[3];
 config.model.vits.noise_scale=0.667f;config.model.vits.noise_scale_w=0.8f;config.model.vits.length_scale=1.0f;
 config.model.num_threads=2;config.model.provider="cpu";config.max_num_sentences=1;config.silence_scale=0.2f;
 const SherpaOnnxOfflineTts *tts=SherpaOnnxCreateOfflineTts(&config);if(!tts){free(text);return 6;}
 SherpaOnnxGenerationConfig generation;memset(&generation,0,sizeof(generation));generation.speed=speed;generation.silence_scale=0.2f;
 const SherpaOnnxGeneratedAudio *audio=SherpaOnnxOfflineTtsGenerateWithConfig(tts,text,&generation,NULL,NULL);
 free(text);int ok=0;if(audio&&audio->n>0)ok=SherpaOnnxWriteWave(audio->samples,audio->n,audio->sample_rate,argv[4]);
 if(audio)SherpaOnnxDestroyOfflineTtsGeneratedAudio(audio);SherpaOnnxDestroyOfflineTts(tts);return ok?0:7;
}
