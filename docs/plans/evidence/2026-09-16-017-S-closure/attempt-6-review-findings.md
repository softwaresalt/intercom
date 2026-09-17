# Plan review attempt 6 — findings and remediation

**Artifact reviewed:** `docs/plans/2026-09-16-intercom-go-017-S-closure-plan.md` revision 7
**Canonical contract:** `docs/plans/2026-09-16-017-S-closure-canonical-contract.md`
**Reviewing SHA:** `2f72187`
**Dispatch mode:** multi-agent (3 personas, parallel, full-plan + full-contract)
**Decision: FAIL** — 4 distinct P0 findings
**Remediated into:** revision 8

---

## Persona verdicts

| Persona | Verdict | P0 | P1 | P2/P3 |
|---|---|---|---|---|
| Correctness Reviewer | **FAIL** | 2 | 10 | 9 |
| Constitution Reviewer | **FAIL** | 2 | 12 | 11 |
| Scope Boundary Auditor | **ADVISORY** | **0** | 4 | 9 |

Scope-boundary returning **no P0** is the first non-FAIL persona verdict in this lineage. It
confirmed the 13-member manifest holds on every traced path — happy path, both halt regimes, the
S-18 pre-mutation refusals, and R-1…R-8 — and that the `018.008-T` scope carve-out matches the
task's recorded scope exactly without over-granting.

---

## The four P0s — all independently verified by Stage against the workspace

None was a reviewer misreading.

### F-1 (P0) — the withdrawn allowlist was still normative

§6.1 withdrew the exhaustive changed-path allowlist, but **AC-18**, the §11 risk-register
mitigation, the §8.1 S-10 halt row and **contract §E rows S-10/S-22** still mandated it.

Not cosmetic: an executor conformant to AC-18 must confine S-10's changed paths to
`{harness_paths} ∪ {impl_paths}` plus three records. By D4, `{harness_paths}` excludes
`.backlogit/`, so the S-2 `mode: pre` reconcile report — mandated by S-2 and measured as
version-controlled by **M-18** — falls in *neither* part of AC-18's allowlist. AC-18 therefore
fails at S-10, **after the one-way door**, stranding `017-S` `active` in the P-001 slot.

*Verified:* AC-18 at L677-680, §11 L624, §8.1 L511, contract L137/L149.

### F-2 (P0) — R-5 had no reachable success path

