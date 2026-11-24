/* PageSpindle native speech adapter. GPL-3.0-or-later.
 * Reads bounded UTF-8 narration from stdin so private text is not in argv.
 * The owner sets the voice working directory; the only argument is speed.
 * Binary WAV goes to stdout, so filenames stay in Go's Unicode-aware I/O. */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "sherpa-onnx/c-api/c-api.h"
#ifdef _WIN32
#include <io.h>
#include <fcntl.h>
#endif
static void u16(unsigned int x){fputc(x&255,stdout);fputc((x>>8)&255,stdout);}
static void u32(unsigned int x){u16(x&65535);u16(x>>16);}
int main(int argc,char **argv) {
 if(argc!=2)return 2;
#ifdef _WIN32
 _setmode(_fileno(stdin),_O_BINARY);_setmode(_fileno(stdout),_O_BINARY);
#endif
 const size_t limit=32768;char *text=(char*)calloc(limit+2,1);if(!text)return 3;
 size_t n=fread(text,1,limit+1,stdin);if(n==0||n>limit||ferror(stdin)){free(text);return 4;}
 float speed=(float)atof(argv[1]);if(speed<0.6f||speed>1.6f){free(text);return 5;}
 SherpaOnnxOfflineTtsConfig config;memset(&config,0,sizeof(config));
 config.model.vits.model="en_US-ljspeech-high.onnx";config.model.vits.tokens="tokens.txt";config.model.vits.data_dir="espeak-ng-data";
 config.model.vits.noise_scale=0.667f;config.model.vits.noise_scale_w=0.8f;config.model.vits.length_scale=1.0f;
 config.model.num_threads=2;config.model.provider="cpu";config.max_num_sentences=1;config.silence_scale=0.2f;
 const SherpaOnnxOfflineTts *tts=SherpaOnnxCreateOfflineTts(&config);if(!tts){free(text);return 6;}
 SherpaOnnxGenerationConfig generation;memset(&generation,0,sizeof(generation));generation.speed=speed;generation.silence_scale=0.2f;
 const SherpaOnnxGeneratedAudio *audio=SherpaOnnxOfflineTtsGenerateWithConfig(tts,text,&generation,NULL,NULL);
 free(text);int ok=0;
 if(audio&&audio->n>0&&audio->n<22050*15*60){unsigned int bytes=(unsigned int)audio->n*2;fwrite("RIFF",1,4,stdout);u32(36+bytes);fwrite("WAVEfmt ",1,8,stdout);u32(16);u16(1);u16(1);u32(audio->sample_rate);u32(audio->sample_rate*2);u16(2);u16(16);fwrite("data",1,4,stdout);u32(bytes);for(int i=0;i<audio->n;i++){float x=audio->samples[i];if(x>1)x=1;if(x< -1)x= -1;short sample=(short)(x*32767);u16((unsigned short)sample);}ok=fflush(stdout)==0&&!ferror(stdout);}
 if(audio)SherpaOnnxDestroyOfflineTtsGeneratedAudio(audio);SherpaOnnxDestroyOfflineTts(tts);return ok?0:7;
}
