# PageSpindle

**Spin text into listening.** A Go application with an embedded, local browser workspace for preparing authorized text, reviewing changes, and making an audiobook on your computer.

[PageSpindle website](https://mguhlin.github.io/pagespindle/) · [Source](https://github.com/mguhlin/pagespindle) · [VCC Audiobook Maker](https://mguhlin.github.io/vcc/guides/skills.html#audiobook-maker)

The app preserves the source and works through import → optional local AI preparation → passage review → voice preview → audiobook export. TXT, Markdown, EPUB, and text-based PDF imports are supported. Scanned PDFs need OCR separately. The optional chatbot bridge prepares a copyable prompt; cloud chatbot use is a separate upload, not offline operation. The included local model is Qwen3-1.7B Q8, and the initial narration voice is English LJ Speech. Model proposals require human review; they are not proof that every word survived.

The Go executable serves only on an ephemeral loopback port and opens the user's browser. It is not a hosted processing service or a native-webview application. The interface, text model, and speech engine run locally after the app runtime and common model pack are installed. A supported browser, matching OS/architecture runtime, and adequate memory/storage are required. FFmpeg is optional for MP3/M4B; the built-in output is a ZIP of WAV chapters, listening text, and a playlist. No account or paid narration service is required.

## Source layout

- `cmd/pagespindle`: launcher and isolated PDF extraction worker.
- `internal/book`: bounded input, source identity, passage review, EPUB reading order, and WAV handling.
- `internal/engine`: local model and speech-process adapters.
- `internal/app`: authenticated loopback API, cancellable jobs, private temporary audio, and embedded interface.
- `native/speech-worker.c`: bounded stdin adapter to sherpa-onnx, keeping narration out of command-line arguments.
- `scripts/prepare_runtime.py`, `dependencies.json`: pinned downloads with published hashes.
- `docs/`: GitHub Pages splash and download site.
- `artwork/`, `ARTWORK_PROMPTS.json`: generated artwork originals and prompts.

Build instructions, package inventories, and platform evidence accompany the release. A build on one machine does not qualify another operating system. See [PROTECTIONS.md](PROTECTIONS.md) for the four independent VCC gates and Go App Builder contract.

## Attribution and licenses

The preparation workflow adapts **Lance Eaton's Book to Audio Edition**, shared in his [MIT-licensed repository](https://github.com/leaton79/book-to-audiobook) and introduced in [AI Audiobooks: An Update](https://aiedusimplified.substack.com/p/ai-audiobooks-an-update). Miguel Guhlin adapted it for VCC and PageSpindle. No endorsement is implied. Eaton's notice remains with the VCC skill; the app implements a conservative proposal-and-review workflow and does not redistribute books.

PageSpindle application code is GPL-3.0-or-later; see `LICENSE`. Qwen model weights are Apache-2.0, llama.cpp is MIT, sherpa-onnx is Apache-2.0, and the speech stack includes MIT and GPL components with retained notices and corresponding source links/archives. Voice model and source-book rights are separate. Original splash/icon artwork and documentation are CC BY 4.0 with attribution to Miguel Guhlin; image generation prompts are retained. No software or model license grants permission to publish somebody else's book. See the component notices in `licenses/`.

No maintenance dates are published. Stable source revisions, model identities, package checksums, substantive citation dates, and required copyright notices serve their own purposes.
