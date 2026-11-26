# PageSpindle

**Spin text into listening.** Prepare authorized text, compare proposed changes, and make an audiobook with local AI and narration. No account or paid narration service is required.

[Open PageSpindle](https://mguhlin.github.io/pagespindle/) · [Browser edition](https://mguhlin.github.io/pagespindle/browser/?v=0.2.0) · [Downloads](https://github.com/mguhlin/pagespindle/releases/tag/v0.3.0) · [VCC Audiobook Maker](https://mguhlin.github.io/vcc/guides/skills.html#audiobook-maker)

## Start with the browser edition

1. Open the browser edition over HTTPS in a browser with WebGPU support, such as a compatible recent Chrome or Edge. AI needs a usable GPU adapter and approximately 3 GB available GPU memory for the default Qwen3-1.7B model; more headroom improves reliability. Float16 alternatives are offered only for GPUs supporting that format. Phones and low-memory devices may be unsuitable.
2. Choose **Install text model** under optional settings for local AI. The English voice installs automatically when first creating audio. Setup downloads hundreds of MB to over 1 GB for a text model plus about 40 MB for the voice; allow at least 4 GB free browser storage, more for alternate models and audio. Download hosts receive model/asset requests, not manuscript text. Models stay in this browser's site storage, which the browser may evict.
3. Paste text or open TXT, Markdown, EPUB, or a searchable PDF. Make a manual review copy or ask the local Qwen model to propose conservative cleanup. You can compare passages with their preserved source. Mark note destinations and deliberate omissions, with reasons. Passage review is optional; Create audiobook starts from the loaded text immediately.
4. Preview the English voice, then create a ZIP of WAV chapters, the listening text, and an M3U playlist. Save a project backup to retain the original, edits, and review decisions. Audio export is bounded to 100,000 UTF-8 text bytes and 96 MiB WAV per job; use parts for long books.
5. After setup, use **Reload offline app** and test a small passage with networking off before relying on offline access. The versioned shell, Go core, models and voice run locally. First visit and model setup require internet. Clearing local work removes in-memory text/audio; removing downloaded models is a separate action. Saved files and browser/OS backups remain.

The optional chatbot step copies a passage with faithful preparation instructions. A local chatbot can keep it offline; a hosted chatbot receives whatever you paste there. Review the result and paste it back into the prepared passage. The built-in local model removes the need for this handoff.

## Install a desktop edition

Choose the installer for your computer from [Downloads](https://github.com/mguhlin/pagespindle/releases/tag/v0.3.0). PageSpindle opens in **its own desktop window**. No external browser or separately installed AI server is required.

| Computer | Installer and runtime setup |
| --- | --- |
| Windows x64 | Run `pagespindle-windows-amd64-setup.exe`. Setup includes the full Microsoft WebView2 installer and downloads/verifies Microsoft's C++ runtime automatically. Administrator permission and internet are required during setup. |
| macOS 15+ Intel | Open `pagespindle-darwin-amd64.pkg`; launch PageSpindle from Applications. Uses macOS's built-in WebView. |
| macOS 15+ Apple Silicon | Open `pagespindle-darwin-arm64.pkg`; launch PageSpindle from Applications. Uses macOS's built-in WebView. |
| Ubuntu 24.04+ / Debian 13, Linux x64 | Install `pagespindle-linux-amd64.deb` with a dependency-aware package manager, or `sudo apt install ./pagespindle-linux-amd64.deb`. The package manager installs GTK4, WebKitGTK6 and audio dependencies automatically. Internet and administrator permission are required for setup. Launch PageSpindle from the application menu. Other distributions are not qualified. |

The first launch automatically downloads a pinned, verified **1.9 GB model pack** inside PageSpindle's setup window. Progress, cancellation and retry are included. Allow at least 8 GB free disk during setup and start with 8 GB RAM or more. Models are stored in the user's application-data folder, separately from the installed program; upgrades and uninstall preserve them. After successful setup, AI preparation and narration work offline on CPU. Test a short passage with networking off before relying on offline use.

Paste text or open a file, then click **Create audiobook**. Passage comparison and AI preparation are optional. The native Save dialog writes a new file and preserves existing files; choose another name if it already exists. The app bundles the voice engine, local AI engine and a minimal **FFmpeg** converter for WAV chapter ZIPs, MP3 chapter ZIPs, and M4B with chapters. English narration is included. Native jobs allow 300,000 UTF-8 text bytes and 768 MiB WAV; save longer books in parts.

Save a project backup to keep source text and edits. Closing the window stops engines and clears app-owned temporary audio; unsaved work prompts before closing. The internal UI uses a private authenticated loopback connection inside the native WebView. Explicit documentation links open in the default browser. Windows WebView2 is a shared Microsoft runtime with its own update policy; macOS/Linux system components follow their OS update policies.

These are unsigned evaluation installers. Publisher signing, macOS notarization and physical-device or human screen-reader qualification are not claimed. Organizations may restrict unsigned software. See [release evidence](VALIDATION.md). Legacy v0.1.0 ZIPs used an external browser and are superseded by these installers.

## Fidelity and privacy

This implements a bounded version of VCC's skill: preserve source, propose conservative mechanical cleanup, review passages, relocate notes by original chapter headings, and narrate. Models can omit or alter words; structured output and matching endpoints cannot certify fidelity. Passage comparison is optional. OCR, interpreting unseen figures, DRM removal, listening-ready PDF output, and automatic book-wide fidelity verification are not implemented. PDF/EPUB extraction can have imperfect reading order.

Imports are explicit browser file selections, never arbitrary renderer paths. Originals are not overwritten. Work stays in memory and, in the desktop edition, private temporary audio until deliberate download. Normal clear/cancel invalidates late results; abrupt power loss can leave temporary files. Clear does not erase saved downloads, clipboard contents, OS/browser backups, or forensic traces. PageSpindle adds no accounts, application analytics, remote fonts, hosted inference, or automatic publication. Windows WebView2 has Microsoft-required diagnostics and OS-controlled optional diagnostics under [Microsoft's privacy policy](https://learn.microsoft.com/en-us/microsoft-edge/webview2/concepts/data-privacy); its shared runtime can update separately. This is not a promise of zero operating-system traffic. Local processing alone does not grant institutional approval for student records.

## Reproduce and recover

Use Go **1.27.4**, Python 3.12+ with `tarfile` data filtering, and a native C compiler for your matching target. Windows builds use Visual Studio C tools for the speech adapter and MSYS2 MinGW GCC/make for native media builds. Linux builds need GTK4/WebKitGTK6 development packages. Native dependencies are hash-pinned in `dependencies.json`. Go dependencies are vendored.

```sh
python3 scripts/prepare_runtime.py --target linux-amd64 --common
python3 scripts/build_media.py --target linux-amd64
python3 scripts/build.py --target linux-amd64
python3 scripts/verify_package.py linux-amd64
python3 scripts/build_installers.py --target linux-amd64
PAGESPINDLE_RUNTIME="$PWD" go test -v -run TestRealOfflineEngines ./internal/engine
```

Substitute the matching target and run on that OS/CPU. `go test ./...`, `go vet ./...` and `go test -race ./...` cover core and boundary regressions where the toolchain supports race checks. `scripts/package_common.py` creates the shared model and corresponding-source archives and checks retained source identities. `SOURCE_ARCHIVES.json` identifies the exact upstream source packages. Release ZIP inventories and `SHA256SUMS.txt` identify downloadable bytes; checksums are not publisher signatures.

For the browser build, use Node.js 20+, `npm ci --ignore-scripts`, the voice from the prepared common pack, then `npm run build:web`. Go compiles `cmd/browsercore` to `js/wasm`; the build copies the matching Go runtime. Immutable browser model coordinates and SRI are in `browser/browser-models.json`. `scripts/prepare_browser_models.py` deliberately re-derives that manifest from the pinned upstream locations; do not silently change model revisions. Serve `docs/` over HTTPS or localhost. `scripts/browser_check.cjs` exercises real voice, backup roundtrip, clear, reflow, automated accessibility and cached offline use; `scripts/browser_model_check.cjs` exercises actual browser AI and cached offline preparation on WebGPU hardware. Native UI mode uses `PAGESPINDLE_TEST_MODE=native`. Browser caches and test fixtures stay in ignored `.cache/`.

## Attribution and licenses

The workflow adapts **Lance Eaton's Book to Audio Edition**, from his [MIT-licensed repository](https://github.com/leaton79/book-to-audiobook) and [AI Audiobooks: An Update](https://aiedusimplified.substack.com/p/ai-audiobooks-an-update). Miguel Guhlin adapted it for VCC/PageSpindle. No endorsement is implied. The VCC skill retains Eaton's full MIT notice; the app implements conservative proposal and review behavior and includes no book content.

Application code is **GPL-3.0-or-later**, compatible with the bundled GPL speech components. Qwen weights are Apache-2.0; other engines, browser libraries, voices and datasets retain their own terms. Full notices and corresponding source accompany the releases; see [THIRD_PARTY.md](THIRD_PARTY.md). Software/model licenses do not grant permission to publish somebody else's book.

Original documentation and splash/icon artwork are CC BY 4.0 by Miguel Guhlin. [Artwork originals](artwork/) and [complete image-generation prompts](ARTWORK_PROMPTS.json) are retained. The artwork was made with built-in OpenAI image generation.

## GPU selection and progress

The desktop text model automatically selects a compatible GPU, preferring a discrete card when one is detected: Vulkan on Windows/Linux and Metal on macOS. GPU backends are packaged with the app; compatible graphics drivers are supplied by the OS/GPU vendor. The app probes before loading the model and falls back to its bundled CPU engine if GPU detection, startup or inference fails. Use `--cpu` when launching the desktop executable to explicitly choose CPU preparation. No CUDA SDK or separate model server is required. The existing models are reused.

The browser's optional text preparation uses WebGPU. The current English narration voice uses CPU on desktop and WebAssembly CPU in the browser. GPU preparation does not imply GPU speech synthesis. The creation panel names the processor used by the active operation.

Progress advances as passages finish inside each chapter. Elapsed time and approximate time remaining sit beside Create audiobook and Cancel. Estimates learn from completed text on the actual device and refine as it runs; first-use downloads, model loading and final encoding are separate phases. Approximate listening length is shown separately from creation time. Performance samples remain in memory and contain no stored manuscript or device profile.

The interface groups audio controls and output in one listening panel, with optional comparison in a disclosure. Its workflow was informed by [NaturalReader's audio conversion controls](https://help.naturalreaders.com/en/articles/11543218-working-with-text-and-audio-personal-version) and [ebook2audiobook's chapter-based interface](https://github.com/DrewThomasson/ebook2audiobook). PageSpindle retains its own artwork, local processing, direct creation and free exports.

## Markdown narration

Preview and audiobook creation interpret Markdown locally before speech, without an AI pass. Heading hashes, emphasis delimiters, bullet/quote markers, code fences, horizontal rules and table separators are silent. Link labels and image alternative text are read without their hidden destinations; no linked file is fetched. Ordered-list numbers and checkbox completion states remain audible. Struck-through words are retained rather than silently omitted.

Literal hashtags (`#learning`), C#, arithmetic such as `2 * 3`, citations and code content remain. Escape a punctuation mark with a backslash or put it in a code span when it should be spoken literally. Unknown/raw HTML remains literal text and is never executed. The original source, prepared passages, saved listening-text master and project backups retain Markdown; only the derived narration changes. Existing projects use this behavior when generating new audio. Previously saved audio must be generated again.
