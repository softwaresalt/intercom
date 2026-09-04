---
title: "Implementation Plan — intercom-go P3: Credential Resolution"
description: "Port the agent-intercom four-tier keychain-first credential chain into internal/credentials behind a provider seam, with a reusable totally-redacting Secret primitive and classified keychain errors"
date: 2026-09-03
status: draft
revision: 3
phase: P3
source_deliberation: "docs/decisions/2026-09-03-intercom-go-p3-credentials-deliberation.md"
parent_deliberation: "docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md"
source_brief: "docs/decisions/2026-07-06-go-port-reference-brief.md"
behavioral_oracle:
  repo: "softwaresalt/agent-intercom"
  commit: "41df772"
  access: "read-only, out-of-tree"
stash_entries: ["037B1552"]
predecessor: "003-F / 003-S (P2) — merge 40c7d289, closure 2a547913"
successor: "P4 models + SQLite risk slice (stash 4989A42D)"
prior_closure: "docs/closure/003-S-003-F-post-merge-closure.md"
requires_plan_hardening: yes
plan_hardened: yes
plan_review_verdict: FAIL
gate_status: blocked-circuit-breaker
plan_review_attempts: 3
escalation_route: "gpt-5.6-sol / openai / high"
tags: ["go", "port", "credentials", "keychain", "secret", "redaction", "p3"]
---

<!-- plan-review-attempt: 3 -->

# Implementation Plan — intercom-go P3: Credential Resolution

> **Revision 3.** Rev 1 failed all five adversarial personas; rev 2 remediated those findings and
> passed Architecture and Schema-CLI-Docs but failed Security, Go, and Scope. Rev 3 remediates the
> remaining findings and **adjudicates a genuine conflict between reviewers**: Scope demanded the
> removal of controls that Security and Schema-CLI-Docs had specifically required. Adjudications
> are recorded explicitly in the *Review Remediation Record* rather than resolved by silently
> siding with one persona. Four findings across the three revisions were **empirically verified**
> by running Go programs against the pinned `go1.26.5` toolchain rather than accepted on argument.

## Problem Frame

Shipment `003-S` delivered `internal/config`: the `config.toml` schema, 17 non-zero defaults,
tolerant decoding, and one unified validation pass. P2 fixed a **binding, permanent boundary**
(P2 plan Decision R7): `internal/config` is a secret-free file-schema package. The four
credential-bearing fields — `app_token`, `bot_token`, `team_id`, `authorized_user_ids` — are
`#[serde(skip)]` in the oracle and are never read from `config.toml`.

P3 delivers credential resolution: the oracle's precedence chain, a source abstraction spanning
OS keychain and environment variables, and a reusable redacting `Secret` primitive that the
oracle does not have.

Every contract below is fixed by the P3 deliberation, which re-read the oracle specifically for
this slice and **corrected six claims** carried in stash `037B1552` (F2, F3, F6, F7, F8, F9). The
corrections are load-bearing: the largest, F2, means a single generic resolver over all four
credentials would be **wrong**.

## Scope Boundaries

**In scope**

* `internal/secret`: the reusable `Secret` redaction primitive. Extracted on **present-tense
  cohesion** grounds — it has no dependency on credential resolution, and `internal/credentials`
  would otherwise hold four unrelated responsibilities. (P13's `ipc_auth_token` is a corroborating
  future consumer, not the justification — see the SCOPE-12 adjudication.) There is **no**
  `type Secret = secret.Secret` alias: callers import `internal/secret` directly, so one type has
  exactly one import path.
* `internal/credentials`: typed errors, the `Source` seam, environment and keychain sources, a
  reusable fake, the four-tier resolver, the authorized-users resolver, the `Resolved` aggregate,
  and source attribution.
* One new direct dependency: `github.com/zalando/go-keyring` v0.2.8.
* A `loadStartup` composition helper in `cmd/intercom`, **called by `RunE`** (rev 3 — see the
  SCOPE-1 adjudication).
* Negative-control leak tests, boundary guards, a secret-scan **Go test** (not a CI script — see
  the SCOPE-13 adjudication), and documentation (`docs/credentials-reference.md`, `README.md`,
  `config.toml.example`, `docs/config-reference.md`).

**Explicitly out of scope**

| Excluded | Owner |
|---|---|
| `internal/models`, SQLite, persistence | P4 — stash `4989A42D` |
| Slack client, Socket Mode, ACP subprocess | P6–P9 — stash `4989A42D` |
| Calling `loadStartup` from `RunE` | **In scope as of rev 3** — see E8 |
| Any server behavior beyond startup validation (`RunE` still returns `errNotImplemented` after validating) | P4+ |
| Keychain **write** path **in product code** (docs may describe operator provisioning) | Not ported — the oracle has none |
| `AppState::ipc_auth_token` resolution | P13 (it will consume `internal/secret`) |
| ACP spawner env-clear allowlist | P6 |
| Credential hot-reload | Not ported — the oracle never reloads credentials |
| A `credentials check` / doctor CLI subcommand | Deferred to the phase that wires `RunE` (Decision R12) |
| Oracle-drift CI gate | stash `8ACF7110` |
| CI-hardening follow-ups, P1/P2 review follow-ups | separate stash entries |
| Multiplexer architecture | stash `A92E3FA0` |
| Operator credential rotation | stash `EF9352FB` |
| Any change to `internal/config` behavior (one test is added; no production code changes) | frozen by P2 |
| Any modification to `C:\Source\GitHub\intercom` or `references\herdr` | read-only / untouched |

## Requirements Trace

| ID | Requirement | Oracle anchor | Unit |
|---|---|---|---|
| R1 | Four-tier chain, keychain-first, in fixed order | `config.rs:857-928` | D1 |
| R2 | Tier order: `agent-intercom-acp` kc → `*_ACP` env → `agent-intercom` kc → shared env | `config.rs:865-911` | D1 |
| R3 | Empty value is treated as absent at every tier | `config.rs:837-840,878,902` | C2, C3, D1 |
| R4 | Per-field resolution; tiers may differ across fields | `config.rs:452-460` | D4 |
| R5 | `app_token`, `bot_token` required; failure aborts | `config.rs:454-455` | D4 |
| R6 | `team_id` optional, never fails, defaults to `""` | `config.rs:928-941` | D4 |
| R7 | `authorized_user_ids` is **env-only, two-tier**, CSV-parsed, required non-empty | `config.rs:484-508` | D2 |
| R8 | Keychain accounts: `slack_app_token`, `slack_bot_token`, `slack_team_id` | `config.rs:454-457` | D1 |
| R9 | Env names `SLACK_APP_TOKEN`, `SLACK_BOT_TOKEN`, `SLACK_TEAM_ID`, `SLACK_MEMBER_IDS` (+ `_ACP`) | `config.rs:454-457,486` | D1, D2 |
| R10 | Errors name the source, never the value | `config.rs:913-925` | B3 |
| R11 | Error kind is config-class, rendered `config: {msg}` | `errors.rs:11-12,44` | B3 |
| R12 | Tokens redacted in all rendered forms | `config.rs:133-147` (improved) | S1–S3, D3 |
| R13 | Credential provenance is **recorded as structured data** for the caller to log; the library never logs | `config.rs:867-908` + `config.Report` Principle V | B2, E4 |
| R14 | Credentials never sourced from `config.toml` | `config.rs:115-121,354` | E3 |
| R15 | Resolution runs after config load+validate, once, before anything else | `main.rs:130-132` | E4 |
| R16 | Keychain unavailable ≠ credential absent (oracle defect F6, fixed) | improvement | C3, B1 |

## Implementation Units

Sizes are t-shirt effort; complexity is difficulty/uncertainty. Both axes are independent.

**Granularity rule, stated honestly (SCOPE-8).** Rev 2 claimed "≤5 exported functions per unit",
which was false on its face: `S2` defines seven methods on one type. The real, checkable rule is:
**≤3 production files, ≤1 exported type's method set or ≤5 free functions, and ≤4 named test
functions** per unit. A cohesive method set on a single type in a single file is one unit of work,
not seven; a table-driven test's rows are one scenario group, and the per-unit named-test-function
count is stated explicitly where it is not obviously 1. Test files are counted in the file budget.

**Every unit is test-first** (constitution requirement): its test file is written first, observed
failing for the intended reason, then made to pass.

**Test package placement (GO-16, ARCH-15).** `credtest` imports `internal/credentials`, so any
test consuming the fake must be declared `package credentials_test`; tests asserting on unexported
state (`Ref`/`Result` fields, `chain()`, `authorizedUserIDs`) must be declared `package
credentials`. Each unit below states which it uses. `ConformanceTest(t, Source)` lives in
`credtest`, never in a production file, so `testing` is never linked into the shipped package.

### Sub-epic S — `internal/secret` (reusable redaction primitive)

Extracted from `internal/credentials` per ARCH-2: `Secret` has no dependency on credential
resolution, and P13's `ipc_auth_token` is a named second consumer in a different subsystem.
`internal/credentials` re-exports it as `type Secret = secret.Secret` so there is exactly one type.

**S1 — Pointer-boxed `Secret` core** · size `S` · complexity `medium`
Files: `internal/secret/secret.go`.
```go
// The value is held behind a pointer so that reflection-based rendering —
// fmt's bad-verb path, which bypasses Formatter entirely, and traversal of an
// unexported field, where CanInterface() is false — can only ever print an
// address. Removing this indirection reintroduces a verified plaintext leak.
type Secret struct{ v *string }
```
`const Redacted = "[REDACTED]"`. `New(string) Secret`, `Expose() string` (nil-safe → `""`),
`IsEmpty() bool`, `Equal(Secret) bool` using `crypto/subtle.ConstantTimeCompare`.
Pointer boxing redefines equality: `New(x) == New(x)` is **false** for identical `x`, and two
equal secrets are two distinct map keys. The doc comment states that `==` compares pointers, that
`Equal` is the only value-equality path, and that `Secret` must never be used as a map key or set
element in product code (the `map[Secret]string` row in S3 exists solely to prove the map-key
render path is safe). `Equal` is constant-time only for equal-length inputs —
`ConstantTimeCompare` early-returns on a length mismatch — which is inherent and documented.
Acceptance: (i) `New(x).Expose() == x` for the canary and for `""`; (ii) `Secret{}.Expose() == ""`,
`Secret{}.IsEmpty()` is true, and `Secret{}.Equal(Secret{})` and `Secret{}.Equal(New(""))` are both
true, with no panic; (iii) `Equal` is true for equal values across distinct instances and false
otherwise; (iv) a test asserts `New(x) != New(x)` **and** `New(x).Equal(New(x))` is true, pinning
the pointer-identity semantics as an executable assertion rather than a doc comment (GO-17, GO-21).

