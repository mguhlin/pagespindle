# Release evidence

This is an unsigned evaluation release. Passing builds and synthetic checks do not establish full accessibility, security, legal compliance, institutional approval, or physical-device compatibility. Native signing/notarization and human screen-reader checks are not performed in this workspace.

## Four independent gates

| Gate | Evidence and boundary |
| --- | --- |
| Design/accessibility | Keyboard-oriented reading workspace, source/proposal comparison, optional review, labeled controls, responsive layout. Browser checks exercise the synthetic workflow and automated accessibility scanning; human assistive-technology qualification remains separate. |
| Education/privacy | No accounts, application analytics, remote fonts, hosted inference, or automatic external content upload. Assets are downloaded during setup; desktop inference calls only loopback; browser inference runs in workers. User-selected imports and explicit backups/audio downloads can contain sensitive text. School authorization is not inferred. |
| Security | Exact local host/origin and session key checks; bounded import/output; immutable source digest and passage validation; safe DOM text rendering; allowlisted download inventory; serial cancellable jobs; protected originals. Unit and regression checks accompany source. |
| Recoverable delivery | Pinned dependency hashes, source and component notices, project save/reopen, private temporary outputs, clear/cancel invariants, native runtime and package inventory verification. Published package checksums identify downloadable bytes. |

The local Linux integration test runs the actual Qwen model and speech voice on invented text and verifies complete structured output plus nonzero mono PCM. It is a small pilot, not a full-book fidelity or speech-quality evaluation. Optional side-by-side comparison can help assess model proposals. The native runner outcomes are recorded in GitHub Actions. Browser evidence is produced by the published reproducible Playwright scripts on local Chrome; verify the run/artifact identity associated with the release.

Windows x64, macOS Intel, macOS Apple Silicon, and Linux x64 each have a native build job. A package is published only after its build/test/package checks succeed. A hosted-runner CLI test is not a real native picker, downloaded-quarantine/Gatekeeper, ordinary-user installation, audiobook-player resume, or physical-device test. Untested combinations remain unqualified. Use source builds or institution-approved software channels where unsigned preview packages are unsuitable; do not disable operating-system protections globally.

## Limits

- Searchable PDFs and EPUBs can extract imperfect reading order. OCR, image interpretation, and DRM removal are not implemented. Preserve the source and inspect extracted text.
- The browser edition uses local WebGPU Qwen preparation and CPU/WASM speech after explicit downloads; caches can be evicted. Its audio jobs are limited to 100,000 UTF-8 text bytes and 96 MiB WAV. Browser output is WAV chapter ZIPs; MP3/M4B require the desktop edition or another local conversion tool.
- The bundled voice is English. Narration is limited to 300,000 UTF-8 bytes and 768 MiB of WAV output per job; save long books in parts. WAV chapter ZIPs include text and a playlist. MP3/M4B use the bundled minimal FFmpeg.
- Models can omit or alter text; structured output, matching endpoints, and word counts do not certify fidelity. Side-by-side review and optional comparison can help identify errors.
- Refresh/close loses unsaved browser work. Save the project backup first and reopen it through the file selector. The desktop app uses its own native window; save a project before closing.
- Clear local work or closing the native desktop window cancels jobs and removes app-owned temporary audio. Abrupt power loss can leave temporary files in the OS temporary directory. No forensic erasure or removal of independent saved/backed-up copies is promised.
- Windows setup automatically obtains the required Microsoft Visual C++ x64 runtime. Hosted runners already contain many system runtimes; clean physical OS installation remains unqualified.
- The source remains public, with exact module and engine identities. No independent security audit, complete software bill of materials, publisher signing, or notarization is claimed.

Release decision: **CONDITIONAL for evaluation with authorized, nonsensitive text**, limited to the checks recorded for each package. Additional qualification is needed before sensitive school use or a trusted signed distribution channel.

## Reproducible checks

