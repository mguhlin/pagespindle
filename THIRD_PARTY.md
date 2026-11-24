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

The text-preparation method is credited to **Lance Eaton's Book to Audio Edition**, from [his repository](https://github.com/leaton79/book-to-audiobook) and [blog explanation](https://aiedusimplified.substack.com/p/ai-audiobooks-an-update), adapted by Miguel Guhlin for VCC/PageSpindle. The app's preparation prompt is an original conservative implementation of that workflow. Eaton's full MIT notice is also retained in `licenses/Lance-Eaton-MIT.txt`. VCC's [downloadable skill](https://mguhlin.github.io/vcc/skills/audiobook-maker/SKILL.md) retains Eaton's full MIT notice. No book content, third-party branding, or consent to publish converted books is included.

Generated splash/icons use built-in OpenAI image generation. Full prompts are in `ARTWORK_PROMPTS.json`, originals in `artwork/`, and optimized versions in the app and Pages assets. They illustrate a workflow rather than certify product behavior.

## Browser edition

The browser runs the same Go book/review core as WebAssembly, paired with matching Go `wasm_exec.js`. Text inference uses WebLLM 0.2.89 (Apache-2.0) with the immutable Qwen3 MLC model/library locations and SRI in `browser/browser-models.json`; Qwen3 weights are Apache-2.0. Browser voice inference uses ONNX Runtime Web 1.31.1 (MIT) and the same LJ Speech voice. WAV ZIP creation uses fflate 0.8.3 (MIT). Exact frontend dependencies, including build/test tools, are in `package-lock.json`; their available full notices and package identities are retained in `licenses/npm/`. Bundled JavaScript retains upstream legal comments.

The browser phonemizer is the GPL-3.0-or-later eSpeak NG build distributed by [offline-speech](https://github.com/dimacurrentai/offline-speech/tree/9fdb771e5bfeec518ed7e9f7afc0f27985623a97), with its Apache-2.0 wrapper attribution to Xenova preserved. The binary module is unchanged; its digest and source identity are in `third_party/espeak-ng/PROVENANCE.json`. The corresponding-source release archive includes that exact integration repository, its build script/patches/wrapper, and eSpeak NG source `890071ff30231883d761e4d59edabd6d69faf159`. Its documented build uses Emscripten 7.0.0. This avoids treating the wrapper's license as the license for the embedded speech engine. These files and the PageSpindle integration source can be downloaded at no charge alongside the app.