**S2 — Total redaction surface** · size `S` · complexity `high`
Files: `internal/secret/render.go`.
Value receivers on **every** method (a pointer receiver would silently un-redact every `Secret`
value — GO-5). `Format(fmt.State, rune)` writing `Redacted` for every verb, except `%q`, which
writes `strconv.Quote(Redacted)` so composed output stays well-formed (GO-12). Width and
precision are deliberately ignored, documented. Plus `String()`, `GoString()`, `MarshalJSON()`
(→ `"[REDACTED]"`), `MarshalText()` (load-bearing: `encoding/json` encodes **map keys** only via
`TextMarshaler`), `LogValue() slog.Value`, and `GobEncode() ([]byte, error)` returning an error so
`encoding/gob` fails closed rather than serialising the value.
**No `UnmarshalJSON`/`UnmarshalText`** — deliberately absent, so a `Secret` cannot be populated by
decoding a file.
Compile-time assertions in the file:
```go
var (
	_ fmt.Formatter          = Secret{}
	_ fmt.Stringer           = Secret{}
	_ fmt.GoStringer         = Secret{}
	_ json.Marshaler         = Secret{}
	_ encoding.TextMarshaler = Secret{}
	_ slog.LogValuer         = Secret{}
	_ gob.GobEncoder         = Secret{}
)
```
Acceptance: (i) the assertion block compiles, proving value-receiver method sets; (ii)
`fmt.Sprintf("%q", s)` yields `"[REDACTED]"` **with** quotes; (iii) `json.Marshal` yields exactly
`"[REDACTED]"`, and `gob` is asserted in **three** distinct cases (GO-23) — encoding a bare
`Secret` errors, encoding a struct with an **exported** `Secret` field errors and emits no partial
value bytes, and encoding a struct with an **unexported** `Secret` field silently omits the field
without error, so "fails closed" is never conflated with "silently omitted"; (iv) a test asserts
`Secret` does **not** implement `json.Unmarshaler` or `encoding.TextUnmarshaler` via a negative
type assertion.
Named test functions: 2.

**S3 — Leak-matrix negative controls** · size `M` · complexity `high`
Files: `internal/secret/leak_test.go`.
One table-driven test over every rendering path, asserting the canary
`xoxb-CANARY-MUST-NOT-APPEAR` is absent **and** `[REDACTED]` (or an address) is present, so a
silently-empty render cannot pass. Rows must include the paths rev 1 missed and this revision
verified as real leaks:
`%v %s %q %#v %+v %d %x %X %T %p`; `fmt.Errorf("...: %w", s)`; `errors.Join`; a `Secret` in an
**unexported** field of another struct; `[]Secret`; `map[string]Secret`; `map[Secret]string`
(map-key path); a nil `*Secret`; `encoding/json`; `encoding/gob`; `text/template`; `slog`
`JSONHandler` **and** `TextHandler`; `slog` with a `ReplaceAttr` that stringifies.
Acceptance: (i) every listed row asserts canary absence; (ii) the `%p`, `%w`, and
unexported-field rows are present and pass; (iii) two **executable** mutation guards, not
comments (SCOPE-7): a `go:build mutation`-tagged variant of the type is compiled in a sub-test
that removes `Format` and one that removes the pointer indirection, each asserted to leak — so
the negative control re-runs on every CI run instead of being a historical note; (iv) one table,
one named test function, plus the two mutation sub-tests.

### Sub-epic B — Attribution vocabulary and errors

