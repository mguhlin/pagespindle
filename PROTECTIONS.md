# Four protections and Go platform contract

PageSpindle's primary task is authorized text → reviewable preparation → local narration → saved audiobook. Adults and educators are the intended initial audience. No student record service, accounts, remote inference, telemetry, or automatic publication is part of the application.

| Gate | Implementation contract | Required evidence |
| --- | --- | --- |
| [Design and accessibility](https://mguhlin.github.io/vcc/WEB_DESIGN.md) | Reading workspace; paste/open, source comparison, explicit review, voice preview, export. Keyboard access, labeled controls, reflow, progress and recovery. | Browser tasks, contrast/axe checks, narrow layout, error flow; assistive-technology gaps named separately. |
| [Education privacy](https://mguhlin.github.io/vcc/EDU_COMPLIANCE.md) | Input files are read through explicit browser file selection. Text stays in the local app and its bundled engines. Work stays in memory/private temporary output until explicit download. | Network/storage inventory, clear/reset, export disclosure, and no student data in fixtures. Local use is not blanket school authorization. |
| [Security](https://mguhlin.github.io/vcc/securevcc/securevcc.md) | Loopback binding; session key, exact host/origin checks, bounded inputs, safe DOM rendering, no generic read/write/execute bridge. Untrusted documents/model output cannot grant authority. | Allowed and denied requests; malformed/oversized imports; stale jobs; safe output names; no outside-tree writes. |
| [Rescue and recoverable delivery](https://mguhlin.github.io/vcc/rescue/VIBE_CODE_RESCUE.md) | Preserve source; proposals require review; cancellation, truthful errors, saved project backups and bounded progress. Pinned engines, model identities and package inventories. | Save/reopen, failure/retry, cancelled jobs, reproducible build, actual extracted-package operation. |

[Go App Builder](https://mguhlin.github.io/vcc/skills/go-app-builder/SKILL.md) governs separation between pure Go text/session logic, browser view, native subprocesses, platform adapters, and release evidence. The app is a Go executable serving an embedded interface in the user's local browser, not a hosted processing service or native-webview claim. It requires a supported browser. Native engines are separate bundled components, not Go implementations of neural inference.

Initial package targets: Windows x64; macOS Intel and Apple Silicon; Linux x64. Each needs its matching runtime folder and the common model pack. Offline processing, offline first launch after extraction, clean install, and offline source builds are separate claims. Extra architectures are not promised through cross-compilation alone.

The design uses a book-preparation workspace rather than a marketing dashboard: navy reading text, white editing surfaces, pale silver dividers, teal actions, coral warnings, gold page markers, system sans-serif controls and Georgia reading text. Source and prepared passages remain adjacent on desktop and stack on narrow screens. Generated paper-sculpture artwork communicates the book-to-sound workflow. Motion is unnecessary. Local inputs and exports are sensitive even without a server account.

Release evidence belongs in `VALIDATION.md`. A passing test is scoped evidence; no certification, complete security, legal compliance, or untested-platform claim follows from it. Preserve substantive source/license dates and stable URLs; omit editorial maintenance timestamps.
