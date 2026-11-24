# Release evidence

This is an unsigned evaluation release. Passing builds and synthetic checks do not establish full accessibility, security, legal compliance, institutional approval, or physical-device compatibility. Native signing/notarization and human screen-reader checks are not performed in this workspace.

## Four independent gates

| Gate | Evidence and boundary |
| --- | --- |
| Design/accessibility | Keyboard-oriented reading workspace, source/proposal comparison, explicit review, labeled controls, responsive layout. Browser checks exercise the synthetic workflow and automated accessibility scanning; human assistive-technology qualification remains separate. |
| Education/privacy | No accounts, telemetry, remote fonts, hosted inference, or automatic external content upload. Assets are downloaded during setup; desktop inference calls only loopback; browser inference runs in workers. User-selected imports and explicit backups/audio downloads can contain sensitive text. School authorization is not inferred. |
| Security | Exact local host/origin and session key checks; bounded import/output; immutable source digest and passage validation; safe DOM text rendering; allowlisted download inventory; serial cancellable jobs; protected originals. Unit and regression checks accompany source. |
| Recoverable delivery | Pinned dependency hashes, source and component notices, project save/reopen, private temporary outputs, clear/cancel invariants, native runtime and package inventory verification. Published package checksums identify downloadable bytes. |

The local Linux integration test runs the actual Qwen model and speech voice on invented text and verifies complete structured output plus nonzero mono PCM. It is a small pilot, not a full-book fidelity or speech-quality evaluation. Model proposals still need review throughout every section. The native runner outcomes are recorded in GitHub Actions. Browser evidence is produced by the published reproducible Playwright scripts on local Chrome; verify the run/artifact identity associated with the release.

Windows x64, macOS Intel, macOS Apple Silicon, and Linux x64 each have a native build job. A package is published only after its build/test/package checks succeed. A hosted-runner CLI test is not a real native picker, downloaded-quarantine/Gatekeeper, ordinary-user installation, audiobook-player resume, or physical-device test. Untested combinations remain unqualified. Use source builds or institution-approved software channels where unsigned preview packages are unsuitable; do not disable operating-system protections globally.

## Limits

- Searchable PDFs and EPUBs can extract imperfect reading order. OCR, image interpretation, and DRM removal are not implemented. Preserve the source and inspect extracted text.
- The browser edition uses local WebGPU Qwen preparation and CPU/WASM speech after explicit downloads; caches can be evicted. Its audio jobs are limited to 100,000 UTF-8 text bytes and 96 MiB WAV. Browser output is WAV chapter ZIPs; MP3/M4B require the desktop edition or another local conversion tool.
- The bundled voice is English. Narration is limited to 300,000 UTF-8 bytes and 768 MiB of WAV output per job; save long books in parts. WAV chapter ZIPs include text and a playlist. MP3/M4B need a separately installed local FFmpeg.
- Models can omit or alter text; structured output, matching endpoints, and word counts do not certify fidelity. Side-by-side review and explicit approval remain essential.
- Refresh/close loses unsaved browser work. Save the project backup first and reopen it through the file selector. A refreshed desktop tab needs the private session URL from the launcher again.
- Normal clear/close cancels jobs and removes app-owned temporary audio. Abrupt power loss can leave temporary files in the OS temporary directory. No forensic erasure or removal of independent saved/backed-up copies is promised.
- The source remains public, with exact module and engine identities. No independent security audit, complete software bill of materials, publisher signing, or notarization is claimed.

Release decision: **CONDITIONAL for evaluation with authorized, nonsensitive text**, limited to the checks recorded for each package. Additional qualification is needed before sensitive school use or a trusted signed distribution channel.