**B1 — Closed outcome/reason/ref vocabulary** · size `S` · complexity `medium`
Files: `internal/credentials/attribution.go`.
Closed types so an invalid state cannot be constructed (SEC-1, GO-6, ARCH-4):
```go
type Outcome uint8
const (
	Miss Outcome = iota // zero value is Miss, never Hit: a zero Result is never success-shaped
	Hit
	Unavailable
)

type Reason uint8 // closed enum; `Reason: err.Error()` must not compile
const (
	ReasonNone Reason = iota
	ReasonNoEntry
	ReasonEmptyEntry
	ReasonUnsupportedPlatform
	ReasonBackendError
	ReasonCanceled
	ReasonTimeout
)

type Kind uint8 // SourceKeychain | SourceEnv

type Ref struct{ kind SourceKind; service, key string } // built by EnvRef / KeychainRef only
func EnvRef(name string) Ref
func KeychainRef(service, account string) Ref
func (r Ref) Kind() SourceKind
func (r Ref) Service() string
func (r Ref) Key() string
func (r Ref) String() string // source descriptor; can never hold a value
```
The type is named `SourceKind`, not `Kind`, because this package also imports `apperr.Kind`
(GO-20). Accessors are declared here because C4 consumes them.
Acceptance: (i) `Outcome(0) == Miss`, asserted; (ii) `Reason` has a `String()` covering every
constant and a test asserts totality by iterating the constant range and requiring no result
matches `Reason(%d)`; (iii) `EnvRef("X").Service() == ""` and `KeychainRef(s,a).Service() == s`,
and a `go/types`-based test asserts no file outside the package constructs a `Ref` composite
literal (replacing rev 2's non-executable "compile-time check" criterion — GO-21, SCOPE-7);
(iv) `Ref.String()` output contains the service and key but is asserted not to contain a canary
when one is passed as a key.
Named test functions: 3. Test package: `credentials` (asserts unexported fields).

**B2 — Attribution records** · size `M` · complexity `medium`
Files: `internal/credentials/record.go`.
`Attempt{Ref, Outcome, Reason}`, `FieldAttribution{Field, Resolved, From, Attempts}`, and
`Attribution{Fields}`. `AnyUnavailable()` and `SourceCoherence()` are **methods computed from
`Fields`**, not stored fields, so the two representations cannot disagree (ARCH-17). Implements
`LogValue() slog.Value` built from **already-terminal** values (`slog.String`, `slog.Int`) so it
does not depend on nested resolution (GO-5).
`SourceCoherence()` broadens rev 2's `MixedNamespace` per SEC-6: it reports the set of distinct
`(SourceKind, service)` tiers across **all three** chain-resolved credentials — `app_token`,
`bot_token`, **and** `team_id` — so it detects keychain-vs-env mixing *within* one namespace as
well as cross-namespace mixing. Rev 2's flag compared only the two tokens and only by namespace,
missing both cases.
Per ARCH-1 this is the **sole** diagnostic egress; no unit in this plan logs from library code.
Acceptance: (i) attempts are ordered by tier and a test asserts the exact recorded sequence;
(ii) `SourceCoherence()` is table-tested across all-same, cross-namespace, same-namespace
keychain-vs-env, and one-unresolved cases; (iii) `LogValue()` output contains no value and no
`Secret`; (iv) a reflection test over `Attribution`'s own type asserts no field can hold a
`secret.Secret` or a `[]string` of IDs.
Named test functions: 4. Test package: `credentials`.

**B3 — Typed credential errors** · size `M` · complexity `high`
Files: `internal/credentials/errors.go`.
The rev-1 design was **verified unimplementable** (GO-3). Use a composite carrying both chains:
```go
// resolveError carries the apperr taxonomy error (for kind matching and the
// frozen "config: {msg}" rendering) and a credentials sentinel (for call-site
// classification). errors.Is traverses Unwrap() []error since Go 1.20.
type resolveError struct{ app *apperr.Error; sentinel error }
func (e *resolveError) Error() string   { return e.app.Error() }
func (e *resolveError) Unwrap() []error { return []error{e.app, e.sentinel} }
```
Sentinels `ErrCredentialNotFound`, `ErrNoAuthorizedUsers`, and `ErrResolutionCanceled`, the last
declared as `fmt.Errorf("credential resolution canceled: %w", context.Canceled)` so a two-child
`Unwrap` still satisfies both the sentinel match and `errors.Is(err, context.Canceled)` (GO-15).
Constructors return `error` (not the concrete type) to avoid the typed-nil trap, and are the only
way to build a `resolveError`, so `e.app` is never nil.
**The not-found message is derived from the recorded `[]Attempt`, not from a literal** (ARCH-9),
so the chain has exactly one source of truth. It renders the oracle form:

```text
credential `slack_app_token` not found: checked keychain services `agent-intercom-acp` and
`agent-intercom`, and environment variables `SLACK_APP_TOKEN_ACP` and `SLACK_APP_TOKEN`
```

When any attempt was `Unavailable`, a clause naming the classification (never backend text) is
appended.
Acceptance: (i) `errors.Is(err, ErrCredentialNotFound)` **and** `errors.Is(err, apperr.ErrConfig)`
**and** `errors.As(err, &*apperr.Error)` are all true for one error value — **verified against the
real `apperr` package before this plan was accepted** (see Review Remediation Record); (ii) the
rendered message enumerates exactly the refs in the attempt list, in order — a test adds a
synthetic fifth tier and asserts the message changes with **no edit to `errors.go`**; (iii) the
canary never appears in any constructed error, table-tested, including an `Unavailable` clause
built from a backend error whose text embeds a canary; (iv) a cancelled resolution returns an
error matching **both** `ErrResolutionCanceled` and `context.Canceled`, and **not**
`ErrCredentialNotFound`.
Named test functions: 4. Test package: `credentials`.

### Sub-epic C — The source seam

**C1 — `Source` interface and constructor-only `Result`** · size `S` · complexity `medium`
Files: `internal/credentials/source.go`.
```go
type Source interface {
	Lookup(ctx context.Context, ref Ref) Result
}

// Result is constructible only through the three constructors, which make the
// "non-Hit carries no value" invariant structural rather than documented.
type Result struct{ outcome Outcome; value secret.Secret; reason Reason }
func HitResult(v secret.Secret) Result   { return Result{outcome: Hit, value: v} }
func MissResult(r Reason) Result         { return Result{outcome: Miss, reason: r} }
func UnavailableResult(r Reason) Result  { return Result{outcome: Unavailable, reason: r} }
func (r Result) Outcome() Outcome        { return r.outcome }
func (r Result) Value() secret.Secret    { return r.value }
func (r Result) Reason() Reason          { return r.reason }
```
`Source.Kind()` is **removed** (ARCH-3): the resolver pairs each source with its ref at chain
construction, so the discriminator has no caller. The interface is read-only by construction —
no `Set`, `Delete`, or `DeleteAll`.
Acceptance: (i) `Result` fields are unexported, so a caller cannot construct a
`Miss`-with-a-value; (ii) a shared `ConformanceTest(t, Source)` helper — homed in the `credtest`
package, never in a production file, so `testing` is never linked into the shipped package
(GO-16, ARCH-15) — asserts the invariant and is exercised by every source; (iii) the zero
`Result` reports `Miss` with a zero value; (iv) a reflection test on the interface type asserts it
declares exactly one method, so no write method can be added without failing CI.
Named test functions: 2. Test package: `credentials` (unexported fields).

**C2 — Environment source** · size `S` · complexity `low`
Files: `internal/credentials/source_env.go`.
`type envSource struct{ lookup func(string) (string, bool) }`, defaulting to `os.LookupEnv`. The
injectable field lets precedence tests avoid `t.Setenv` entirely and stay parallel-safe (GO-13).
Empty string maps to `MissResult(ReasonEmptyEntry)` (R3); env lookup can never be `Unavailable`.
Acceptance: (i) unset → `Miss`/`ReasonNoEntry`; (ii) set-but-empty → `Miss`/`ReasonEmptyEntry`;
(iii) set → `Hit` with the exact value via `Expose()`; (iv) passes `ConformanceTest` and never
returns `Unavailable` across all rows.

**C3 — Keychain adapter and error classification** · size `M` · complexity `high`
Files: `internal/credentials/source_keychain.go`; `go.mod`, `go.sum` (dependency addition).
```go
type getFunc func(service, user string) (string, error)
type keychainSource struct{ get getFunc; timeout time.Duration }
```
The `get` field is set at construction (default `keyring.Get`) — **not** a package-level `var`,
which `technology-go.instructions.md:107` forbids and which would race under `-race` (GO-10,
ARCH-8). Classification uses `errors.Is`, never `==`, so a wrapped sentinel from any platform
provider is still classified correctly (GO-7):

| Library outcome | Our outcome | `Reason` |
|---|---|---|
| value, non-empty | `Hit` | `ReasonNone` |
| value, empty | `Miss` | `ReasonEmptyEntry` |
| `errors.Is(err, keyring.ErrNotFound)` | `Miss` | `ReasonNoEntry` |
| `errors.Is(err, keyring.ErrUnsupportedPlatform)` | `Unavailable` | `ReasonUnsupportedPlatform` |
| any other error | `Unavailable` | `ReasonBackendError` |

The raw backend error is **never** wrapped into `Reason` — a backend may echo the requested
account or arbitrary daemon text.
Acceptance: (i) all five rows classified, including a **wrapped** `ErrNotFound`
(`fmt.Errorf("dbus: %w", keyring.ErrNotFound)` → `Miss`); (ii) the `get` field is non-nil in every
test and the E3 guard-2 test independently proves `keyring` is imported by exactly one file — so
"no live keychain I/O" is enforced by construction plus a guard rather than by assertion prose
(SCOPE-7); (iii) a stub error whose text embeds a canary is asserted absent from both `Reason`
and the rendered error; (iv) passes `ConformanceTest`.
Named test functions: 3. Test package: `credentials`.

**C4 — Bounded lookup with real cancellation** · size `S` · complexity `high`
Files: `internal/credentials/source_keychain_timeout.go`.
Rev 1 claimed a pre-flight `ctx.Err()` check "bounded" the call. It does not — Go cannot cancel a
blocking D-Bus round trip or an `exec` inside the dependency (GO-4, SEC-4). Off-load and select:
```go
if ctx == nil { ctx = context.Background() }                 // GO-19: WithTimeout panics on nil
ctx, cancel := context.WithTimeout(ctx, s.timeout)           // s.timeout defaults to 5s
defer cancel()

ch := make(chan got, 1) // buffered: the worker can always send and exit even
                        // after we abandon it, so no goroutine leaks permanently
go func() { v, err := s.get(ref.Service(), ref.Key()); ch <- got{v, err} }()
select {
case r := <-ch:      return classify(r.v, r.err)
case <-ctx.Done():   return UnavailableResult(reasonFor(ctx.Err()))
}
```
`reasonFor` maps `context.DeadlineExceeded → ReasonTimeout` and `context.Canceled →
ReasonCanceled`; both are `Unavailable`, never `Miss`, so PI-4 holds and the chain continues.
**In-flight bounding (SEC-14):** the resolver is documented and asserted **single-use** — one
`Resolve` per process at startup — and the source caps concurrent outstanding lookups at one per
`Ref` via a guard map, so a wedged backend cannot accumulate workers across repeated calls. A
wedged `get` still holds its plaintext until it returns; that residual is documented alongside
R11's memory-residency boundary.
Acceptance: (i) a `get` that blocks past the timeout yields `Unavailable`/`ReasonTimeout` within
the budget; (ii) an already-cancelled context yields `Unavailable`/`ReasonCanceled` and the
resolver still falls through to the next tier; (iii) `-race` clean, and a repeated-timeout test
asserts the outstanding-worker count is capped rather than growing linearly with call count;
(iv) a nil context does not panic.
Named test functions: 4. Test package: `credentials`.

**C5 — Reusable deterministic fake and conformance helper** · size `S` · complexity `low`
Files: `internal/credentials/credtest/fake.go`, `internal/credentials/credtest/conformance.go`.
Rev 1 put the fake in `fake_test.go`, which cannot be imported by `tests/integration` (SCOPE-10).
A dedicated `credtest` package is importable by both. Programmable per-`Ref` outcomes plus an
ordered call log, so tests can prove both what was returned **and** that later tiers were never
consulted. `ConformanceTest(t, Source)` lives here (GO-16).
`credtest` ships in the module, so its test-only status is **guarded, not assumed** (SEC-16): the
E3 guard proves every importer of `credtest` is either a `_test.go` file or the
`tests/integration` package.
Acceptance: (i) programmable `Hit`/`Miss`/`Unavailable` per `Ref`; (ii) records call order and
exposes it; (iii) unprogrammed refs default to `Miss`; (iv) `credtest` itself passes
`ConformanceTest`.
Named test functions: 2. Test package: `credtest_test`.

### Sub-epic D — Resolution

**D1 — Chain construction and the tier walk** · size `M` · complexity `medium`
Files: `internal/credentials/chain.go`, `internal/credentials/resolve.go`.
`ServiceACP = "agent-intercom-acp"`, `ServiceShared = "agent-intercom"`, `EnvSuffixACP = "_ACP"`.
Sources and refs are **paired at chain construction** so a mismatched call is unrepresentable
(ARCH-3):
```go
type step struct{ src Source; ref Ref }
func (r *Resolver) chain(sp spec) []step // exactly 4 steps, R2 order
```
The walk returns on the first `Hit` and records an `Attempt` per tier. It performs **no logging**
(ARCH-1, Principle V) — provenance leaves via `Attribution` only.
Acceptance: (i) `chain()` returns exactly the four R2 refs in order, asserted element-by-element;
(ii) first `Hit` wins and later tiers are never consulted, proved by the fake's call log;
(iii) every attempt is recorded with its outcome and reason; (iv) a grep-based test asserts the
`internal/credentials` package imports neither `log` nor `log/slog` outside `LogValue` methods.

**D2 — Authorized-users resolver** · size `M` · complexity `medium`
Files: `internal/credentials/authusers.go`.
The **separate** env-only, two-tier path (R7, correction F2): `SLACK_MEMBER_IDS_ACP` then
`SLACK_MEMBER_IDS`, empty treated as absent; CSV split on `,`, trim, drop empties. Never consults
the keychain.
**Adjudication (SEC-7 vs SCOPE-17).** Security asked for strict ID-shape rejection; Scope objected
that rejection is a behavioral divergence beyond a source/precedence phase. Rev 3 splits the
difference along the parity line: **acceptance semantics stay oracle-identical** — any non-empty
trimmed token is accepted, so no previously-working configuration breaks — while the
non-behavioral hardening is kept: de-duplication preserving first-seen order, an 8 KiB raw-input
cap, and a 1000-ID cap (both pure DoS bounds that no valid configuration reaches). IDs not
matching `^[UW][A-Z0-9]{2,20}$` are **recorded in attribution as non-conforming, not rejected**,
giving operator visibility without changing behavior. Strict rejection is reassigned to P9/P10,
the phase that actually consumes the ACL for authorization.
Acceptance: (i) the keychain source is never invoked — fake call log is empty; (ii) `_ACP` wins
over shared, and `" U1ABC , ,U2DEF "` parses to `["U1ABC","U2DEF"]`; (iii) a non-conforming ID is
**accepted** and surfaces in attribution as non-conforming, with the ID value absent from the
attribution record; (iv) duplicates collapse, an over-cap input returns a typed error containing
no ID value, and all-empty input returns `ErrNoAuthorizedUsers`.
Named test functions: 4. Test package: `credentials_test` (uses the fake).

**D3 — `Resolved` aggregate and its redaction** · size `M` · complexity `high`
Files: `internal/credentials/resolved.go`.
```go
// Resolved is the Slack credential set. Additional secret-bearing subsystems
// (P13 IPC) model their own types over internal/secret rather than extending
// this one.
//
// authorizedUserIDs is BOXED behind a pointer, for the same verified reason
// Secret is: making a field unexported does NOT protect it from fmt. It does
// the opposite — CanInterface() is false for an unexported field, so
// printValue SKIPS handleMethods and falls through to raw reflective
// printing. A plain unexported []string therefore leaks verbatim under %p,
// under %w on a non-error, and whenever a Resolved is embedded in another
// struct's unexported field. Boxing renders an address instead. Do not
// unbox this field.
type Resolved struct {
	AppToken          secret.Secret
	BotToken          secret.Secret
	TeamID            string
	authorizedUserIDs acl // opaque wrapper over *[]string, with its own total Format
	Attribution       Attribution
}
func (r Resolved) AuthorizedUserIDs() []string // returns a defensive copy
```
Value receivers on all rendering methods. `MarshalJSON` marshals an explicit unexported **shadow
struct** — never the receiver — so it cannot recurse infinitely (GO-11), emitting
`authorized_user_ids_count` and no ID array. `TeamID` renders in the clear (Decision R6 —
an identifier, visible in the oracle too), asserted deliberately so the divergence is intentional.
Acceptance: (i) **the same row set as S3** — `%v %+v %#v %s %q %p %T`, `fmt.Errorf("%w", …)`,
`errors.Join`, a `Resolved` embedded in another struct's **unexported** field, JSON, gob, and both
slog handlers — asserted to contain neither a token canary nor a member-ID canary, for both
`Resolved` and `*Resolved` (this row set is what rev 2 omitted, and the omission was a verified
leak); (ii) `json.Marshal` output contains `authorized_user_ids_count` and no ID array, and
`Resolved` and `*Resolved` produce identical output; (iii) `AuthorizedUserIDs()` returns a copy —
mutating the result does not affect the receiver, and mutating the slice passed to the constructor
afterwards does not affect the receiver either; (iv) a mutation guard asserts that unboxing `acl`
reintroduces the `%p` leak, so the boxing cannot be silently removed.
Named test functions: 3. Test package: `credentials`.

**D4 — `Resolve` orchestration** · size `M` · complexity `medium`
Files: `internal/credentials/credentials.go`.
Sources are passed as a **named struct**, not two positionally-identical interface arguments: rev
2's `NewResolver(env, keychain Source, …)` made a transposed call compile cleanly and silently
invert the whole chain (GO-20, ARCH-13). There is **no `Option` variadic** — rev 2 declared one
with no concrete option, which was speculative and made the exported surface uncountable
(SCOPE-5):
```go
type Sources struct{ Env, Keychain Source }
func NewResolver(src Sources) (*Resolver, error) // error if either field is nil
func DefaultSources() Sources                    // composition root only
func (r *Resolver) Resolve(ctx context.Context) (*Resolved, error)
```
Nil semantics, specified explicitly (SCOPE-11): a nil `Source` field is a construction error; a
nil `ctx` is replaced with `context.Background()`; `Resolve` returns a nil `*Resolved` on any
error and never a partially-populated one; `Resolve` is single-use per resolver (SEC-14) and
returns an error on a second call. Per-field semantics: `app_token`/`bot_token` required (R5);
`team_id` optional, never failing, defaulting to `""` (R6); `authorized_user_ids` via D2 (R7).
Acceptance: (i) `NewResolver(Sources{Env: nil, Keychain: kc})` and the mirror case both return an
error and a nil resolver, and a transposition is impossible because the fields are named;
(ii) a missing required credential returns the R10 error **and** a nil `*Resolved`;
(iii) a missing `team_id` yields `""` with no error while the other fields resolve;
(iv) mixed-tier resolution succeeds, `SourceCoherence()` reports the distinct tiers, and a second
`Resolve` call returns an error.
Named test functions: 4. Test package: `credentials_test`.

### Sub-epic E — Verification, integration, documentation

**E1 — Precedence matrix** · size `M` · complexity `medium`
Files: `internal/credentials/resolve_test.go`.
One table over the four-tier matrix — exactly what the oracle never tested (F8): each of the four
tiers winning in isolation; `_ACP` env beating the shared **keychain**; empty-at-every-tier;
`Unavailable` at tiers 1 and 3 still falling through to env. Uses injected sources, no `t.Setenv`,
so it is parallel-safe.
Acceptance: (i) each of the four tiers is proved to win in isolation; (ii) short-circuit proved by
call log on every row; (iii) an `Unavailable` row never yields a value and still reaches tier 4;
(iv) `t.Parallel()` is enabled and the test passes under `-race`.

**E2 — Keychain classification and timeout** · size `M` · complexity `medium`
Files: `internal/credentials/source_keychain_test.go`.
Drives the C3 table plus the C4 timeout/cancellation paths through an injected `get`.
Acceptance: (i) all five C3 rows plus the wrapped-sentinel row; (ii) timeout and cancellation
rows yield `Unavailable` with the right `Reason`; (iii) no live keychain call — asserted by the
fact that `get` is always injected, with a comment naming PI-5; (iv) backend error text never
reaches `Reason` (type-enforced) and never reaches the rendered error, asserted with a canary
embedded in the stub error.

**E3 — Boundary guards** · size `M` · complexity `high`
Files: `internal/config/boundary_test.go`, `internal/credentials/import_guard_test.go`.
Two guards, each placed in the package that owns its invariant (ARCH-7):
1. **Secret-free config** (in `internal/config`, so it needs no import of `credentials`): a
   reflection type-walk over `config.Config` rejecting any field whose **type name** is `Secret`
   or whose name/`toml:` tag matches a credential vocabulary — `app_token`, `bot_token`,
   `team_id`, `authorized_user_ids`, plus synonyms `password`, `secret`, `api_key`, `oauth_token`,
   `signing_secret`, `credential`. Matching normalises case and underscores, because Go field
   names are `AppToken`, not `app_token` (GO-9 — rev 1's literal patterns would have matched
   nothing). The walk carries a `visited map[reflect.Type]bool` for cycle protection, iterates all
   `NumField()` including unexported and embedded fields, and recurses into pointer, slice, array,
   map **key and value**, failing loudly on an interface-typed field.
2. **Dependency and test-support containment** (in `internal/credentials`): a `go/packages`-based
   test (type-aware, not string-matching) asserting that `github.com/zalando/go-keyring` resolves
   as an import in exactly one file, `source_keychain.go`; that no selector resolves to its `Set`,
   `Delete`, or `DeleteAll` objects; that dot-imports of it are rejected; and that every importer
   of `internal/credentials/credtest` is a `_test.go` file or `tests/integration` (SEC-16). Using
   `go/types` object resolution rather than syntax defeats aliasing and function-value assignment
   (`write := keyring.Set`) that a selector grep would miss (SEC-17). This replaces rev 1's false
   claim that the write surface was unreachable "by construction" (SEC-8).
Acceptance: (i) adding a `secret.Secret` field to `config.Config` fails guard 1, proved by a
`go:build mutation` sub-test rather than a recorded comment; (ii) adding a `team_id` field fails
guard 1 (rev 1 omitted `team_id`); (iii) a self-referential type does not hang the walk;
(iv) four guard-2 mutation sub-tests fail as expected — a second importer, an aliased import, a
dot import, and `write := keyring.Set`.
Named test functions: 2 (one per guard), each with mutation sub-tests. Test packages: `config`
and `credentials_test`.

**E4 — Startup composition helper** · size `M` · complexity `medium`
Files: `cmd/intercom/startup.go`, `cmd/intercom/startup_test.go`.
```go
// loadStartup performs the oracle's startup order (main.rs:130-132): config
// load, validation, then credential resolution, exactly once.
func loadStartup(ctx context.Context, path string, src credentials.Sources) (*config.Config, *credentials.Resolved, error)
```
The test asserts against this function, not a re-implementation. `RunE` wiring is E8.
Acceptance: (i) an ordered call trace proves `Validate` precedes the first credential lookup;
(ii) an invalid config short-circuits with an **empty** fake call log; (iii) credential resolution
occurs exactly once, asserted by call count; (iv) a nil `ctx` and a zero `Sources` each return an
error rather than panicking.
Named test functions: 4.

**E8 — Wire startup validation into `RunE`** · size `S` · complexity `medium`
Files: `cmd/intercom/main.go`, `cmd/intercom/main_test.go`.
Resolves SCOPE-1/SCOPE-11/ARCH-6/ARCH-14 outright. Rev 2 shipped `loadStartup` uncalled, which
three personas judged to be relocated dead code rather than integration. `RunE` now:
1. emits the existing single `intercom starting` Info line **first** (preserving the
   `TestRootCmd_LogLevelDebugEmitsJSON` contract, which parses stderr as one JSON object);
2. calls `loadStartup(cmd.Context(), configPath, credentials.DefaultSources())`;
3. on success, logs the returned `Attribution` via its `LogValue` — the composition root doing the
   logging the library refuses to do (R14) — then returns `errNotImplemented`;
4. on failure, returns that error (it does not log; `main` already logs returned errors).

**Verified safe against the existing suite** before adoption: the Info line still precedes any
failure, so the single-JSON-line test passes; `RunE` never logs at `error` level, so
`TestRootCmd_LogLevelErrorSuppressesInfo` still sees empty stderr; `--help` and the unknown-flag
test never reach `RunE`; `config_flag_test` inspects flag metadata only.
Acceptance: (i) all five pre-existing `cmd/intercom` tests pass **unmodified**; (ii) with a valid
synthetic config and injected env credentials, `RunE` returns `errNotImplemented` and stderr
contains exactly two JSON lines (startup + attribution), neither containing a canary; (iii) with a
missing config, `RunE` returns a `config`-kind error and not `errNotImplemented`; (iv) a new test
asserts the attribution log record names the winning source per field and no value.
Named test functions: 3.

**E5 — Credential reference documentation** · size `M` · complexity `low`
Files: `docs/credentials-reference.md`.
Per-field table — four rows × four tier columns — **not** a single prose chain, with the keychain
cells for `authorized_user_ids` explicitly marked *n/a* (DOC-1). Must state:
* that `authorized_user_ids` never consults the keychain, resolves through exactly two env tiers,
  and that no `slack_member_ids` keychain account exists or is read (DOC-1);
* that keychain-first holds **within** a namespace tier, **not globally** — with a worked example
  showing `SLACK_APP_TOKEN_ACP` (env) beating an `agent-intercom` keychain entry, flagged as a
  divergence from the oracle's documented global rule (DOC-2);
* all eight env vars including `SLACK_TEAM_ID_ACP`, written down for the first time anywhere (F4);
* the three keychain accounts and both service names, and that the `agent-intercom` prefix is
  deliberately retained from the Rust implementation and must not be renamed to match the binary
  (DOC-11);
* required vs optional semantics and unavailable-keychain behavior;
* copy-pasteable provisioning commands per backend — macOS
  `security add-generic-password -s agent-intercom-acp -a slack_app_token -w '<SECRET-VALUE>'`,
  Linux `secret-tool store` with the exact attribute names `go-keyring` queries, Windows the exact
  Credential Manager target-name form — each row carrying an explicit **status token**
  (`verified` | `recipe-unverified`). A `recipe-unverified` row must state that keychain tiers 1
  and 3 still resolve normally on that platform and that only the *provisioning command* is
  unverified, so the wording cannot be read as "the keychain does not work here" (DOC-15). The
  closure document records, per backend, the `keyring.Get` round-trip result or the reason
  verification was impossible from the build host;
* the diagnostic path: attribution is returned as data and logged by the composition root at Info;
  `--log-level=error` suppresses it (DOC-8);
* a **status note** stating that credential resolution runs at startup and validates, but the
  server itself is not implemented (`RunE` returns not-implemented after validating) — the
  successor to rev 2's "not yet wired" note, now accurate because E8 wires it (DOC-4, DOC-16).
Placeholders come from a closed vocabulary only: `xoxb-EXAMPLE-NOT-A-REAL-TOKEN`,
`xapp-1-EXAMPLE-NOT-A-REAL-TOKEN`, `T0123456789`, `C0123456789`, `U0123456789,U9876543210`, and
`<SECRET-VALUE>` for provisioning-command secret arguments — the last two added because rev 2's
vocabulary omitted the channel-ID form already used in `config.toml.example` and the placeholder
its own mandated macOS command required, making its acceptance criterion self-contradictory
(DOC-17).
Acceptance: (i) the per-field table marks `authorized_user_ids` keychain cells *n/a* and states
the env-only rule; (ii) the worked precedence counter-example and the migration cross-link are
present; (iii) every backend row carries a status token and every `recipe-unverified` row carries
the tiers-still-work clause; (iv) markdownlint passes, the E7 bidirectional coupling test passes,
and only closed-vocabulary placeholders appear.

**E6 — Operator surfaces** · size `M` · complexity `low`
Files: `README.md`, `config.toml.example`, `docs/config-reference.md`.
`config.toml.example` gains a `[slack]` body comment stating no credential is ever read from the
file, plus a short header block naming the **two service names only** and pointing to
`docs/credentials-reference.md` — it deliberately does **not** re-enumerate the accounts and eight
env vars, because duplicating the canonical table across surfaces is the drift hazard DOC-13
identified and SCOPE-6 objected to (DOC-7, DOC-13). `README.md` gains a short credential section
pointing at the canonical doc, a `--config`/`--log-level` flag table, and a single consolidated
**Status** section covering both the P2 config note and the new credential status, so there are
not two independent "not yet wired" notes with no owner (DOC-16, DOC-19).
`docs/config-reference.md` — omitted from rev 1's file list — gains `authorized_user_ids` in its
`[slack]` secret-field note (it currently lists three where V6 lists four), a link to the new
reference, the migration bullet, and a rewrite of its forward-tense migration sentence into the
present tense, which becomes stale the moment P3 merges (DOC-3, DOC-6, DOC-18):
> `_ACP`-namespaced credentials are now checked first, unconditionally. If you ran the Rust
> server in default (MCP) mode with both a shared and an `_ACP` identity provisioned,
> intercom-go resolves the `_ACP` identity. Unset the `_ACP` variables or remove the
> `agent-intercom-acp` keychain entries to preserve the previous behavior.

Acceptance: (i) all three files updated and cross-linked, and the config-reference migration
sentence is present-tense; (ii) the `[slack]` note lists all four credential-bearing fields;
(iii) the README flag table lists exactly the flags registered in `newRootCmd` with their real
defaults and notes the unrecognized-value→Info fallback, asserted by the E7 coupling test against
`cmd/intercom` (DOC-19); (iv) markdownlint passes and the E7 scan finds no token shape.

**E7 — Doc/identifier coupling and secret scan (as Go tests)** · size `M` · complexity `medium`
Files: `internal/credentials/docs_coupling_test.go`, `internal/secret/scan_test.go`.
**Adjudication (SEC-10/DOC-9 vs SCOPE-13/SCOPE-15).** Security and Schema-CLI-Docs required a
mechanised scan and a doc/error binding; Scope objected that a `scripts/*.ps1` plus a new CI gate
crosses the excluded "CI hardening follow-ups" boundary. Rev 3 keeps the controls but implements
them as **ordinary Go tests** — no new script, no workflow edit, no new CI gate; they run inside
the existing `go test ./...` step. That satisfies both.

The coupling test is **bidirectional and code-constant-driven** (DOC-14), not error-text-driven:
it asserts every element of `{ServiceACP, ServiceShared, EnvSuffixACP, the three keychain
accounts, the four env base names}` appears verbatim in `docs/credentials-reference.md`, **and**
that the doc contains no `SLACK_*` or `agent-intercom*` identifier absent from that set. This
covers the `team_id` path, which produces no error text at all and was therefore invisible to rev
2's error-driven check — including `SLACK_TEAM_ID_ACP`, the newest and least-reviewed identifier.
It also asserts every identifier enumerated in `README.md` and `config.toml.example` appears in
the canonical doc (DOC-13), and that the oracle's `set X or X` duplication wart (F5) has not
regressed.
The scan test walks `git ls-files`, **fails separately if `.env.local` is tracked**, then scans
every other tracked file — including `.env.example`-style templates, which rev 2's blanket
`:!*.env*` exclusion wrongly skipped (SEC-18) — matching `xox[baprsce]-[A-Za-z0-9-]{10,}` and
`xapp-[0-9]-[A-Za-z0-9-]{10,}`. Allowlisting is **exact full-token equality** against the closed
placeholder vocabulary, never substring containment of `EXAMPLE`/`CANARY`. Match locations are
reported without printing matched values. It never opens `.env.local` because that path is
excluded from the walk and asserted untracked.
Acceptance: (i) every code constant is found in the reference doc and the doc introduces no
unknown identifier — both directions asserted; (ii) mutation sub-tests: removing an env var from
the doc fails, and adding an undeclared `SLACK_*` string to the doc fails; (iii) a planted
synthetic token in a tracked file fails the scan, and a near-miss token *containing* `CANARY` plus
extra secret material also fails (proving the allowlist is not a substring bypass); (iv) a test
asserts `.env.local` is untracked and that the walked file list contains no path matching
`*.env.local`.
Named test functions: 4. Test packages: `credentials_test`, `secret_test`.

## Dependency Graph

Rev 1's graph carried three false edges (ARCH-10). Corrected:

```text
S1 ──> S2 ──> S3
        │
        ├──────────────> B1 ──> B2 ──> B3
        │                 │      │
        │                 └──> C1 ──> C2 ──┐
        │                        │    C3 ──> C4
        │                        │    C5 ──┤
        │                        └─────────┴──> D1 ──> D3 ──> D4
        │                                        │      │
        │                              B1,C1,C2 ─┴─> D2 ┘
        │
        └──> E3(guard 1)

D1 ──> E1        C3,C4 ──> E2        C3,C5 ──> E3(guard 2)
D4,C5 ──> E4 ──> E8     D1,D2 ──> E5 ──> E6        D1,E5,E6 ──> E7
```

`D2` depends on `B1`/`C1`/`C2` — **not** on `D1` — because correction F2 makes it an independent
env-only path; rev 1's `D1 → D2` edge re-coupled exactly what F2 separated. `B1 → B2 → B3` holds
because the error text is derived from the attempt vocabulary (ARCH-9). `E7` depends on `D1` (code
constants) and on `E5`/`E6` (the surfaces it binds), not on `B3`, since rev 3 made the coupling
check constant-driven rather than error-text-driven (DOC-14). The graph *permits* reordering among
independent leaves; **policy** (P-016) forbids parallelism regardless.

Serial execution order: **S1 → S2 → S3 → B1 → B2 → B3 → C1 → C2 → C3 → C4 → C5 → D1 → D2 → D3 →
D4 → E1 → E2 → E3 → E4 → E8 → E5 → E6 → E7**. One implementation branch, one worktree.
Twenty-three units × 2 hours ≈ 46 hours of human-equivalent effort.

**Shipment sizing (ARCH-16, SCOPE-16).** Architecture proposed splitting at the `internal/secret`
seam; Scope argued the whole slice is too large. Adjudication: keep **one shipment**, because a
split would put a redaction primitive with zero consumers through a full ship cycle and then
require a second cycle to prove it is actually used — and the primitive's value is only
demonstrable through `Resolved` and the leak matrix that consumes it. Instead, adopt
Architecture's fallback: an explicit **mid-slice gate checkpoint after S3**, running gates 1–5 and
the full leak matrix before any credential code is written, so the highest-security-density units
get an isolated verification point without a second shipment. This is recorded as a required
checkpoint in the execution order, not merely advice.

## Decisions and Rationale

| ID | Decision | Rationale |
|---|---|---|
| R1 | Static four-tier ACP chain, no mode parameter | ACP is the surviving runtime after MCP-mode retirement. Strict superset of *reachability* — but **precedence changes** for operators who provisioned both namespaces, so E6 carries a migration bullet (DOC-3). Deliberation P3-D1. |
| R2 | Provider seam plus a real adapter plus a reusable fake | The oracle has no seam, so its precedence is untested and its tests invert on any provisioned machine (F8). Deliberation P3-D2. |
| R3 | `zalando/go-keyring` v0.2.8 | Verified `CGO_ENABLED=0` on all four targets, `govulncheck` clean, MIT/MIT/BSD-2-Clause. `keybase/go-keychain` needs cgo on darwin. Deliberation P3-D2/F13. |
| R4 | No bespoke platform adapters | The dependency covers all four targets plus a distinct unsupported-platform error. Deliberation P3-D3. |
| R5 | Classify `Miss` vs `Unavailable`; both continue the chain | Keeps outcomes byte-for-byte oracle-equivalent while fixing the F6 diagnostic collapse. Deliberation P3-D4. |
| R6 | `AppToken`/`BotToken` are `Secret`; `TeamID` plain; ACL unexported and count-redacted | Matches the oracle's own choices for tokens and `team_id`, and closes its `GlobalConfig` ACL leak. Deliberation P3-D6. |
| R7 | Pointer-boxed `Secret` with a total `fmt.Formatter` | **Verified**: `Format` alone leaks via `%p`, `%w`, and unexported-field traversal. Pointer boxing closes all three. See Review Remediation Record. |
| R8 | `loadStartup` in `cmd/intercom`, **called by `RunE`** | Rev 3 change. Fully satisfies the operator's "startup validation integration" requirement, which three personas flagged as unmet across two revisions. Verified safe against all five existing `cmd/intercom` tests before adoption (E8). |
| R13 | `Secret` lives in `internal/secret`, with **no alias** in `credentials` | Justified on present-tense cohesion: `Secret` has no dependency on credential resolution. An alias would give one type two import paths and re-advertise `credentials` as a source for the primitive, undoing the split (ARCH-12). |
| R9 | `authorized_user_ids` gets its own env-only resolver, with ID-shape validation | Correction F2, hardened per SEC-7. |
| R10 | No unmarshal methods on `Secret`; `GobEncode` fails closed | Makes the secret-free config boundary structural, and closes the `gob` path. |
| R11 | No zeroization | Go strings are immutable and GC-managed; a wipe would be security theatre. The memory-residency boundary is documented instead (SEC-9). |
| R12 | No credential CLI subcommand at P3 | Deferred explicitly to the phase that wires `RunE`; recorded as a decision rather than a risk-table aside (DOC-8). |
| R13 | `Secret` lives in `internal/secret`, not `internal/credentials` | Superseded by the R13 row above (rev 3 removed the alias). |
| R14 | The library never logs; `Attribution` is the sole diagnostic egress, logged by `RunE` | `internal/config` already pins this precedent ("Library code never logs directly (Principle V)", `config.go:124`). Two contradictory diagnostic protocols at phase 3 of 15 would leave every later phase without a tiebreaker, on the package the plan itself calls the highest-risk leak vector (ARCH-1). Rev 3's E8 gives the returned attribution a real consumer. |
| R15 | Precedence poisoning is an **accepted residual risk** for P3 | Security (SEC-5) correctly notes documentation is not a control. But every available control — a source-trust policy, operator-configurable namespaces, or fail-closed conflict detection — either resurrects the mode-switching P2-D5 retired, or abandons the short-circuit that makes outcomes oracle-equivalent (PI-6). P3 therefore accepts the risk explicitly: the trust model is documented (E5), anomalies are surfaced via `SourceCoherence()`, and the enforcement decision is assigned to the phase that owns a startup policy surface. Recorded as an acceptance, not a fix. |

## Risks and Caveats

| Risk | Severity | Mitigation | Unit |
|---|---|---|---|
| New dependency breaks the CGO-free matrix | High | Measured on all four targets; CI cross-compile re-proves each run | C3 |
| Secret leaks through an unconsidered render path | High | Pointer boxing + total `Format` + marshal/log/gob methods; leak matrix covering the three verified-leaking paths and two recorded negative controls | S1–S3 |
| Secret reaches an error string | High | Errors carry source descriptors derived from attempts; canary-absence asserted on every constructed error | B3, E2 |
| Backend error text echoes the account or daemon data | Medium | `Reason` is a closed `uint8` enum — assigning `err.Error()` does not compile | B1, C3 |
| Blocking keychain call hangs startup | Medium | Goroutine + buffered channel + select with a 5s default; timeout classified `Unavailable` | C4 |
| Precedence poisoning: a same-user process writes a keychain entry that beats a supervisor's env var | Medium | **Accepted residual risk** (R15): documented as an explicit trust property (E5); `SourceCoherence()` surfaces the anomalous case; enforcement assigned to the phase that owns a startup policy surface | B2, E5 |
| Cross-tier mixing pairs credentials from different Slack apps or tenants | Medium | Oracle behavior preserved deliberately, but made visible via `SourceCoherence()` across **all three** chain-resolved credentials and both mixing modes; a blocking tenant-identity check is required of P8 before the tokens are used (SEC-6) | B2, D4 |
| A test touches a real keychain | Medium | `get` is an injected field, never a package var; the fake lives in `credtest`; the import guard pins the dependency to one file | C3, C5, E3 |
| P2 secret-free boundary erodes | Medium | Guard 1 lives in `internal/config` and matches normalised names and `toml:` tags, with cycle protection | E3 |
| Operators cannot actually provision a readable keychain entry | Medium | Per-backend commands verified by a manual round-trip recorded in closure; unverified backends explicitly marked | E5 |
| Migrating dual-identity operators silently switch Slack identity | Medium | Migration bullet in `docs/config-reference.md`; recorded in the closure divergence list | E6 |
| Headless CI has no keychain backend | Medium | `Unavailable` is first-class and continues the chain; env-only is a tested path | C3, E1 |
| Chain choice (ACP) proves wrong | Low | Four-tier ACP is a reachability superset; shared credentials still resolve at tiers 3–4 | D1 |
| Scope creep into P4+ | Low | Explicit exclusion table; no models, persistence, Slack, ACP, or CLI subcommands | — |

## Plan Hardening Signals (REQUIRED)

| Signal | Present | Evidence |
|---|---|---|
| Security-sensitive surface | **Yes** | The slice handles live Slack credentials; leak prevention is the primary acceptance property. |
| New third-party dependency | **Yes** | `go-keyring` v0.2.8 plus three transitive modules, on a security-critical path. |
| Cross-platform / OS-integration boundary | **Yes** | Windows Credential Manager, macOS `/usr/bin/security`, Linux D-Bus, plus an unsupported-platform path. |
| Behavioral divergence from the oracle | **Yes** | Error classification (R5), ACL redaction (R6), pointer-boxed total redaction (R7), static chain (R1), ID validation (R9), no library logging (R14). |
| Corrections to prior recorded analysis | **Yes** | Six stash claims corrected (F2, F3, F6, F7, F8, F9), plus three rev-1 plan defects empirically disproved. |
| **Concurrency surface** | **Yes** (changed in rev 2) | C4 introduces a goroutine + channel handoff; `-race` now exercises a real concurrent path (GO-4). |
| Irreversible or stateful migration | No | No schema, no persisted state; rollback is a single revert. |
| Public API/contract freeze | Partial | `internal/secret` and `internal/credentials` are new; the `apperr` display contract they render through is frozen. |

**Requires plan hardening: yes**

## Plan Hardening

### Instructions and learnings consulted

| Source | Constraint absorbed |
|---|---|
| `constitution.instructions.md:20-23` | **Test-first is mandatory** — every unit carries a red-green anchor. |
| `constitution.instructions.md:34-35` | No secrets or credentials may be committed. |
| `constitution.instructions.md:58` | Traceable output — satisfied by `Attribution` as returned data. |
| `technology-go.instructions.md:14,75,107` | Go 1.22+, `log/slog`, **no global mutable state** — directly drove C3's injected `get` field over a package var. |
| `ci-security.instructions.md:12-40` | Action SHA pinning — no workflow changes in this slice. |
| `coding-discipline.instructions.md:37` | Touch only what you must. |
| Compound: `external-spec-yields-to-workspace-instructions` | Workspace floors outrank the port brief — applied to the dependency choice. |
| Compound: `go-race-requires-cgo-fails-closed` | Verify real tool behavior rather than guarding assumed failure modes — applied by measuring the dependency and by empirically testing the `fmt` claims. |

### Protected invariants

| ID | Invariant | Enforced by |
|---|---|---|
| **PI-1** | A credential value never appears in `fmt`, `slog`, JSON, `gob`, or an error string, **through the standard rendering surfaces**. Reflection/`unsafe` and explicit `Expose()` cross the boundary by design (SEC-3 — the rev-1 guarantee was overstated). | S1 pointer boxing; S2 methods; S3 matrix; B3 canary assertions |
| **PI-2** | `internal/config` never carries a credential value or a `Secret`-typed field. | E3 guard 1, owned by `internal/config` |
| **PI-3** | A non-`Hit` result never carries a value. | C1 constructor-only `Result` (structural) |
| **PI-4** | `Unavailable` never yields a value or a default — it only annotates. | C1 constructors; C3/C4 classification; E1 fallthrough rows |
| **PI-5** | No test performs live OS keychain I/O. | C3 injected `get`; C5 `credtest`; E3 guard 2 |
| **PI-6** | Resolution outcomes remain oracle-equivalent; only diagnostics improve. | E1 matrix mapped 1:1 to R1–R9 |
| **PI-7** | The `apperr` display contract (`config: {msg}`) is unchanged. | B3 renders through `apperr.KindConfig` |
| **PI-8** | `go-keyring`'s write API is imported by exactly one file and never called. **This is a CI-enforced import constraint, not a Go capability guarantee** (SEC-8 — rev 1 claimed the stronger, false form). | E3 guard 2 |
| **PI-9** | Neither reference repository is modified, and `.env.local` is never read — including by the secret-scan gate, which operates on `git ls-files` excluding `*.env*`. | Execution precheck; E7 script construction |

### Risky actions

`strict_safety.enabled: false`, so these are classifications for reviewer attention, not approval
gates. None is `destructive`; there is no migration, deletion, or data mutation.

| ProposedAction | ActionRisk | Approval | Notes |
|---|---|---|---|
| Add `go-keyring` v0.2.8 + 3 transitive modules | **moderate** | Not required | Measured CGO-free, `govulncheck` clean, permissive licenses; re-verified by gates 3, 5, 6. |
| Introduce two new packages (`internal/secret`, `internal/credentials`) | **moderate** | Not required | New contracts, but `internal/`; no existing API changes. |
| Call an OS credential store at runtime | **moderate** | Not required | Read-only, import-guarded (PI-8), never exercised in tests (PI-5). |
| Add `loadStartup` to `cmd/intercom` | **low** | Not required | Unexported, uncalled by `RunE`; existing tests unmodified. |
| Add one test to `internal/config` | **low** | Not required | Test-only; no production change to a frozen package. |
| Add a secret-scan and doc-coupling **Go test** | **low** | Not required | Read-only over tracked files; no CI script or workflow change. |
| Documentation updates (4 files) | **low** | Not required | Closed placeholder vocabulary only. |

### Test-first execution posture

| Unit | Red-green anchor |
|---|---|
| S1 | Zero-value and `Equal` assertions before `secret.go` exists |
| S2 | The `%q`-quoting and `gob`-error rows fail against S1-only code |
| S3 | The `%p`/`%w`/unexported-field rows fail if the pointer indirection is removed |
| B1–B3 | `Outcome(0) == Miss`, `Reason.String()` totality, and the dual-`errors.Is` assertion precede the constructors |
| C1–C5 | `ConformanceTest` precedes every source implementation |
| D1–D4 | Chain-order, short-circuit, and nil-construction assertions precede resolver code |
| E1–E4, E7 | These units *are* tests; their red state is the absence of the behavior asserted |
| E5, E6 | markdownlint, the E7 coupling test, and the secret-scan gate stand in for red-green |

### Environment prechecks

1. Confirm `git rev-parse HEAD` in `C:\Source\GitHub\intercom` is still `41df772`, then leave it alone.
2. Confirm `.env.local` is never opened, staged, or read, in either repository.
3. Confirm the locally modified `.gitignore`, `.claude/`, and `.backlogit/hooks_queue.jsonl` remain unstaged.
4. Confirm exactly one implementation worktree exists (P-016).
5. Confirm `go env GOTOOLCHAIN` resolves to the pinned `go1.26.5`.

### Rollback triggers and procedure

Roll back if: any canary or real-shaped credential appears in CI logs or artifacts; the
cross-compile matrix fails on any target; `govulncheck` reports an affecting vulnerability in the
new dependency tree; or the C4 timeout is observed firing routinely on a healthy platform
(indicating a mis-tuned budget rather than a broken backend).

Procedure: `git revert` the single feature merge. The slice adds two packages, one uncalled
function in `cmd/intercom`, one test in `internal/config`, one script, one dependency, and
documentation. It changes no existing package **behavior**, so the revert is complete and
self-contained. No persisted state, schema, migration, or deployed configuration to unwind, and
zero importers on merge.

### Monitoring signals for the closure record

* Count of `Unavailable` keychain outcomes per startup, by `Reason` — distinguishes a locked or
  absent backend from a missing credential, the distinction the oracle cannot make.
* `SourceCoherence()` occurrences reporting more than one distinct tier — legal but worth
  surfacing (SEC-6).
* Per-field winning tier — detects the DOC-3 identity-switch case.
* Terminal not-found errors by credential name.

### Review-gate capability risks (P-012)

Adversarial review across five personas: Security, Go/Correctness, Architecture, Scope,
Schema-CLI-Docs. The gate must emit literal `decision:` and `dispatch_mode:` markers. If any
persona is unavailable, the gate declares `dispatch_mode: degraded` naming the missing persona; a
degraded **security** review on a credential-handling slice is blocking, not advisory. All P0/P1
findings must be remediated and re-reviewed until `decision: PASS`, or the session halts.

### Unresolved operator decisions

None block execution. Carried forward: whether the oracle's keychain path is live in production
(F7 → `8ACF7110`); whether to keep `SLACK_TEAM_ID_ACP` (kept, documented); the source-trust policy
for precedence poisoning (SEC-5 → the phase that wires startup); and the blocking tenant-identity
check required of P8 before tokens are used (SEC-6).

## Review Remediation Record

Rev 1 was reviewed by five adversarial personas; **all five returned `decision: FAIL`**
(`dispatch_mode: full` — no persona was unavailable). Twenty P1 findings and no P0 findings were
raised. Three of the most severe were empirically verified rather than accepted on argument:

| Verified claim | Method | Result |
|---|---|---|
| `Format` alone does not redact every verb | Go program on `go1.26.5`: `%w` via `fmt.Errorf`, `%p`, and a `Secret` in an unexported field | **All three printed the plaintext canary.** `%v`/`%d` were safe. Rev 1's PI-1 was false. |
| Pointer boxing closes them | Same program with `v *string` | **All three safe** — rendered as an address. `slog` `TextHandler` likewise flipped from leaking to safe. |
| B1's dual-`errors.Is` criterion was implementable | Go test against the real `internal/apperr` | **Refuted.** `errors.Is(err, ErrCredentialNotFound)` returned `false`; `Wrap` destroyed the message. The `Unwrap() []error` composite returns `true` for both and preserves the message. |

Findings and dispositions:

| Finding | Severity | Disposition |
|---|---|---|
| SEC-1 / GO-6 / ARCH-4 — `Result` and `Reason` unenforceable | P1 | **Fixed**: unexported fields, three constructors, closed `Reason` enum, `Outcome(0) == Miss` (B1, C1) |
| SEC-2 / GO-5 — receivers unspecified | P1 | **Fixed**: value receivers pinned, compile-time assertion block (S2) |
| SEC-3 — redaction guarantee overstated | P1 | **Fixed**: PI-1 narrowed to standard surfaces; ACL field unexported; `GobEncode` fails closed |
| SEC-4 / GO-4 — context does not bound a blocking call | P1 | **Fixed**: goroutine + buffered channel + select (C4); hardening signal flipped to concurrency-present |
| SEC-5 — precedence poisoning | P1 | **Partially deferred, recorded**: documented as a trust property, `MixedNamespace` surfaces it; source-trust policy assigned to the startup-wiring phase |
| SEC-6 — cross-tier identity mixing | P1 | **Fixed as far as P3 can**: `MixedNamespace` flag; blocking tenant check required of P8 |
| GO-1 / GO-2 — `%p`/`%w`/unexported-field leaks | P1 | **Fixed**: pointer boxing (S1), matrix rows added (S3) — verified |
| GO-3 — dual-`errors.Is` unimplementable | P1 | **Fixed**: `Unwrap() []error` composite (B3) — verified |
| ARCH-1 — library logging contradicts Principle V | P1 | **Fixed**: all library logging removed; `Attribution.LogValue()` is the sole egress (R14, B2, D1) |
| SCOPE-1 / ARCH-6 — startup integration shortfall | P1 | **Fixed**: `loadStartup` lands as real production code in `cmd/intercom` (E4) |
| SCOPE-7 — non-machine-verifiable criteria | P1 | **Fixed**: every "a comment records…" / "go doc states…" criterion replaced with an executable assertion |
| SCOPE-11 — nil handling, CLI docs missing | P1 | **Fixed**: nil semantics specified (D4); flag table and diagnostic path documented (E6, E5) |
| DOC-1 — env-only exception not required by docs | P1 | **Fixed**: per-field table with *n/a* cells and an explicit statement (E5) |
| DOC-2 — oracle's global keychain-first rule now false | P1 | **Fixed**: worked counter-example required (E5) |
| DOC-3 — dual-identity migration silently switches identity | P1 | **Fixed**: R1 rationale softened; migration bullet added (E6) |
| DOC-4 — docs describe a contract the binary never runs | P1 | **Fixed**: wiring-status note required on every operator surface (E5, E6) |
| DOC-5 — provisioning guidance insufficient | P1 | **Fixed**: per-backend verified commands; exclusion reworded to "in product code" |
| SEC-7 — ACL parser unvalidated | P2 | **Fixed**: shape validation, dedupe, caps, unexported field with defensive copy (D2, D3) |
| SEC-8 — PI-8 false guarantee | P2 | **Fixed**: reworded; CI import guard added (E3 guard 2) |
| SEC-10 / DOC-10 — secret-scan gate unmechanised | P2 | **Fixed**: pinned script over `git ls-files` excluding `*.env*` (E7) |
| GO-7 — sentinel matching | P2 | **Fixed**: `errors.Is`, wrapped-sentinel row (C3, E2) |
| GO-8 / ARCH-5 — `NewResolver()` defaults wrong either way | P2 | **Fixed**: required positional sources, `(*Resolver, error)` (D4) |
| GO-9 / ARCH-7 / SCOPE-2 — boundary guard misplaced and would match nothing | P2 | **Fixed**: moved to `internal/config`, normalised matching, `toml:` tags, cycle protection, `team_id` added (E3) |
| GO-10 / ARCH-8 — stub as package var | P2 | **Fixed**: injected `get` field (C3) |
| GO-11 — `MarshalJSON` recursion | P2 | **Fixed**: explicit shadow struct (D3) |
| GO-12 — `%q` unquoted | P2 | **Fixed**: verb-aware `%q` (S2) |
| GO-13 / SCOPE-10 — matrix gaps, fake unreachable from integration test | P2 | **Fixed**: matrix extended; fake moved to `credtest` (S3, C5) |
| ARCH-2 — `Secret` in the wrong package | P2 | **Fixed**: extracted to `internal/secret` with a type alias (R13, S1–S3) |
| ARCH-3 — `Ref`/`Source.Kind()` leaky union | P2 | **Fixed**: `step{src, ref}` pairing; `Kind()` removed (C1, D1) |
| ARCH-9 — error text duplicates chain knowledge | P2 | **Fixed**: message derived from `[]Attempt` (B3) |
| ARCH-10 — false graph edges | P3 | **Fixed**: graph redrawn; policy separated from dependency |
| ARCH-11 / DOC-11 — Slack-shaped aggregate, retained service prefix | P3 | **Fixed by documentation**: `Resolved` doc names it the Slack set; E5 pins the prefix rationale |
| SEC-9 — memory residency | P2 | **Recorded**: R11 documents the boundary; dump/child-env controls noted for the deployment phase |
| SEC-12 — attribution as an existence oracle | P2 | **Recorded**: attribution is operator-only data, never a public response; the composition root chooses the level |
| SEC-13 — no explicit threat model | P2 | **Fixed**: the three named threats appear in the risk table with prevention and detection |
| SCOPE-3 / SCOPE-4 — attribution and `Resolved` rendering overbuilt | P2 | **Partially accepted**: `Attribution` is retained because R14 makes it the *only* diagnostic egress once logging is removed; `Resolved`'s `String`/`GoString` are dropped, keeping `Format`, `MarshalJSON`, `LogValue` |
| SCOPE-6 — three doc locations | P2 | **Accepted with structure**: `docs/credentials-reference.md` is the single canonical source; the other three link to it |
| SCOPE-8 / SCOPE-9 — unit sizing | P2 | **Fixed**: re-split into 22 units; test-function counts pinned per unit; table rows declared as one scenario group |

## Runtime Verification and Closure

### Rev 3 remediation and reviewer adjudications

Attempt 2 returned **Architecture: PASS, Schema-CLI-Docs: PASS**; Security, Go, and Scope
returned FAIL. Rev 3 dispositions:

| Finding | Sev | Disposition |
|---|---|---|
| GO-14 — `Resolved.authorizedUserIDs` leaks via `%p`, `%w`, and unexported-field embedding; the rev-2 comment stated the inverse of how `fmt` treats unexported fields | P1 | **Fixed, empirically verified.** A Go program on `go1.26.5` printed the member-ID canary through all three paths with a plain unexported slice, and rendered an address through all three once boxed. D3 now boxes the ACL behind an opaque `acl` type with its own total `Format`, adopts S3's full row set, and adds an unboxing mutation guard. |
| SEC-5 — precedence poisoning | P1 | **Accepted residual risk, recorded as Decision R15.** Every available control resurrects retired mode-switching or abandons the oracle-equivalent short-circuit. Documented trust model + `SourceCoherence()` + enforcement assigned to the startup-policy phase. Security's dissent is recorded rather than papered over. |
| SEC-6 — cross-tier identity mixing | P1 | **Fixed.** `MixedNamespace` (two tokens, namespace only) replaced by `SourceCoherence()` over all three chain-resolved credentials, detecting keychain-vs-env mixing within a namespace as well. |
| SCOPE-1 / SCOPE-11 — startup integration | P1 | **Fixed outright.** New unit E8 wires `loadStartup` into `RunE`. Verified safe against all five existing `cmd/intercom` tests before adoption. |
| SCOPE-7 — non-machine-verifiable criteria | P1 | **Fixed.** Every "recorded as a comment", "documented in-file", "enforced by the type", and "asserted by construction" criterion replaced with an executable assertion or a `go:build mutation` sub-test. |
| SEC-14 — unbounded timed-out workers | P2 | **Fixed.** Single-use resolver + per-`Ref` in-flight cap; repeated-timeout test asserts the worker count is capped. |
| SEC-16 — `credtest` ships in the module | P2 | **Fixed.** E3 guard 2 proves every `credtest` importer is a `_test.go` file or `tests/integration`. |
| SEC-17 — guard 2 bypassable | P2 | **Fixed.** `go/packages`/`go/types` object resolution instead of syntax matching; mutation sub-tests for alias, dot import, and function-value assignment. |
| SEC-18 — scan blind spots | P2 | **Fixed.** `.env.local` asserted untracked and excluded; all other tracked env templates scanned; `xoxe`/`xoxc` added; allowlist is exact-token equality with a near-miss bypass test. |
| GO-15 — cancellation chain unsatisfiable | P2 | **Fixed.** `ErrResolutionCanceled` wraps `context.Canceled`. |
| GO-16 / ARCH-15 — `credtest` cycle, `ConformanceTest` homeless | P2 | **Fixed.** Helper homed in `credtest`; every unit states its test package. |
| GO-17 / GO-21 / GO-22 — equality semantics, non-executable criteria | P2 | **Fixed.** Pointer-identity semantics asserted; map-key prohibition documented; zero-vs-empty `Equal` pinned. |
| GO-19 — timeout derivation missing | P3 | **Fixed.** `WithTimeout` + `reasonFor` + nil-ctx handling shown. |
| GO-20 / ARCH-13 — transposable `Source` args, two `Kind` types | P2 | **Fixed.** Named `Sources` struct; `SourceKind` rename. |
| GO-23 — gob needs three cases | P3 | **Fixed.** |
| ARCH-12 — dead `Secret` alias | P2 | **Fixed.** Alias dropped; extraction re-justified on present-tense cohesion. |
| ARCH-14 — no forcing function for wiring | P2 | **Moot.** E8 wires it now. |
| ARCH-16 / SCOPE-16 — shipment too large | P2 | **Adjudicated.** One shipment retained (a `secret`-only shipment would ship a primitive with zero consumers), with Architecture's fallback adopted: a mandatory mid-slice gate checkpoint after S3. |
| ARCH-17 — derived predicates stored as fields | P3 | **Fixed.** `AnyUnavailable()`/`SourceCoherence()` are methods. |
| DOC-13 — canonical/duplicate drift | P2 | **Fixed.** README and `config.toml.example` no longer re-enumerate accounts and env vars; they link to the canonical doc, and E7 binds whatever they do enumerate. |
| DOC-14 — coupling test blind to `team_id` | P2 | **Fixed.** Coupling is now constant-driven and bidirectional. |
| DOC-15 — manual round-trip not verifiable | P2 | **Fixed.** Per-backend status tokens with an explicit tiers-still-work clause on unverified rows. |
| DOC-16 — wiring-note staleness | P2 | **Fixed.** E8 wires startup, so the note becomes an accurate consolidated README **Status** section rather than three ownerless "not yet wired" notes. |
| DOC-17 — placeholder vocabulary self-contradictory | P3 | **Fixed.** `<SECRET-VALUE>` and `C0123456789` added. |
| DOC-18 / DOC-19 — stale tense, unbound flag table | P3 | **Fixed.** |

**Adjudicated reviewer conflicts.** Scope's re-review demanded removal of controls that Security
and Schema-CLI-Docs had specifically required. These were resolved on the merits, not by deferring
to the most recent reviewer:

| Conflict | Adjudication |
|---|---|
| SCOPE-12 (drop `internal/secret`; it serves excluded P13) vs ARCH-2 (extract it) | **Extraction kept**, re-justified on present-tense cohesion alone; the alias — the part that genuinely served only future callers — dropped. |
| SCOPE-13 (secret scan is excluded CI hardening) vs SEC-10 (mechanise the gate) | **Both satisfied**: implemented as an ordinary Go test inside `go test ./...`, with no script, workflow edit, or new CI gate. |
| SCOPE-14/15 (drop the import guard and doc coupling) vs SEC-8/SEC-17/DOC-9/DOC-14 (both required) | **Controls kept.** PI-8's rev-1 claim was demonstrably false without guard 2, and DOC-14 showed the newest identifier (`SLACK_TEAM_ID_ACP`) is unverifiable without the coupling test. Both are small Go tests, not frameworks. |
| SCOPE-17 (ACL validation exceeds the phase) vs SEC-7 (validate the ACL) | **Split on the parity line**: acceptance semantics stay oracle-identical (no rejection), while dedupe and DoS caps — which change no valid behavior — are kept, and non-conforming IDs are reported rather than refused. Strict rejection reassigned to P9/P10. |

**Machine-verifiable gates**, runnable locally and in CI:

1. `go build ./...` and `go vet ./...` clean.
2. `go test ./... -race` green, including the S3 leak matrix and both E3 negative controls.
3. `CGO_ENABLED=0` build for `linux/amd64`, `windows/amd64`, `darwin/amd64`, `darwin/arm64`.
4. `golangci-lint run ./...` clean at the pinned v2.13.2.
5. `govulncheck ./...` reports zero affecting vulnerabilities.
6. `go mod tidy` produces no diff; `go.sum` includes only the four verified modules.
7. The E7 secret-scan test passes over `git ls-files` (`.env.local` asserted untracked and
   excluded from the walk), as part of `go test ./...` — no separate CI gate (SCOPE-13).
8. `markdownlint` clean on all four changed/added documents.
9. The E7 doc/error coupling test passes, binding every error identifier to the reference doc.

**Negative controls** — each is an **executable `go:build mutation` sub-test** that re-runs on
every CI run, not a historical note (SCOPE-7):

* Deleting `secret.Format` fails the S3 `%d` row.
* Removing `Secret`'s pointer indirection fails the S3 `%p`, `%w`, and unexported-field rows.
* Unboxing `Resolved.authorizedUserIDs` fails the D3 `%p` row (**verified**: the plain unexported
  slice leaks under `%p`, `%w`, and unexported-field embedding).
* Adding a `Secret` or a `team_id` field to `config.Config` fails E3 guard 1.
* A second importer, an aliased import, a dot import, or `write := keyring.Set` each fail E3
  guard 2; a non-test importer of `credtest` also fails it.
* Removing an env var from `docs/credentials-reference.md`, or adding an undeclared `SLACK_*`
  identifier to it, each fail E7.
* A planted synthetic token, and a near-miss token containing `CANARY` plus extra secret material,
  each fail the E7 scan.

**Rollback boundary.** As above: one revert, no state, zero importers on merge.

**Closure artifacts.** A post-merge closure document under `docs/closure/` following the `003-S`
precedent, recording the merge SHA, verification results, the six oracle divergences (R1, R5, R6,
R7, R9, R14), the keychain provisioning round-trip results per backend (E5), the observed output
of each negative control, and the questions carried forward to `8ACF7110`, P8, and P13.

## Gate Status — BLOCKED (circuit breaker, P-013.6)

Plan review ran three adversarial rounds (5 personas each, `dispatch_mode: full` every round — no
persona was ever unavailable, so no P-012 degradation applies).

| Attempt | Security | Go/Correctness | Architecture | Scope | Schema-CLI-Docs |
|---|---|---|---|---|---|
| 1 (rev 1) | FAIL | FAIL | FAIL | FAIL | FAIL |
| 2 (rev 2) | FAIL | FAIL | **PASS** | FAIL | **PASS** |
| 3 (rev 3) | FAIL | FAIL | FAIL | FAIL | FAIL |

The plan-review attempt counter reached 3 with consecutive FAILs, opening the circuit breaker.
Stage therefore **halts the review loop and does not harvest**. Creating a backlog hierarchy or a
queued shipment from a FAILed plan would be a P-005 violation.

**Escalation route resolved** from a freshly re-read `.autoharness/config.yaml`
(`model_routing.stage.escalation`): `gpt-5.6-sol` / `openai` / `high`. The legacy flat
`model_routing.escalation` block is empty, so there is no F02FD596 both-present ambiguity. Stage's
own role route is `claude-opus-5`, so the escalation tuple differs — **not** `ESCALATION_DEGRADED`.
Payload: `docs/memory/2026-09-03-stage-p3-credentials-escalation.md`.

### Convergence analysis

The three rounds show real convergence on substance and divergence on specification detail:

* **Round 1 → 2:** 20 P1s across five personas → two personas passing. Every P1 was a genuine
  design or security defect, and three were *empirically verified* (the `fmt` bad-verb bypass, the
  unexported-field bypass, the unimplementable dual-`errors.Is` contract).
* **Round 2 → 3:** the surviving P1s were the ACL leak (verified and fixed) and the startup-wiring
  shortfall (fixed). But rev 3's added detail became its own defect surface.
* **Round 3:** **every persona confirms no P0**, and every persona confirms the security-critical
  design is sound — pointer-boxed `Secret` plus opaque `acl` closes all three verified reflective
  leak paths, `Reason` is a closed enum so backend text is unassignable, the library never logs,
  and the error composite satisfies all three `errors.Is` targets.

The residual P1s are almost entirely **plan-specification defects about test scaffolding** — which
Go test package a file belongs to, whether a `go:build mutation` tag is reachable from the gate
command, whether a canary literal collides with a scan allowlist. These are implementation-time
decisions being litigated at plan granularity, which is why each remediation round creates new
ones. That is precisely the signal the circuit breaker exists to catch, and it is the reason Stage
stops here rather than writing a rev 4.

### Consolidated remediation queue (for the escalated review or the next session)

Ordered by severity. Every item is concrete and small; none requires re-deliberation.

| # | Finding | Fix |
|---|---|---|
| 1 | **`credtest` import cycle** (Go NEW-1, ARCH NEW-1, P1). C1/C2/C3 declare `Test package: credentials` while their acceptance requires `credtest.ConformanceTest`; `credtest` imports `credentials`, so an in-package test importing it cannot compile. | Split each of C1/C2/C3 into an in-package file (unexported-state assertions) and a `package credentials_test` file (conformance). Add the missing `Test package:` lines to S1–S3, C2, E1, E2, E4, E8. Budget the extra test file in the granularity rule. |
| 2 | **E8 performs live keychain I/O** (Scope, P1). `RunE` calls `DefaultSources()`, so the startup test can hit the real OS keychain — violating PI-5 and the deterministic-fake requirement. | Give `RunE` a package-level source seam injected in tests (mirroring C3's `get` field pattern), or have `newRootCmd` accept a `credentials.Sources`. Assert the startup test never constructs `DefaultSources()`. |
| 3 | **Canary vs scan allowlist collision** (DOC NEW-1, P1). E7's exact-token allowlist contains only doc placeholders, but S3/D3/B3/C3/E8 mandate `xoxb-CANARY-MUST-NOT-APPEAR` in tracked test files, so the scan fails on the plan's own artifacts. | Export canonical canary constants (e.g. `secrettest.TokenCanary`, `MemberIDCanary`) and define the allowlist as {doc placeholders} ∪ {canary constants}, still exact-token. Keep the near-miss row. |
| 4 | **`SourceCoherence()` collapses env tiers** (SEC-6, P1). Env `Ref`s have `Service() == ""`, so ACP-env and shared-env are indistinguishable and a genuinely mixed pair reads as coherent. | Use a closed four-value tier identity (`ACPKeychain`, `ACPEnv`, `SharedKeychain`, `SharedEnv`) rather than `(SourceKind, service)`. Add the ACP-env-vs-shared-env table row. |
| 5 | **`go:build mutation` controls never run** (Go NEW-3, Scope, P1). No gate passes `-tags mutation`, and a tagged variant cannot redeclare `Secret`/`Resolved` in-package anyway. | Declare leaky variants under distinct names (`leakyResolved`, `unboxedSecret`) in ordinary untagged `_test.go` files so gate 2 runs them. |
| 6 | **Remaining non-executable criteria** (SCOPE-7, P1). C1, C3, D4, E2, E5, E6 retain "enforced by construction", "type-enforced", "impossible because the fields are named", and prose-presence doc criteria. | Replace each with a named test plus expected result, or move the claim to rationale prose and drop it from acceptance. |
| 7 | **ACL dedupe/caps still diverge from the oracle** (SCOPE-17, P1). Dedupe and the 8 KiB / 1000-ID caps change results for inputs the oracle accepts, contradicting the stated parity line. | Either drop dedupe and caps to restore strict parity, or move them (with the non-conforming-ID reporting) out of P3 and reclassify the divergence explicitly in the divergence list. |
| 8 | **Flag-table test cannot import `package main`** (DOC NEW-5, P2). | Home the flag assertion in `cmd/intercom/docs_flags_test.go` (`package main`); keep identifier coupling in `internal/credentials`. |
| 9 | **`go/packages` is an undeclared dependency** (Scope, P2). E3 guard 2 needs `golang.org/x/tools`, contradicting "one new direct dependency" and gate 6. | Either add `x/tools` to the declared budget and re-verify gates 3/5/6, or implement guard 2 with `go/ast` + `go/parser` from the standard library. |
| 10 | **Stale registries** (Go NEW-8, ARCH NEW-2, DOC, P2/P3). Risky-actions row still says `loadStartup` is "uncalled by `RunE`"; rollback still lists "one script" and "changes no existing package behavior"; PI-9 still says the scan excludes `*.env*`; gate 9 still describes error-text-driven coupling; duplicate `R13` decision IDs; unit count says 22 in one place and 23 in another; `type Kind uint8` still declared where `SourceKind` is used; Sub-epic S preamble still promises the dropped alias; `errors.As(err, &*apperr.Error)` is not valid Go; GO-18 and GO-24 have no recorded disposition. | Mechanical sweep. |
| 11 | **Attribution unreachable on the failure path** (ARCH NEW-4, P2). `Resolve` returns nil `*Resolved` on error, and `Attribution` is a field of it, so the diagnostic is lost exactly when needed. | Carry the partial `Attribution` on `resolveError` (accessor or `errors.As`-able), or return it as a third value. |
| 12 | **`acl` is a third hand-rolled redaction impl** (ARCH NEW-5, P2). Same cohesion argument that moved `Secret` applies to `acl`; its method set and receivers are unspecified. | Move it to `internal/secret` and factor a shared total-redaction helper + assertion block, or specify its method set explicitly. |
| 13 | **C4 guard map is unsynchronized** (Go NEW-5, P2). The abandoned worker clears its own entry, racing the next caller — a fatal concurrent map write, not just a `-race` report. | Specify a `sync.Mutex` (or channel semaphore) and state who removes the entry and when. |
| 14 | **Mid-slice checkpoint is unenforceable** (ARCH NEW-3, P2). Not present in the execution order; gates 3 and 5 are vacuous at S3 because `go-keyring` is not added until C3. | Insert an explicit `[GATE]` token between S3 and B1, scope it to gates 1/2/4/8 plus the leak matrix, and state the halt condition. |
| 15 | **DOC-14 mechanism gaps** (P2). The constant set holds four env *base* names while the doc must list all eight full names, so the reverse check fails on `SLACK_TEAM_ID_ACP`; and `agent-intercom*` collides with `ipc_name`, `agent-intercom.db`, and the repo slug. | Compose the expected set as `{base} × {"", EnvSuffixACP}`; scope service matching to whole tokens with explicit exclusions. |
| 16 | **Operator-visible behavior change undocumented** (DOC NEW-4, P2). After E8, `intercom` fails on missing config/credentials instead of returning not-implemented; README's "not yet wired" line becomes false. | Require the README Status section to state the change and delete the stale claim; add a `docs/config-reference.md` migration bullet and an E8 acceptance row. |
| 17 | **Self-referential deferrals** (DOC, P2). R12 and the SEC-5 carry-forward both defer to "the phase that wires `RunE`" — which is now this phase. | Re-key to a concrete owner (the phase that ships server behavior / owns a startup policy surface). |

### What is NOT in dispute

All five personas, in all three rounds, agree on: no P0 at any revision; the oracle analysis and
the six stash corrections; the four-tier precedence design; the provider seam; the dependency
choice and its measured CGO-free/vulnerability/licence posture; `Result`-over-`error`; the closed
`Reason` enum; the no-library-logging decision; and the pointer-boxing redaction mechanism. The
disagreement is entirely about specification completeness of the test scaffolding.
