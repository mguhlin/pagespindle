# Release evidence

This is an unsigned evaluation release. Passing builds and synthetic checks do not establish full accessibility, security, legal compliance, institutional approval, or physical-device compatibility. Native signing/notarization and human screen-reader checks are not performed in this workspace.

## Four independent gates

| Gate | Evidence and boundary |
| --- | --- |
| Design/accessibility | Keyboard-oriented reading workspace, source/proposal comparison, optional review, labeled controls, responsive layout. Browser checks exercise the synthetic workflow and automated accessibility scanning; human assistive-technology qualification remains separate. |
| Education/privacy | No accounts, telemetry, remote fonts, hosted inference, or automatic external content upload. Assets are downloaded during setup; desktop inference calls only loopback; browser inference runs in workers. User-selected imports and explicit backups/audio downloads can contain sensitive text. School authorization is not inferred. |
| Security | Exact local host/origin and session key checks; bounded import/output; immutable source digest and passage validation; safe DOM text rendering; allowlisted download inventory; serial cancellable jobs; protected originals. Unit and regression checks accompany source. |
| Recoverable delivery | Pinned dependency hashes, source and component notices, project save/reopen, private temporary outputs, clear/cancel invariants, native runtime and package inventory verification. Published package checksums identify downloadable bytes. |

The local Linux integration test runs the actual Qwen model and speech voice on invented text and verifies complete structured output plus nonzero mono PCM. It is a small pilot, not a full-book fidelity or speech-quality evaluation. Model proposals still need review throughout every section. The native runner outcomes are recorded in GitHub Actions. Browser evidence is produced by the published reproducible Playwright scripts on local Chrome; verify the run/artifact identity associated with the release.

Windows x64, macOS Intel, macOS Apple Silicon, and Linux x64 each have a native build job. A package is published only after its build/test/package checks succeed. A hosted-runner CLI test is not a real native picker, downloaded-quarantine/Gatekeeper, ordinary-user installation, audiobook-player resume, or physical-device test. Untested combinations remain unqualified. Use source builds or institution-approved software channels where unsigned preview packages are unsuitable; do not disable operating-system protections globally.

## Limits

- Searchable PDFs and EPUBs can extract imperfect reading order. OCR, image interpretation, and DRM removal are not implemented. Preserve the source and inspect extracted text.
- The browser edition uses local WebGPU Qwen preparation and CPU/WASM speech after explicit downloads; caches can be evicted. Its audio jobs are limited to 100,000 UTF-8 text bytes and 96 MiB WAV. Browser output is WAV chapter ZIPs; MP3/M4B require the desktop edition or another local conversion tool.
- The bundled voice is English. Narration is limited to 300,000 UTF-8 bytes and 768 MiB of WAV output per job; save long books in parts. WAV chapter ZIPs include text and a playlist. MP3/M4B need a separately installed local FFmpeg.
- Models can omit or alter text; structured output, matching endpoints, and word counts do not certify fidelity. Side-by-side review and optional comparison can help identify errors.
- Refresh/close loses unsaved browser work. Save the project backup first and reopen it through the file selector. A refreshed desktop tab needs the private session URL from the launcher again.
- Clear local work or a graceful Ctrl+C stop of the desktop launcher cancels jobs and removes app-owned temporary audio. Closing only its browser tab does not stop the launcher. Abrupt power loss can leave temporary files in the OS temporary directory. No forensic erasure or removal of independent saved/backed-up copies is promised.
- Windows local AI needs the separately installed Microsoft Visual C++ x64 runtime; it is present on the test runner. Clean OS installation remains unqualified.
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