R-5 targeted `{closure_merge_sha}` with `git revert -m 1`. But the closure branch carries **two**
commits: the S-15 baseline `{pre_cascade_sha}` (containing S-14's `018-F → done`) and the S-21
cascade `{cascade_commit_sha}`. The merge contains both, so `-m 1` unwinds the baseline as well —
returning `018-F` to `active`/queue-located. `{pre_paths}` was captured *after* the baseline
commit, so **R-6 (tree set-equality) and R-7 (record equality) are guaranteed to fail**, forcing
R-8's "HALT, no second automated unwind" on *every* execution of R-5.

R-4 was internally consistent (it targets only `{cascade_commit_sha}`, preserving the baseline);
the R-5 rewrite in revision 7 broke that symmetry. The defect was in the **contract** (D-K7 target
+ D-K9 anchor), not merely a plan divergence.

*Verified:* S-12 creates the branch, S-14 mutates, S-15 commits the baseline on it, S-21 commits
the cascade on it, S-22 merges it. Topology confirmed against plan L294/L310/L367/L370.

### C-01 (P0) — S-0 substituted notification for authorization

Installed **P-002 Violation Action is literally *"Halt and suggest running the harness-architect."***
S-0 read *"Proceeding requires the operator to have been informed"* — and was the only step in the
document with no pass criterion, no failure disposition, and no row in the §8.1 halt table, while
guarding the single most irreversible step.

No disposition in the installed registry authorizes proceeding past a Halt-mandated gate on
disclosure. Every override this registry defines is an **explicit recorded authorization**:
P-001 `skip_policy`, P-014 merge approval, P-018 audited force, P-012 declared degradation,
`PA-017-CASCADE` escrow.

*Verified:* `workflow-policies.md` §P-002; plan L193-197.

### C-02 (P0) — Ship self-applied `harness-ready`, bypassing P-004's producer

Installed **P-004 Violation Action is literally *"Do NOT apply `harness-ready` label."*** P-002
names the **harness-architect skill as the producer**; Ship is only the consumer. S-7 said
*"Apply the `harness-ready` label to `018.008-T` once red is confirmed"* — naming no producer,
requiring no `go vet`, no marker string, and recording no harness manifest. It was also weaker
than `018.008-T`'s own recorded **AC-13**.

Applying the label outside the producer's mandated procedure emits a **false compliance signal on
the exact label P-002 consumes** — worse than omitting it, and it voided §4.1's claim to preserve
P-002's protective intent.

*Verified:* `workflow-policies.md` §P-004; `.backlogit/queue/018.008-T.md` AC-13 / L101-106;
plan L234.

---

## Diagnosis — the dominant defect class recurred for a third revision

**Three of the four P0s were introduced by revision 7's own remediations.**

| P0 | Origin |
|---|---|
| F-1 | §6.1 withdrew the allowlist; AC-18, §11, §8.1 and contract §E were never propagated |
| F-2 | The R-5 rewrite re-targeted the revert without re-checking R-6/R-7's anchor |
| C-01 | The newly added S-0 gate specified disclosure where its governing policy mandates a halt |
| C-02 | Pre-existing, newly detected |

The lineage's repeated root cause is unchanged: **a fix is validated against the finding it
answers and never propagated to the clauses that reference the mechanism it changed.** Textual
re-reading has now failed three consecutive times to catch it.

What *did* work: the measurement pass. The two P0s it targeted (the `status: done → archive/`
routing collision and the tool-emitted-artifact collision) are gone, and **no measured `M-` row
was refuted by any reviewer**. The scope persona's move from FAIL to ADVISORY is attributable to
the same work.

---

## Countermeasure adopted for revision 8 — mechanism-reference closure

Every mechanism changed in revision 8 is treated as a **named token**. Every occurrence of that
token across **both** artifacts is enumerated by literal search, each occurrence is dispositioned,
and a **post-edit re-search proves zero residue**. This is mechanically checkable; "re-read the
document" is not.

**The sweep immediately justified itself.** After all targeted edits were applied and would
previously have been declared complete, the token sweep for `{closure_merge_sha}`-as-revert-target
found **three surviving stale references** that the edits had missed:

* plan L440 — S-22 still called `{closure_merge_sha}` *"R-5's revert target"*
* plan L738 — risk row still described `-m 1` parent-counting for R-4/R-5
* contract L166 — binding table still listed `{closure_merge_sha}` consumed at R-5 unqualified

Each of these would have been an attempt-7 P0 of exactly the F-2 class.

---

## Remediation applied (revision 8)

Canonical contract updated **first**, then the plan regenerated from it.

| Finding | Resolution |
|---|---|
| **F-1** | AC-18 rewritten to the §6.1 control verbatim; §11 risk row rewritten; §8.1 S-10 row now cites the D1/P-breach conditions; contract §E S-10/S-22 now read "§H changed-path control asserted". Sweep confirms no normative changed-path allowlist survives |
| **F-2** | **D-K7/R-5 retargeted to `{cascade_commit_sha}`** with a plain revert (never `-m 1`); `{closure_merge_sha}` demoted to a two-parent *evidence* check; D-K9/R-6 anchor relationship stated explicitly; new **I-12** protects the invariant; AC-19 rewritten |
| **C-01** | **D-D6 and S-0 converted to a fail-closed operator-authorization gate** reusing the existing `skip_policy: P-002` shape (no new token); S-0 row added to the §8.1 halt table; AC-26 rewritten to verify the authorization, not the notification; new **I-13** |
| **C-02** | **New D-D7 and §4.2**: the `harness-architect` skill is the sole applier; S-7 rewritten to invoke it with P-004's precondition plus AC-13's literals verbatim; producer probed (P-012) because the task is already `active`; new **AC-29**, **I-11** |

Convergent P1/P2 findings also resolved: D1 deny-list extended with governance/control-plane
surfaces (F-8/SB-01); D2 closed to three literally named paths (F-6/SB-06/C-18); new **P6**
sub-file content lock (F-7/SB-02/C-06); P4 scoped to modes that persist a report (F-4/SB-07);
D3 `updated_at` allowance (F-9); R-1 `{post_paths}` anchor corrected (F-10); P-005's third
required action restored (C-03); R-5 P-011/P-016 prechecks (C-04); §14.2 R-5 risk row split
(C-05); §13 constitution map rewritten with the correct **P-015** (single-artifact shipment
closure, this route's authorizing exception — previously mislabelled) plus new **P-020**, P-012,
P-008 rows (F-5/C-09/C-11); escrow condition 9 re-anchored to PF-4 in both documents (C-12);
P-021 C2 six-field payload (C-10); new step **S-21.5** authoring the closure artifacts and
discharging P-020; M-19/M-22/S-2 measurement rows corrected; `revision: 8`; P5 made independent
of the changed-path enumeration (SB-05); D1(e) worktree clause relocated to where it is decidable
(F-13).

---

## Mechanical validation after remediation

* AC list strictly sequential **AC-1 … AC-31**, no duplicate or dangling references
* **I-1 … I-13 parity** between plan and contract (previously plan stopped at I-8)
* Steps **S-0 … S-22 plus S-21.5**, matching the contract's declared sequence
* 29 M-rows, no dangling `M-` reference
* No orphan line fragments, no truncated table rows
* Token sweeps clean for: changed-path allowlist, revert target, `harness-ready` applier,
  `disclosed-deviation`, "each mode invoked", `P1–P5` arity

---

## Status

`017-S` remains **NOT claimable**. Eligibility ground 3 (a review-PASSed governing closure plan
present on `origin/main`) is still **OPEN** — it is the sole remaining blocker. Grounds 1, 2 and 4
are discharged.

Revision 8 is remediated but **unreviewed**. Attempt 7 requires explicit operator authorization.
