# PageSpindle

**Spin text into listening.** Prepare authorized text, compare proposed changes, and make an audiobook with local AI and narration. No account or paid narration service is required.

[Open PageSpindle](https://mguhlin.github.io/pagespindle/) · [Browser edition](https://mguhlin.github.io/pagespindle/browser/) · [Downloads](https://github.com/mguhlin/pagespindle/releases/tag/v0.1.0) · [VCC Audiobook Maker](https://mguhlin.github.io/vcc/guides/skills.html#audiobook-maker)

## Start with the browser edition

1. Open the browser edition over HTTPS in a browser with WebGPU support, such as a compatible recent Chrome or Edge. AI needs a usable GPU adapter and approximately 3 GB available GPU memory for the default Qwen3-1.7B model; more headroom improves reliability. Float16 alternatives are offered only for GPUs supporting that format. Phones and low-memory devices may be unsuitable.
2. Choose **Install text model** for optional local AI, and **Install English voice** for narration. Setup downloads hundreds of MB to over 1 GB for a text model plus about 40 MB for the voice; allow at least 4 GB free browser storage, more for alternate models and audio. Download hosts receive model/asset requests, not manuscript text. Models stay in this browser's site storage, which the browser may evict.
3. Paste text or open TXT, Markdown, EPUB, or a searchable PDF. Make a manual review copy or ask the local Qwen model to propose conservative cleanup. Compare every passage with its preserved source. Mark note destinations and deliberate omissions, with reasons. Narration requires recorded review.
4. Preview the English voice, then create a ZIP of WAV chapters, the listening text, and an M3U playlist. Save a project backup to retain the original, edits, and review decisions. Audio export is bounded to 100,000 UTF-8 text bytes and 96 MiB WAV per job; use parts for long books.
5. After setup, use **Reload offline app** and test a small passage with networking off before relying on offline access. The versioned shell, Go core, models and voice run locally. First visit and model setup require internet. Clearing local work removes in-memory text/audio; removing downloaded models is a separate action. Saved files and browser/OS backups remain.

The optional chatbot step copies a passage with faithful preparation instructions. A local chatbot can keep it offline; a hosted chatbot receives whatever you paste there. Review the result and paste it back into the prepared passage. The built-in local model removes the need for this handoff.

## Install a desktop edition

Download the matching **platform ZIP** and the shared **common model pack** from the release. Extract both so their `PageSpindle/` contents merge into one new folder, without overwriting an existing project. The executable, `runtime/` and `models/` must sit together:

```text
PageSpindle/
  pagespindle.exe                 (Windows; pagespindle on macOS/Linux)
  runtime/windows-amd64/          (or your matching target)
  models/Qwen3-1.7B-Q8_0.gguf
  models/vits-piper-en_US-ljspeech-high-int8/
```

The common pack is about 1.9 GB; allow at least 5 GB free disk for downloads, extraction, and outputs. Start with 8 GB system RAM or more and close memory-heavy applications; actual speed and memory requirements depend on the machine and document. Native AI can run on CPU and does not require browser WebGPU.

| Download | Launcher | Qualification boundary |
| --- | --- | --- |
| Windows x64 | `Start PageSpindle.cmd` | Built and real engines exercised on Windows Server 2022 runner; Windows desktop installation/SmartScreen qualification remains separate. |
| macOS Intel | `Start PageSpindle.command` | Built and real engines exercised on macOS 15 Intel runner. |
| macOS Apple Silicon | `Start PageSpindle.command` | Built and real engines exercised on macOS 15 ARM runner. |
| Linux x64 | `Start PageSpindle.sh` | Built and real engines exercised on Ubuntu 24.04 runner and local Linux; requires glibc/libstdc++ compatible with the bundled distributions. |

These are unsigned evaluation packages. Publisher signing, macOS notarization, physical-device installation and human screen-reader qualification are not claimed. Follow your organization's software rules; do not disable OS protections globally. Older OS versions and additional architectures have not been qualified. See [release evidence](VALIDATION.md).

Windows AI also requires Microsoft’s [Visual C++ x64 Redistributable](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist), which provides MSVCP140/VCRUNTIME140 for the pinned llama.cpp build. Install it from Microsoft or your organization’s approved offline installer before offline use if it is not already present. It is a separate system dependency, not bundled or installed automatically. The Windows runner already has it; a clean Windows installation remains unqualified.

Keep the launcher open: it serves an embedded workspace on a private ephemeral `127.0.0.1` port and opens your normal browser. The URL contains a private session key; do not share it. Press Ctrl+C in the launcher to stop engines and clear temporary audio. After extraction, all required runtime/model files are local, including on first launch without internet. Refresh loses unsaved work and needs the private launcher URL again. Save a project backup before closing or refreshing.

The desktop edition uses Qwen3-1.7B Q8 and the English LJ Speech voice. WAV chapter ZIPs work with the bundled engine. A separately installed local **FFmpeg** enables MP3 chapter ZIPs and M4B with chapter metadata. Native jobs allow 300,000 UTF-8 text bytes and 768 MiB WAV; long books should be processed in parts. No model, engine or content upload occurs during processing.

## Fidelity and privacy

This implements a bounded version of VCC's skill: preserve source, propose conservative mechanical cleanup, review passages, relocate notes by original chapter headings, and narrate. Models can omit or alter words; structured output and matching endpoints cannot certify fidelity. Review every passage. OCR, interpreting unseen figures, DRM removal, listening-ready PDF output, and automatic book-wide fidelity verification are not implemented. PDF/EPUB extraction can have imperfect reading order.

Imports are explicit browser file selections, never arbitrary renderer paths. Originals are not overwritten. Work stays in memory and, in the desktop edition, private temporary audio until deliberate download. Normal clear/cancel invalidates late results; abrupt power loss can leave temporary files. Clear does not erase saved downloads, clipboard contents, OS/browser backups, or forensic traces. No accounts, telemetry, remote fonts, hosted inference, or automatic publication are included. Local processing alone does not grant institutional approval for student records.

## Reproduce and recover

Use Go **1.27.4**, Python 3.12+ with `tarfile` data filtering, and a native C compiler for your matching target. Windows builds use Visual Studio C tools. Native dependencies are hash-pinned in `dependencies.json`. Go dependencies are vendored.

```sh
python3 scripts/prepare_runtime.py --target linux-amd64 --common
python3 scripts/build.py --target linux-amd64
python3 scripts/verify_package.py linux-amd64
PAGESPINDLE_RUNTIME="$PWD" go test -v -run TestRealOfflineEngines ./internal/engine
```

Substitute the matching target and run on that OS/CPU. `go test ./...`, `go vet ./...` and `go test -race ./...` cover core and boundary regressions where the toolchain supports race checks. `scripts/package_common.py` creates the shared model and corresponding-source archives and checks retained source identities. `SOURCE_ARCHIVES.json` identifies the exact upstream source packages. Release ZIP inventories and `SHA256SUMS.txt` identify downloadable bytes; checksums are not publisher signatures.

For the browser build, use Node.js 20+, `npm ci --ignore-scripts`, the voice from the prepared common pack, then `npm run build:web`. Go compiles `cmd/browsercore` to `js/wasm`; the build copies the matching Go runtime. Immutable browser model coordinates and SRI are in `browser/browser-models.json`. `scripts/prepare_browser_models.py` deliberately re-derives that manifest from the pinned upstream locations; do not silently change model revisions. Serve `docs/` over HTTPS or localhost. `scripts/browser_check.cjs` exercises real voice, backup roundtrip, clear, reflow, automated accessibility and cached offline use; `scripts/browser_model_check.cjs` exercises actual browser AI and cached offline preparation on WebGPU hardware. Native UI mode uses `PAGESPINDLE_TEST_MODE=native`. Browser caches and test fixtures stay in ignored `.cache/`.

## Attribution and licenses

The workflow adapts **Lance Eaton's Book to Audio Edition**, from his [MIT-licensed repository](https://github.com/leaton79/book-to-audiobook) and [AI Audiobooks: An Update](https://aiedusimplified.substack.com/p/ai-audiobooks-an-update). Miguel Guhlin adapted it for VCC/PageSpindle. No endorsement is implied. The VCC skill retains Eaton's full MIT notice; the app implements conservative proposal and review behavior and includes no book content.

Application code is **GPL-3.0-or-later**, compatible with the bundled GPL speech components. Qwen weights are Apache-2.0; other engines, browser libraries, voices and datasets retain their own terms. Full notices and corresponding source accompany the releases; see [THIRD_PARTY.md](THIRD_PARTY.md). Software/model licenses do not grant permission to publish somebody else's book.

Original documentation and splash/icon artwork are CC BY 4.0 by Miguel Guhlin. [Artwork originals](artwork/) and [complete image-generation prompts](ARTWORK_PROMPTS.json) are retained. The artwork was made with built-in OpenAI image generation.