| Check | Result and scope |
| --- | --- |
| Go unit/regression tests, vet, race | PASS on local Linux. Source identity/Unicode, bounded input, EPUB spine order and traversal, project roundtrip, note grouping, optional passage checks, host/origin/session denial, cancel/clear and stale-worker checks. |
| Actual native model and voice | PASS on Windows x64, macOS Intel/ARM and Linux x64 hosted runners. Real Qwen structured proposal, real 22,050 Hz PCM through a Unicode filename, and owner-pipe EOF shutdown. See the Native package checks run tied to the release source. |
| Native browser workspace | PASS on local Linux Chrome: prompt copy, exact original, backup/reopen, real speech preview, WAV chapter ZIP, listening-text export, clear, 320 CSS-pixel reflow and axe checks. Real M4B export decoded by FFprobe with two chapters and nonzero duration. |
| Browser local AI | PASS on Linux Chrome 158 with NVIDIA RTX 2060 through WebGPU, Qwen3-1.7B q4f32. The clean pilot retained its complete prose and 9:00/5:00 numbers; proposals remained unreviewed. Cached offline reload and another real preparation passed. |
| Browser narration and recovery | PASS on local Chrome: actual ONNX voice/phonemizer, preview decode, WAV ZIP, text/project exports, clear, backup reimport, offline shell reload and real offline narration, axe checks and 320 CSS-pixel reflow. |
| Observed browser transfers | PASS for synthetic tests: model/asset GET requests, no manuscript POST or document-body upload. This is scoped observation, not an independent network audit. |
| Dependency checks | Go vulnerability check reported no vulnerabilities; npm audit reported zero vulnerabilities with the pinned lock. These results are not a permanent security guarantee. |
| Package identity | Platform ZIP inventory compares every staged component to archive bytes; common-model packaging verifies its complete inventory and hashes. Release checksums and artifact source revisions identify published downloads. |
| Human AT, physical installations, signing | NOT_RUN. Keyboard/axe/reflow checks do not establish screen-reader compatibility or WCAG conformance. |

Browser first model startup includes substantial downloads and GPU compilation. A software-only WebGPU renderer exceeded the preparation limit in the pilot; use supported hardware acceleration or the desktop CPU route. The float16 alternative additionally requires `shader-f16` and has not been qualified on the local adapter, which did not expose that feature. The default float32 path was exercised. A small invented pilot is not a whole-book completeness or pronunciation evaluation.

## Native installers — v0.2.0

Wails v3.0.0 replaces the external browser launcher. Windows, macOS Intel/ARM and Linux installers include local AI/speech engines and a built-from-source AAC/MP3 converter. First launch downloads and verifies the pinned released model pack, with cancellation/retry and preservation of existing model folders. Review does not gate narration.

PASS locally: Go unit, vet and race checks; model archive checksum, extraction, per-file verification and tamper rejection; path/corruption/cancellation regressions; a real Linux WebKitGTK native-window pilot with direct unreviewed narration, actual Qwen preparation, decoded preview, project save adapter, M4B export and clear. The fixed synthetic pilot supplies a save-dialog destination; it does not claim human OS picker gestures. Docker virtual-display checks use a sandbox exception only inside the isolated test container, never shipped in the application.

