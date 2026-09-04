# Stage session — intercom-go P2 configuration package

- **Date:** 2026-09-03
- **Agent:** Stage (`claude-opus-5`, high; escalation route `gpt-5.6-sol`/openai/high, unused)
- **Mode:** standard sequential (not dark factory)
- **Branch:** `chore/stage-003-S` from `main` @ `5635e77`
- **Outcome:** shipment `003-S` queued and ready for Ship

## Scope

Single dependency-ready slice: **P2 configuration package** from stash `037B1552`, under
roadmap entry `4989A42D`. P3 credential resolution deliberately preserved as the next successor.

## Step trace

| Step | Result |
|---|---|
| 0.0 Tool gate | `TOOL_OK` — backlogit CLI 1.10.1 (manual/CLI mode, no MCP surface) |
| 0.1 Index sync | `INDEX_SYNC_OK` — 41 artifacts |
| Recovery | Zero checkpoints → normal startup, no recovery needed |
| 1 Triage | 16 active stash entries classified; 1 in scope, 15 deferred |
| 1.5 Grouping | **Skipped** — single feature-shaped entry, explicitly targeted (documented fallback) |
| 2 Deliberation | Addendum created; parent Unresolved Question 3 resolved |
| 3 Plan | Plan created; hardening signals `yes` → `## Plan Hardening` authored |
| 4 Review | Attempt 1 **FAIL** (1 P0, 8 P1) → remediated → attempt 2 **PASS** |
| 5 Harvest | 1 feature + 5 sub-epics + 13 units; P-003 chain validated |
| 5.5 Shipment | `003-S` assembled, 19 items, parent-first, verified by read-back |
| 5.6 Stash | `037B1552` updated (P2 consumed, P3 preserved, kept **active**); `4989A42D` status line refreshed |

## Oracle findings (softwaresalt/agent-intercom @ 41df772, read-only, unmodified)

Three quantitative claims carried in the stash were **wrong or misleading**:

1. **17 non-zero defaults, not 16** — `slack_detail_level = Standard` was missed. Would have
   left an enum silently at Go's zero value (`minimal`).
2. **"9 validation rules" is a total across three functions**; only 5 ran on the normal load
   path. The other 4 were ACP-mode-gated — and mode is retired.
3. **`channel_id` uniqueness is enforced**, at ACP session start (`slack/commands.rs:731`), not
   at hot reload. The parent deliberation had recorded it as simply unenforced.

Also confirmed: no `deny_unknown_fields` anywhere; `protocol_mode` is a DB column not a config
key; `team_id` **and** `channel_id` are unredacted in the oracle's `Debug`; there is no
`Secret<T>` type; all four secret-bearing fields are `#[serde(skip)]` and never read from TOML.

## Decisions

`P2-D1` BurntSushi/toml v1.6.0 · `P2-D2` decode over pre-populated `Default()` ·
`P2-D3` tolerant but observable unknown keys · `P2-D4` required scalars, relaxed sections ·
`P2-D5` one unconditional `Validate()` · `P2-D6` eager `channel_id` uniqueness ·
`P2-D7` `pathsafe.NewRoot` reuse · `P2-D8` case-fold collisions **rejected** ·
`P2-D9` no mode/protocol_mode.

`P2-D6` and `P2-D8` were **amended during plan review** and the addendum was reconciled.

## Review remediations that changed the design

- **P0** — `apperr` cannot express custom-message-plus-retained-cause; the plan's acceptance
  criterion was unimplementable under its own invariant I1. Verified against source. The oracle
  also stringifies, so cause retention was beyond-oracle invention. Added **Constraint K1**.
- **P1** — rule-ordering contradiction (found independently by two reviewers): a single mapping
  validator covering rules 3–5 *and* 8 cannot sit at one point in the sequence. Split into two
  unexported stages.
- **P1** — empty `default_workspace_root` silently binds to the process CWD via
  `filepath.Abs("")`. Go-specific hazard with no oracle analogue. Added rule 2a.
- **P1** — case-fold key collisions select a field nondeterministically; could shadow `host_cli`
  or the containment root. Inverted from "accepted" to "rejected".
- **P1** — `workspace.path` and `database.path` were unvalidated escapes that P6/P4 would
  inherit. Added rules 9 and 10.
- **P1** — the example config could not have passed its own test; the drift test was
  one-directional. Both fixed.

## Verification-only artifacts (not committed)

Library claims were verified against `BurntSushi/toml` v1.6.0 source over HTTPS rather than
assumed: `MetaData.IsDefined`, `MetaData.Undecoded`, and `unifyStruct` iterating only
document-present keys with no `reflect.Zero`. No scratch files were written.

## Preservation evidence

`.gitignore`, `.claude/`, and `.backlogit/hooks_queue.jsonl` were left untouched and unstaged
throughout. Both reference repositories (`C:\Source\GitHub\intercom`, `references/herdr`) remain
read-only and uncommitted.

## Next action

Ship claims `003-S`. Execute units in dependency order A1 → E3 on a single branch, single
worktree (P-016). Two operator checkpoints are recorded in the plan: before B1 (new dependency)
and before C4 (stricter-than-oracle containment rules).
