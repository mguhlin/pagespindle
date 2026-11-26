# Four protections and Go platform contract

PageSpindle's primary task is authorized text → reviewable preparation → local narration → saved audiobook. Adults and educators are the intended initial audience. No student record service, accounts, remote inference, application analytics, or automatic publication is part of the application.

| Gate | Implementation contract | Required evidence |
| --- | --- | --- |
| [Design and accessibility](https://mguhlin.github.io/vcc/WEB_DESIGN.md) | Reading workspace; paste/open, source comparison, optional review, voice preview, export. Keyboard access, labeled controls, reflow, progress and recovery. | Browser tasks, contrast/axe checks, narrow layout, error flow; assistive-technology gaps named separately. |
| [Education privacy](https://mguhlin.github.io/vcc/EDU_COMPLIANCE.md) | Input files are read through explicit browser file selection. Text stays in the local app and its bundled engines. Work stays in memory/private temporary output until explicit download. | Network/storage inventory, clear/reset, export disclosure, and no student data in fixtures. Local use is not blanket school authorization. |
| [Security](https://mguhlin.github.io/vcc/securevcc/securevcc.md) | Loopback binding; session key, exact host/origin checks, bounded inputs, safe DOM rendering, no generic read/write/execute bridge. Untrusted documents/model output cannot grant authority. | Allowed and denied requests; malformed/oversized imports; stale jobs; safe output names; no outside-tree writes. |
| [Rescue and recoverable delivery](https://mguhlin.github.io/vcc/rescue/VIBE_CODE_RESCUE.md) | Preserve source; proposals can be reviewed optionally; cancellation, truthful errors, saved project backups and bounded progress. Pinned engines, model identities and package inventories. | Save/reopen, failure/retry, cancelled jobs, reproducible build, actual extracted-package operation. |

[Go App Builder](https://mguhlin.github.io/vcc/skills/go-app-builder/SKILL.md) governs separation between pure Go text/session logic, browser view, native subprocesses, platform adapters, and release evidence. The desktop app is a Go executable using Wails v3.0.0 to host the embedded interface in its own native WebView window (WebView2/WKWebView/WebKitGTK6). It does not launch an external browser for the workspace. Native engines are separate bundled components, not Go implementations of neural inference.

Initial package targets: Windows x64; macOS Intel and Apple Silicon; Linux x64. The OS installers include application engines and the media converter, install required system runtimes, and automatically download verified models on first launch. Setup requires internet; processing after setup is local. Clean physical-OS installation and offline source builds remain separate claims. Extra architectures are not promised through cross-compilation alone.

The design uses a book-preparation workspace rather than a marketing dashboard: navy reading text, white editing surfaces, pale silver dividers, teal actions, coral warnings, gold page markers, system sans-serif controls and Georgia reading text. Source and prepared passages remain adjacent on desktop and stack on narrow screens. Generated paper-sculpture artwork communicates the book-to-sound workflow. Motion is unnecessary. Local inputs and exports are sensitive even without a server account.

Release evidence belongs in `VALIDATION.md`. A passing test is scoped evidence; no certification, complete security, legal compliance, or untested-platform claim follows from it. Preserve substantive source/license dates and stable URLs; omit editorial maintenance timestamps.

## Browser platform contract

The no-install browser edition compiles the same Go book/source/review core to WebAssembly. Workers run WebLLM/Qwen on WebGPU and the LJ Speech model on ONNX Runtime Web CPU/WASM. The browser needs HTTPS/localhost, workers, WebAssembly, Cache Storage/IndexedDB, and a usable WebGPU adapter for AI. Float16 options require that GPU feature; the default uses float32. Browser AI memory guidance is approximately 3 GB available GPU memory plus system headroom, rather than a guarantee for all devices. Explicit setup downloads models; inference does not send manuscript text to a server.

| Task | Desktop | No-install browser | Recovery or limit |
| --- | --- | --- | --- |
| Paste; TXT/MD/EPUB/searchable PDF import | SUPPORTED | SUPPORTED, same Go logic | 16 MiB input, 2 MiB extracted text; reject invalid input without replacing current work. |
| Local AI cleanup proposals | SUPPORTED, CPU Qwen3-1.7B | SUPPORTED on compatible WebGPU hardware; Qwen3-1.7B | Original retained; complete structured proposals only; passage comparison is optional. |
| Manual preparation/chatbot prompt | SUPPORTED | SUPPORTED | Hosted chatbot is a separate user-directed transfer. |
| Notes/review/project backup | SUPPORTED | SUPPORTED, same format/core | Refresh requires deliberate backup/reopen. |
| English WAV chapters/ZIP/playlist | SUPPORTED | SUPPORTED | Native 300,000 text bytes/768 MiB audio; browser 100,000/96 MiB. |
| MP3/M4B | SUPPORTED with bundled FFmpeg | UNSUPPORTED | Browser users can download WAV and convert locally outside the app. |
| Offline processing | SUPPORTED after automatic first-launch setup | SUPPORTED after shell and model setup | Browser caches can be evicted; test offline readiness before travel. |
| OCR, DRM removal, visual interpretation, listening PDF | UNSUPPORTED | UNSUPPORTED | Prepare accessible text separately and inspect extraction. |
| Physical-device install/signing/human AT | UNKNOWN | UNKNOWN across the full browser/device matrix | Hosted-runner and automated UI evidence is not device/AT qualification. |

Browser clear/cancel terminates inference workers, rejects pending work and invalidates late results; manuscripts and audio remain in memory until explicit download. Model removal is a separate action limited to selected PageSpindle model records and its voice cache. Text-model caches can share assets with other pages on the same GitHub Pages origin; the UI discloses that removal can require those pages to re-download matching assets. The service worker caches a coherent versioned shell, retains old shell caches for existing clients, and applies updates through an explicit save-and-reload action. It does not clear unrelated origin storage. No cache deletion promises forensic erasure.

Windows WebView2 is a separately licensed OS runtime with Microsoft-required diagnostics and OS-controlled optional diagnostic collection. PageSpindle does not add application analytics or content uploads. See [Microsoft data/privacy documentation](https://learn.microsoft.com/en-us/microsoft-edge/webview2/concepts/data-privacy). System runtime diagnostics are outside the manuscript pipeline; do not describe the entire Windows stack as having zero telemetry.

Markdown is parsed only to derive narration text, before preview truncation and speech chunking. The parser runs locally and does not render HTML, execute code or fetch link/image destinations. Original text, prepared passages and project backups retain their formatting. Literal symbols and code content are preserved; recognized formatting delimiters are omitted from speech. Regression fixtures verify both derived speech and unchanged project/source data.