**PASS on all four targets:** [Native package checks](https://github.com/mguhlin/pagespindle/actions/runs/45523461210), application source `00a05f657ac1119088c3b4a9b584182da65231d2`. Windows Server 2022 x64, macOS 15 Intel, macOS 15 Apple Silicon and Ubuntu 24.04 x64 built the actual installers, installed them twice, compared installed file identity, exercised their own native window with actual Qwen preparation plus WAV/MP3/M4B exports, and uninstalled them while preserving an external recovery sentinel. Mac receipts identify the bundle after ad-hoc signing. `RELEASE_EVIDENCE.json` records the scope; release checksums identify the installer bytes.

**PASS locally:** automatic first-launch model download, verified extraction and transition into the native workbench; a fresh Ubuntu 24.04 container and fresh Debian 13 container installed the declared system dependencies and completed the actual native workflow; a Linux native pilot completed with external networking disabled using the automatically installed models. The latest Go vulnerability scan included the native desktop packages and reported no known vulnerabilities. The live browser direct-creation, real voice, backup/recovery, axe/reflow and offline-narration suite passed. Physical clean OS, downloaded-quarantine interaction, human assistive technology, cross-version package upgrades and publisher signing remain NOT_RUN. Model files and shared OS runtimes are intentionally retained on uninstall. No full-book accuracy guarantee follows from a short synthetic pilot. Windows WebView2 has its own required diagnostics, OS-controlled optional diagnostics and shared-runtime update policy; PageSpindle does not claim to suppress all OS/runtime traffic.

## GPU and workbench update — v0.3.0

Local PASS: actual Qwen preparation with Vulkan selecting NVIDIA GeForce RTX 2060, structured-source checks, real CPU narration, and owner-pipe cleanup. Tests cover discrete-device preference, exclusion of software renderers, Metal device IDs, forced CPU offload flags, GPU startup failure falling back to CPU, and failed GPU inference retrying once on CPU without changing the source. Race checks exercise restarts and stale backend reports.

The interface puts audio progress and completion beside Create audiobook, reports passage work within chapters, and derives approximate remaining time from measured progress. CPU narration and GPU AI preparation are labeled separately; the current voice does not use the GPU. No fixed GPU speedup or full-book completion time is promised. Optional comparison, source preservation, project recovery and limits remain in place. Cross-platform installer and actual Metal/AMD/Intel hardware outcomes must be read separately from the local NVIDIA result.

An experimental ONNX Runtime WebGPU speech run on the local RTX 2060 took 10.7–12.4 seconds for an invented passage, compared with 5.9–6.1 seconds using the existing CPU/WASM path. This scoped comparison supports keeping the current voice on CPU; it is not a universal GPU benchmark. Experimental GPU speech assets are not shipped.

**PASS on all four targets:** [v0.3.0 Native package checks](https://github.com/mguhlin/pagespindle/actions/runs/45891375164), application source `5415c632ffabb0efac47a2e72c7d7396f79dcd87`. Windows x64, macOS Intel/Apple Silicon and Ubuntu x64 built and verified the packaged runtimes, ran real local Qwen and speech engines, installed twice, exercised the installed native window with WAV/MP3/M4B, checked installed component identity, and uninstalled while preserving the external recovery sentinel. Mac installation is fixed to Applications; the Intel Metal build rejects dependencies on developer-machine library paths. Hosted Mac checks used the available runner hardware and do not qualify physical Metal GPUs.

**PASS locally for this revision:** Go unit, vet and race checks; the native Go vulnerability scan; the actual NVIDIA Vulkan model-to-voice pilot; actual browser WebGPU preparation and cached offline preparation; and live browser direct narration, measured partial-chapter progress beside Create audiobook, preview, project recovery, offline narration, automated accessibility and 320px reflow. The published release includes full installer hashes, Mac post-sign receipts, the unchanged model-pack pointer and machine-readable `RELEASE_EVIDENCE.json`. The earlier v0.2.0 clean-container, first-launch-download and network-isolation pilots remain separately scoped evidence; they were not repeated for this revision. Physical OS installs, human AT, publisher signing/notarization and a physical GPU-vendor matrix remain NOT_RUN.

## Markdown narration correction

The shared Go core derives speech text with the pinned, vendored Goldmark parser before splitting chapters, previews or voice chunks. Regression fixtures cover all heading levels and setext headings, nested emphasis, quotes, lists, references, image alt text, tables, checkboxes, literal symbols, code fences, long passage boundaries, formatting-only rejection and byte-preserved project recovery. The native preview now truncates the derived speech rather than cutting Markdown delimiters. Browser checks inspect actual voice-worker inputs and generate real audio. Candidate installer evidence is recorded with the associated release; previous release checks do not qualify rebuilt packages.
