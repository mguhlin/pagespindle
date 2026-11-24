# Components, attribution, and redistribution

PageSpindle code is GPL-3.0-or-later. It preserves compatible third-party notices rather than claiming ownership of their work. Original documentation and generated illustrations are CC BY 4.0 by Miguel Guhlin. Model, engine, voice, and source-book rights are distinct.

| Component | Identity | Terms and source |
| --- | --- | --- |
| Qwen3-1.7B Q8 GGUF | Official Qwen revision `90862c4b9d2787eaed51d12237eafdfe7c5f6077`; file digest in `dependencies.json` | Apache-2.0, full notice in `licenses/`. [Model repository](https://huggingface.co/Qwen/Qwen3-1.7B-GGUF). |
| llama.cpp | `b12681`, platform archive hashes in the lock | MIT; full notice retained. [Source](https://github.com/ggml-org/llama.cpp/tree/b12681). |
| sherpa-onnx | `v1.12.35`, platform archive hashes in the lock | Apache-2.0 source with a GPL-containing speech dependency chain; distribute the speech adapter under GPL-3.0-or-later. [Source](https://github.com/k2-fsa/sherpa-onnx/tree/v1.12.35). |
| piper-phonemize | `78a788e0b719013401572d70fef372e77bff8e43` | MIT; full notice retained. [Source](https://github.com/csukuangfj/piper-phonemize/tree/78a788e0b719013401572d70fef372e77bff8e43). |
| eSpeak NG | Speech dependency pinned by piper-phonemize to `7251208ee427ef87e94f39d0e99b15d287224ed6` | GPL-3.0-or-later; full notice retained. [Corresponding source](https://github.com/csukuangfj/espeak-ng/tree/7251208ee427ef87e94f39d0e99b15d287224ed6). |
| ONNX Runtime | `1.23.2` in the pinned speech distribution | MIT; full notice and third-party notices recovered from Microsoft's official PyPI wheel, with artifact digest in `licenses/onnxruntime-license-source.json`. |
| LJ Speech voice | `vits-piper-en_US-ljspeech-high-int8`; artifact digest in the lock | Model card identifies public-domain LJ Speech data and training from scratch. Retain its model card and [dataset context](https://keithito.com/LJ-Speech-Dataset/). The packaged model also contains eSpeak data subject to its own terms. |
| Go PDF reader and HTML parser | Exact module versions in `go.mod` and `go.sum` | BSD notices preserved in the vendored source and `licenses/`. |
| Optional user-installed FFmpeg | Detected on the owner's executable search path, not bundled | Each installation retains its own build/license terms. WAV output requires no FFmpeg. |

The speech/runtime source archive accompanies binary releases, including the pinned sherpa-onnx, piper-phonemize, eSpeak source and their build instructions. The dependency lock identifies original published hashes. Review these obligations when changing components; a top-level engine license does not override dependency or voice terms. Source availability and license review are separate from a claim of security or legal clearance.

The text-preparation method is credited to **Lance Eaton's Book to Audio Edition**, from [his repository](https://github.com/leaton79/book-to-audiobook) and [blog explanation](https://aiedusimplified.substack.com/p/ai-audiobooks-an-update), adapted by Miguel Guhlin for VCC/PageSpindle. The app's preparation prompt is an original conservative implementation of that workflow. VCC's [downloadable skill](https://mguhlin.github.io/vcc/skills/audiobook-maker/SKILL.md) retains Eaton's full MIT notice. No book content, third-party branding, or consent to publish converted books is included.

Generated splash/icons use built-in OpenAI image generation. Full prompts are in `ARTWORK_PROMPTS.json`, originals in `artwork/`, and optimized versions in the app and Pages assets. They illustrate a workflow rather than certify product behavior.
