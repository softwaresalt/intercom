# 052-S (060-F) compacted memory

- Scope: pathsafe `addLongPathPrefix` byte-length fast-path guard. Tasks 060.001-T (characterization test) and 060.002-T (guard) done. PR #162 merged as a merge commit 640625cf5a411ca42a405bac954e108bf5ca17ef (head 0253e393, CI green, local review READY).
- Decision: the harness red phase cannot be confirmed because the change is verdict-neutral. The test passes before and after, and soundness is shown by mutation checks.
- Finding: the task's example mutation `len(path) < 130` is verdict-neutral (units <= bytes < 130 < 260) and cannot fail the test. The unsound mutations `len(path) < 2*longPathThreshold` and `len([]rune(path)) < longPathThreshold` both fail it.
- Stage-side context (originals archived): the stage session for the 052-S shipment is in `docs/archive/memory/2026-10-17-stage-pathsafe-fastpath-shipment-052s.md`.
- Closure: 060-F moved active->done via the a1 gate. Classify-only returned CASCADE (engine VERIFIED). The mutating cascade-close awaits per-closure operator approval.
