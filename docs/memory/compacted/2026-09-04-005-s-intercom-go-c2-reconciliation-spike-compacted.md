---
title: "Compacted memory — 005-S intercom-go C2 design reconciliation + Copilot SDK proving spike"
date: 2026-09-04
shipment: 005-S
feature: 006-F
status: shipped
pr_main: 17
pr_closure: 18
merge_commit_sha: 3fdb871b5367f7cde54af9406840f15669e4f1ca
compacted_from:
  - docs/archive/memory/2026-09-04-ship-005-s-session.md
---

# Compacted memory: 005-S / 006-F

**Verbose original:** `docs/archive/memory/2026-09-04-ship-005-s-session.md`
**PR:** #17, merged `2026-09-05T03:19:31Z`, merge commit `3fdb871`
**Shipment:** 005-S -- shipped, archived. **Feature:** 006-F -- done, archived.

## Decisions made

* Real live SDK proving chosen over the plan's zero-execution fallback (a
  genuine, authenticated Copilot backend connection was available).
* SQ-b answer (SERIALISED) adopted as design section 4.2's normative adapter
  rule, explicitly hedged as pin-specific / re-verify-before-upgrade.
* SQ-d answer (YES, stable `SessionEvent.ID`) adopted as section 5.1's
  primary de-duplication branch, but the buffer-then-reconcile fallback
  branch was **retained** (not deleted) pending a resume/`GetEvents`-path
  probe that was out of this spike's tested scope.
* SQ-a (NO) and SQ-c (YES) answers back-filled into sections 3.1 and 5.4's
  previously-open "assume it does not" / "until answered" language.
* P-015 fully-covered-root classification confirmed for 006-F (root feature,
  11 descendants at every depth, all in-manifest) -> closed via
  `backlogit shipment ship` directly rather than the generic `move` sequence
  (which backlogit 1.10.1 hard-rejects for shipment artifacts -- already
  documented in `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`).

## Files modified (by category)

* Docs: `docs/design-docs/intercom-architecture-implementation-design.md`
  (A1, additions-only, byte-for-byte-preserved base), 
  `docs/design-docs/intercom-go-backend-architecture.md` (A2 + B7 +
  remediation amendments across sections 3.1/4.2/5.1/5.4/8/9),
  `docs/decisions/2026-09-04-intercom-go-c2-sdk-spike-findings.md` (B1
  skeleton -> B6 complete -> remediation clarifications).
* Code: `internal/copilotprobe/{client,fixture,permission}.go` +
  `{permission,events,shutdown,concurrency}_test.go`,
  `go.mod`/`go.sum` (pinned `github.com/github/copilot-sdk/go@v1.0.11`).
* Backlog: 005-S + 006-F + 11 descendants archived; source artifact cleanup
  N/A (006-F had no `source_stash_id`/`source_deliberation_id` fields).

## Key learnings (see full compound entries)

* `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` --
  confirmed again this session; `backlogit move <shipment_id> --status
  shipped` always fails for shipment artifacts (exit 9); use `backlogit
  shipment ship` for a verified fully-covered-root feature, or
  `shipment-reconcile`'s Safe-Close Mode otherwise.
* `docs/compound/concurrency-issues/prefer-atomic-typed-fields-2026-09-04.md`
  (new) -- prefer `atomic.Int64`/`atomic.Int32` typed fields over a raw
  `int64`/`int32` field manipulated via `atomic.AddInt64`-style functions;
  caught by GitHub Copilot PR review, fixed in scope.

## Failed approaches / rejected paths

* None required a full retry cycle -- all builds/tests passed on first
  attempt after each fix; no circuit-breaker thresholds were approached.
* Considered but rejected: collapsing design section 5.1's de-duplication
  spec to a single branch unconditionally (adversarial review F4) --
  restored the fallback branch and downgraded to provisional instead.

## Review cycle outcome

Standard review (6 personas) -> P1/P2 remediated. Adversarial review
(5-reviewer panel, operator-mandated) -> initial BLOCKED on 9 findings, all
remediated, re-verified READY. Copilot review (PR #17) -> 2 comments, 1
fixed in scope (atomic alignment), 1 matched an existing P-021 deferred
entry (`A0A2D049`) and was disposed via reply + resolve, no new entry.

## Follow-up items (P-021 deferred stash entries, still open)

`A0A2D049` (opt-in gate for live-SDK tests), `6E701953` (unrelated
commit-message.instructions.md acronym drift), `309FBF5A` (C3 disposal
criterion / depguard for `internal/copilotprobe`).

## Next steps

None outstanding for 005-S itself (shipped, closed). Follow-ups above are
Stage-triage candidates for a future planning cycle. C3 is unblocked
(go/no-go: GO per the findings artifact).
