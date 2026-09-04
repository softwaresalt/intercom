---
title: "Implementation Plan — intercom-go P3: Credential Resolution"
description: "Implement the oracle-grounded credential chain, read-only keychain seam, and pointer-boxed redaction types"
date: 2026-09-03
status: blocked
revision: 6
phase: P3
source_deliberation: "docs/decisions/2026-09-03-intercom-go-p3-credentials-deliberation.md"
parent_deliberation: "docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md"
source_brief: "docs/decisions/2026-07-06-go-port-reference-brief.md"
behavioral_oracle:
  repo: "softwaresalt/agent-intercom"
  commit: "41df772"
  access: "read-only, out-of-tree"
stash_entries: ["037B1552"]
predecessor: "003-F / 003-S (P2) — shipped at 40c7d289a60548e689ce714dd705411d9002a214"
successor: "P4 models + SQLite risk slice (stash 4989A42D)"
requires_plan_hardening: yes
plan_hardened: yes
plan_review_verdict: FAIL
gate_status: blocked-escalation-failed
normal_plan_review_attempts: 3
escalation_review_attempts: 3
escalation_route: "gpt-5.6-sol / openai / high"
tags: ["go", "port", "credentials", "keychain", "secret", "redaction", "p3"]
---

<!-- plan-review-attempt: 3 -->
<!-- escalation-plan-review-attempt: 3 -->

# Implementation Plan — intercom-go P3: Credential Resolution

## Escalation Context

Revision 3 reached the normal Stage review circuit breaker after three full five-persona reviews.
No P0 was raised. Revision 4 was the first configured escalation-route repair; its five-persona
review found eight P1s. Revision 5 resolved those but its re-review found four new P1s. Revision 6
resolves all four and the associated P2 contract gaps while preserving the accepted deliberation.

The exact revision-3 residual P1s and their final resolutions are:

| # | Residual P1 | Final resolution |
|---|---|---|
| 1 | `credtest` import cycle: in-package `credentials` tests could not import a helper that imports `credentials` | Remove the shared conformance helper. In-package tests inspect unexported state directly; external `credentials_test` tests alone import `credtest`. Every test file's package is specified below. |
| 2 | `RunE` tests could perform live keychain I/O through `DefaultSources()` | Remove revision 3's unauthorized `RunE` wiring and restore accepted deliberation P3-D8(b). A `tests/integration` test composes the real config and credential APIs with explicit fakes; `cmd/intercom` remains unchanged and cannot touch a keychain in P3. |
| 3 | The custom secret scan matched the plan's own Slack-shaped canary | Remove the redundant custom repository scanner and rely on the existing CI secret scanner. Rendering tests use non-token-shaped canaries (`token-canary-must-not-appear` and `member-canary-must-not-appear`), while documentation uses the existing closed placeholder vocabulary. |
| 4 | `SourceCoherence()` collapsed ACP and shared env tiers because both had an empty service | Add a closed `Tier` enum (`ACPKeychain`, `ACPEnv`, `SharedKeychain`, `SharedEnv`) to every attempt and winning attribution. Coherence compares `Tier`, with an explicit ACP-env/shared-env test. |
| 5 | `go:build mutation` fixtures were unreachable and could not redeclare production types | Use ordinary, untagged `_test.go` fixtures with distinct names. `unboxedSecret` and `unboxedStringList` deliberately leak on reflective fallbacks; `unformattedSecret` fails exact direct-format expectations. All controls run under normal `go test`. |
| 6 | Several acceptance criteria were prose claims rather than executable checks | Structural claims move to rationale. Acceptance criteria below name concrete test functions, inputs, and expected results. The standard-library AST containment guard replaces undeclared `go/packages` machinery. |
| 7 | ACL deduplication and size caps contradicted oracle parity | Remove deduplication, caps, and ID-shape reporting. The P3 parser exactly performs split, trim, and empty-drop while preserving order and duplicates. Authorization validation belongs to P9/P10. |

Revision 4 also closes the ten P2/P3 queue items from the failed gate: failure-path attribution is
returned separately from `*Resolved`; ACL storage uses a reusable pointer-boxed
`secret.StringList`; keychain concurrency uses a one-slot semaphore instead of an unsynchronized
map; the mid-slice gate is executable and correctly scoped; doc coupling derives all eight env
names from production chain data; the accepted unwired startup boundary is preserved; deferrals name concrete successor phases; stale
counts, aliases, error syntax, and rollback statements are removed.

## Problem Frame and Fixed Oracle Contract

P2 permanently established `internal/config` as a secret-free file-schema package. P3 must not
add `app_token`, `bot_token`, `team_id`, or `authorized_user_ids` to `config.Config`.

The implementation oracle is `softwaresalt/agent-intercom` at commit `41df772`, read-only at
`C:\Source\GitHub\intercom`. The fixed P3 contract is:

1. `app_token`, `bot_token`, and optional `team_id` use a static ACP four-tier chain:
   `agent-intercom-acp` keychain, `*_ACP` environment, `agent-intercom` keychain, shared
   environment.
2. Resolution is per field. Empty values are misses. First hit wins and later tiers are not read.
3. `app_token` and `bot_token` are required. `team_id` is optional and defaults to `""`.
4. `authorized_user_ids` is a separate env-only path:
   read `SLACK_MEMBER_IDS_ACP`; only when it is absent read `SLACK_MEMBER_IDS`; then split the
   selected value on comma, trim, drop empty, preserve order and duplicates, and require at least
   one result. A present-but-empty ACP value is selected and fails required-nonempty validation;
   it does not fall through to shared. It never reads a keychain.
5. Keychain `Miss` and `Unavailable` both permit fallback. `Unavailable` is retained in
   attribution; raw backend error text is never retained. Caller cancellation aborts resolution;
   an internal keychain timeout is `Unavailable` and permits env fallback.
6. The package never logs. It returns structured attribution for the first future startup
   composition root to log once at WARN when any source is unavailable.
7. The real adapter is read-only and is the only product file importing
   `github.com/zalando/go-keyring` v0.2.8.

The Go port intentionally improves the oracle by distinguishing `Miss` from `Unavailable` and by
redacting tokens and the operator ACL. These changes do not alter which credential value wins.

## Scope

**In scope:** `internal/secret`, `internal/credentials`, one boundary-only test in
`internal/config`, the read-only `go-keyring` adapter, deterministic test support, architecture
guards under `tests/arch`, an executable load/validate/resolve composition test under
`tests/integration`, and credential/operator documentation.

**Out of scope:** models, persistence, Slack networking, ACP processes, credential writes or
rotation, strict Slack member-ID validation, tenant-consistency enforcement, credential hot
reload, a credential-management CLI, IPC credentials, CI workflow changes, oracle-drift
automation, wiring credentials into `cmd/intercom`, and all P4+ work in `4989A42D`.

No product, test, or configuration code is changed during Stage. This document is the execution
specification for Ship.

## Exact API and Ownership Contract

### `internal/secret`

```go
package secret

const Redacted = "[REDACTED]"

type Secret struct {
	value *string
}

func New(value string) Secret
func (s Secret) Expose() string
func (s Secret) IsEmpty() bool
func (s Secret) Equal(other Secret) bool
func (s Secret) Format(state fmt.State, verb rune)
func (s Secret) String() string
func (s Secret) GoString() string
func (s Secret) MarshalJSON() ([]byte, error)
func (s Secret) MarshalText() ([]byte, error)
func (s Secret) LogValue() slog.Value
func (s Secret) GobEncode() ([]byte, error)

type StringList struct {
	values *[]string
}

func NewStringList(values []string) StringList
func (s StringList) Values() []string
func (s StringList) Len() int
func (s StringList) Format(state fmt.State, verb rune)
func (s StringList) String() string
func (s StringList) GoString() string
func (s StringList) MarshalJSON() ([]byte, error)
func (s StringList) MarshalText() ([]byte, error)
func (s StringList) LogValue() slog.Value
func (s StringList) GobEncode() ([]byte, error)
```

Both types are immutable value handles over pointer-boxed storage. Constructors copy caller-owned
input where applicable; `StringList.Values` returns a fresh copy. Zero values are valid and safe.
All rendering methods use value receivers and one unexported `writeRedacted` helper. `%q` renders
the quoted redaction marker; width and precision are ignored. JSON and text marshaling return only
the marker. `slog` returns only the marker. Gob returns a fixed error and no bytes.

No unmarshal method exists. No alias is exported from `credentials`. `Secret.Expose` and
`StringList.Values` are the explicit trusted-boundary APIs. Direct, deliberate reflection that
dereferences unexported storage (or `unsafe`) is outside the guarantee because in-process Go code
cannot be made opaque from itself; the protected invariant covers every standard `fmt`,
`encoding/json`, `encoding/gob`, `text/template`, `slog`, and error-composition path, including
their ordinary reflection fallback.

`Secret.Equal` compares contents with `subtle.ConstantTimeCompare`; length remains observable.
Plain `==` compares pointer identity and is forbidden for semantic equality and map/set use.

### `internal/credentials`

```go
package credentials

type Outcome uint8
const (
	Miss Outcome = iota
	Hit
	Unavailable
)

type Reason uint8
const (
	ReasonNone Reason = iota
	ReasonNoEntry
	ReasonEmptyEntry
	ReasonUnsupportedPlatform
	ReasonBackendError
	ReasonBusy
	ReasonTimeout
	ReasonCanceled
)

type SourceKind uint8
const (
	SourceKeychain SourceKind = iota + 1
	SourceEnv
)

type Tier uint8
const (
	TierNone Tier = iota
	TierACPKeychain
	TierACPEnv
	TierSharedKeychain
	TierSharedEnv
)

type Ref struct { /* unexported fields */ }
func EnvRef(name string) Ref
func KeychainRef(service, account string) Ref
func (r Ref) Kind() SourceKind
func (r Ref) Service() string
func (r Ref) Key() string
func (r Ref) String() string

type Result struct { /* unexported fields */ }
func HitResult(value secret.Secret) Result
func NoEntryResult() Result
func EmptyResult() Result
func UnsupportedResult() Result
func BackendUnavailableResult() Result
func BusyResult() Result
func TimedOutResult() Result
func CanceledResult() Result
func (r Result) Outcome() Outcome
func (r Result) Value() secret.Secret
func (r Result) Reason() Reason

type Source interface {
	Lookup(ctx context.Context, ref Ref) Result
}

type Attempt struct {
	// fields are unexported; accessors return immutable values.
}

type FieldAttribution struct {
	// fields are unexported; Attempts returns a deep copy.
}

type Attribution struct {
	// fields are unexported; Fields returns a deep copy.
}

type Field uint8
const (
	FieldAppToken Field = iota + 1
	FieldBotToken
	FieldTeamID
	FieldAuthorizedUsers
)

func (a Attempt) Tier() Tier
func (a Attempt) Ref() Ref
func (a Attempt) Outcome() Outcome
func (a Attempt) Reason() Reason
func (f FieldAttribution) Field() Field
func (f FieldAttribution) Resolved() bool
func (f FieldAttribution) From() Tier
func (f FieldAttribution) Attempts() []Attempt
func (a Attribution) Fields() []FieldAttribution
func (a Attribution) AnyUnavailable() bool
func (a Attribution) SourceCoherence() []Tier
func (a Attribution) LogValue() slog.Value

type Sources struct {
	Env      Source
	Keychain Source
}

func DefaultSources() Sources
func NewResolver(sources Sources) (*Resolver, error)
func (r *Resolver) Resolve(ctx context.Context) (*Resolved, Attribution, error)

type Resolved struct {
	// all fields are unexported; authorizedUserIDs is secret.StringList.
}

func (r Resolved) AppToken() secret.Secret
func (r Resolved) BotToken() secret.Secret
func (r Resolved) TeamID() string
func (r Resolved) AuthorizedUserIDs() []string
func (r Resolved) Format(state fmt.State, verb rune)
func (r Resolved) String() string
func (r Resolved) GoString() string
func (r Resolved) MarshalJSON() ([]byte, error)
func (r Resolved) LogValue() slog.Value
func (r Resolved) GobEncode() ([]byte, error)
```

`HitResult(secret.New(""))` returns the same state as `EmptyResult`; no hit can carry an empty
value. The other constructors fix the only valid outcome/reason combinations and accept no
caller-provided reason. Zero `Result` is `Miss`/`ReasonNone` and is treated defensively as a miss.
Unknown result/ref states are unrepresentable to external `Source` implementations because fields
are unexported and constructors return only valid states. In-package constructor tests pin the
zero state; the resolver does not carry an unreachable unknown-enum branch.

`Resolve` returns:

* success: non-nil `*Resolved`, complete `Attribution`, nil error;
* failure: nil `*Resolved`, partial `Attribution` covering every attempted lookup, non-nil error.

`Resolved` owns copies of the ACL input, exposes values only through value-returning accessors,
and does not duplicate the separately returned attribution. `Attribution` owns source descriptors
and enums only; every slice is unexported and every accessor deep-copies nested slices, so callers
cannot mutate resolver state or insert loggable values. It
can never contain a token, team ID, member ID, or backend error string. `SourceCoherence` returns
the sorted distinct winning tiers for `app_token`, `bot_token`, and resolved `team_id`; zero or
one tier is coherent, two or more tiers is mixed. Authorized-user env selection is reported but
excluded from this three-field coherence set because it has no keychain chain.

Plaintext access is deliberately narrow and greppable. P3 product code may call `Secret.Expose`
only in `resolve.go` to convert the non-secret `team_id` and in `authusers.go` to parse the selected
ACL CSV. It may call `StringList.Values` only in `resolved.go` to implement the defensive-copy
accessor. E1b rejects other product-code call sites. Later phases may add consumers only through
their own reviewed plans.

Sentinels are `ErrCredentialNotFound`, `ErrNoAuthorizedUsers`, `ErrInvalidSources`,
`ErrResolutionCanceled`, and `ErrResolverAlreadyUsed`. `NewResolver` rejects either nil source,
returns a nil resolver, and returns a config-class error matching `ErrInvalidSources` and
`apperr.ErrConfig`. Credential failures use:

```go
type resolveError struct {
	app      *apperr.Error
	sentinel error
	cause    error
}

func (e *resolveError) Error() string
func (e *resolveError) Unwrap() []error
```

`Error` delegates to `app.Error()`. `Unwrap` returns the non-nil members of
`[]error{app, sentinel, cause}`. Ordinary resolution failures have no cause. Caller cancellation
stores `ctx.Err()` as the cause, preserving `context.Canceled` or `context.DeadlineExceeded`
identity without changing the fixed `apperr` display. Tests use valid Go:

```go
var appErr *apperr.Error
errors.As(err, &appErr)
```

and require the same error to match its credential sentinel and `apperr.ErrConfig`; cancellation
also matches the exact context cause. Error text is derived only from `[]Attempt`, naming refs,
tiers, and closed reasons; it never uses a credential value or backend error text.

### Provider behavior and lifecycle

`envSource` stores `lookup func(string) (string, bool)`, defaulting to `os.LookupEnv`. It is
synchronous, immutable after construction, and has no `Unavailable` path.

`keychainSource` stores an injected
`get func(service, account string) (string, error)`, a timeout, and a one-slot channel semaphore.
The exact constructors are:

```go
func newKeychainSource() Source // keyring.Get plus the five-second production timeout
func newKeychainSourceFor(
	get func(service, account string) (string, error),
	timeout time.Duration,
) Source // tests only
```

Tests use only `newKeychainSourceFor`. The AST guard rejects `newKeychainSource`,
`DefaultSources`, and `keyring.Get` from test files.

For each lookup:

1. nil context becomes `context.Background`;
2. `ctx.Err()` is checked before semaphore acquisition; cancellation returns
   `ReasonCanceled` without calling `get`;
3. failure to acquire the semaphore immediately yields `ReasonBusy`;
4. after acquisition, `ctx.Err()` is checked again before launching a worker; cancellation
   releases the permit and returns `ReasonCanceled` without calling `get`;
5. the acquired worker calls `get`, sends once to a buffer-one result channel, and releases the
   semaphore in a defer;
6. the caller selects between the result, caller cancellation, and the per-call timer;
7. timeout yields `ReasonTimeout`; the worker may remain blocked, but it retains the only
   semaphore permit, so no later call can create another orphan;
8. `ErrNotFound` is `Miss`/`ReasonNoEntry`; `ErrUnsupportedPlatform` is
   `Unavailable`/`ReasonUnsupportedPlatform`; all other errors are
   `Unavailable`/`ReasonBackendError`; raw error text is discarded.

The resolver is single-use and protects its `used` flag with a mutex, so simultaneous calls yield
one execution and `ErrResolverAlreadyUsed` for every loser. Caller cancellation aborts the
resolver with `ErrResolutionCanceled`; backend unsupported/error/busy/timeout outcomes continue
to the next tier. The resolver performs at most one real keychain call at a time and creates no
unbounded worker set. The semaphore is shared across both service tiers: if an ACP-service lookup
times out and remains blocked, the later shared-service lookup returns `ReasonBusy` without
calling its backend. `TestKeychainSourceTimeoutAndBusy` asserts that exact timeout-then-busy
sequence. Credentials are loaded once at startup and are never refreshed.

`Resolver.Resolve` checks `ctx.Err()` before every chain step and again before the
authorized-user path. Pre-canceled and expired-deadline resolver tests require nil `*Resolved`,
the exact partial attribution accumulated before cancellation, no later source lookup, zero
injected keychain calls when canceled before the first step, and matches for
`ErrResolutionCanceled`, `apperr.ErrConfig`, and the exact context error.

### Composition boundary

Accepted deliberation P3-D8(b) remains binding: P3 does not wire credentials into
`cmd/intercom.RunE` and does not add a production bootstrap helper. `cmd/intercom/main.go` and its
existing tests remain unchanged.

`tests/integration/credentials_startup_test.go` (`package integration_test`) composes the actual
public APIs in the future startup order:

```go
cfg, report, err := config.Load(path) // Load includes Validate.
if err != nil {
	// Assert the fake source call log is empty.
}
resolver, err := credentials.NewResolver(fakeSources)
resolved, attribution, err := resolver.Resolve(ctx)
```

The test is not a second implementation: it calls the real package entry points in sequence and
asserts an invalid config leaves both source call logs empty, while a valid config produces one
complete resolution. It also asserts partial attribution remains available on credential failure.
The package documentation for `internal/credentials` states this required ordering.

`DefaultSources()` is product-only construction for the future composition root. The AST guard
rejects any `_test.go` reference to `DefaultSources`, whether qualified, aliased, unqualified, or
assigned as a function value. `source_keychain_test.go` may import `go-keyring` only to reference
`ErrNotFound` and `ErrUnsupportedPlatform`; the guard rejects `Get` and every write selector from
tests. Therefore P3 tests can classify the real sentinels without reaching a real adapter.

The first phase that introduces a real server startup pipeline must call this sequence once and
must log exactly one WARN record when `Attribution.AnyUnavailable()` is true. That concrete owner
is P8, where Slack identity is first consumed and source/tenant policy can be enforced together.

## Test Scaffolding and Acceptance Matrix

No test accesses a real keychain. No test uses a mutable package-level provider hook. Synthetic
values are never real-token-shaped.

| File | Package | Exact responsibility |
|---|---|---|
| `internal/secret/secret_test.go` | `secret` | Zero/value ownership, equality, all rendering matrices, direct unexported-field fixtures |
| `internal/secret/stringlist_test.go` | `secret` | List ownership/copies, zero value, rendering, and unboxed fixture |
| `internal/secret/leak_external_test.go` | `secret_test` | Exported API behavior from a real consumer package |
| `internal/credentials/attribution_test.go` | `credentials` | Closed enums, refs, ordered attempts, four-tier coherence |
| `internal/credentials/errors_test.go` | `credentials` | Composite `errors.Is`/`errors.As`, derived messages, no canary |
| `internal/credentials/source_test.go` | `credentials` | Constructor-only result states and one-method source surface |
| `internal/credentials/source_env_test.go` | `credentials` | Injected env lookup only |
| `internal/credentials/source_keychain_test.go` | `credentials` | Injected `get`, classification, timeout, semaphore, cancellation |
| `internal/credentials/credtest/fake_test.go` | `credtest_test` | Fake programming and ordered call log |
| `internal/credentials/chain_test.go` | `credentials` | Unexported step/tier/ref definitions and canonical identifiers |
| `internal/credentials/resolve_test.go` | `credentials_test` | Four-tier matrix using `credtest`; no unexported access |
| `internal/credentials/authusers_test.go` | `credentials` | Exact two-tier CSV parity with a local counting source stub |
| `internal/credentials/resolved_test.go` | `credentials` | Aggregate ownership and full leak matrix |
| `internal/config/credential_boundary_test.go` | `config` | Reflection walk proving the file schema has no credential fields |
| `tests/arch/credentials_imports_test.go` | `arch_test` | Standard-library AST import/write-surface/test-support guard |
| `tests/integration/credentials_startup_test.go` | `integration_test` | Ordered public-API composition with explicit fakes |
| `internal/credentials/docs_coupling_test.go` | `credentials` | Production chain constants ↔ canonical credential reference |

In-package `credentials` tests never import `credtest`. Any test importing `credtest` uses an
external package.

### Rendering matrix

The `Secret`, `StringList`, and `Resolved` tests cover both values and pointers where applicable:

* `fmt`: `%v`, `%+v`, `%#v`, `%s`, `%q`, `%d`, `%x`, `%X`, `%T`, `%p`;
* error composition: `fmt.Errorf("wrapped: %w", value)` and
  `errors.Join(fmt.Errorf("wrapped: %w", value))`;
* ordinary reflection fallback: values nested in exported and unexported struct fields, slices,
  string-keyed maps, and `map[Secret]string`;
* `encoding/json`: bare value, pointer, struct field, and map key;
* `encoding/gob`: bare value and exported field must return a fixed error with no secret bytes;
  an unexported field is documented and tested as omitted by gob;
* `text/template`;
* `slog` JSON and text handlers, both direct and through a `ReplaceAttr` that calls `fmt.Sprint`;
* nil `*Secret`, zero values, and nil/empty lists.

Direct supported formatting rows (`%v`, `%+v`, `%#v`, `%s`, `%d`, `%x`, `%X`) require exactly
`[REDACTED]`; `%q` requires exactly `"[REDACTED]"`. `%T` requires the expected type name and `%p`
requires a non-secret address representation. Every composite row asserts that the token and
member canaries are absent and that expected redaction/count/type/address output is present, so
empty output cannot create a false pass. Untagged `unboxedSecret` and `unboxedStringList` fixtures
must expose their canaries on the known reflective fallback paths. `unformattedSecret` remains
pointer-boxed and safe from disclosure, but its direct-format output must differ from
`[REDACTED]`, proving exact-output tests detect removal of `Format`.

### Resolution matrix

One parallel table in `credentials_test` proves, separately for each chain-resolved field:

1. each of the four tiers wins in isolation;
2. ACP env beats shared keychain;
3. each empty tier is a miss;
4. keychain `ErrNotFound` falls through without unavailable attribution;
5. unsupported, backend error, busy, and timeout each fall through with unavailable attribution;
6. first hit short-circuits and the fake call log contains no later ref;
7. required-field exhaustion yields nil `*Resolved`, partial attribution, and the correct
   sentinel/config error chain;
8. optional `team_id` exhaustion yields `""`;
9. mixed ACP-env/shared-env winners produce two distinct `Tier` values;
10. concurrent calls on one resolver execute once and every other call returns
    `ErrResolverAlreadyUsed`;
11. pre-canceled and expired-deadline contexts abort with the exact error identities, partial
    attribution, no later source calls, and zero keychain `get` calls when cancellation precedes
    the first step.

`authusers_test.go` separately proves `_ACP` precedence; shared fallback only when ACP is absent;
present-but-empty ACP fails without shared lookup; exact whitespace/empty handling; duplicate
preservation; all-empty failure; and zero keychain calls.

### Boundary and containment tests

The config boundary walk recursively visits structs, pointers, arrays, slices, and map keys and
values with cycle protection. It rejects normalized field names or TOML tags matching
`app_token`, `bot_token`, `team_id`, `authorized_user_ids`, `password`, `secret`, `api_key`,
`oauth_token`, `signing_secret`, or `credential`, and rejects any `secret.Secret` or
`secret.StringList` field.

The import guard lives in `tests/arch`, uses only `go/parser`, `go/ast`, and `go/token`, finds the
module root by walking upward until `go.mod`, and positively enumerates the owned Go roots
`cmd`, `internal`, and `tests`. Any future top-level Go root requires an explicit guard update. It
maps every import alias for `github.com/zalando/go-keyring`, rejects dot imports, requires the
only product import in `internal/credentials/source_keychain.go`, and rejects any selector named
`Set`, `Delete`, or `DeleteAll` through that alias whether called or assigned. In
`source_keychain_test.go`, only `ErrNotFound` and `ErrUnsupportedPlatform` selectors are allowed;
`Get` is rejected. It also requires
every importer of `internal/credentials/credtest` to end in `_test.go`, and rejects
`DefaultSources` or `newKeychainSource` in any `_test.go` file. Fixture snippets
parsed from strings prove aliased imports, dot imports, second importers, function-value
assignment, and direct/aliased `DefaultSources` references are rejected.

## Implementation Units

Every unit is test-first, changes at most two files, stays within one skill domain, has no more
than four named test functions, and is estimated at no more than two human hours. `M` means the
full two-hour budget, not more than two hours. No task is `complexity: high`.

| ID | Unit | Files | Size | Complexity | Named tests and exact exit state |
|---|---|---|---|---|---|
| S1 | Pointer-boxed `Secret` | `internal/secret/secret.go`, `secret_test.go` | M | medium | `TestSecretOwnershipAndEquality`, `TestSecretRendering`, `TestSecretSerialization`, `TestSecretNegativeControls`; direct verbs match exact markers, all composite rows omit the canary, and unboxed/unformatted controls fail the protected expectations. |
| S1b | External `Secret` contract | `internal/secret/leak_external_test.go` | XS | low | `TestSecretExternalRendering`, `TestSecretHasNoDecodeSurface`; an outside package sees only the explicit API and negative interface assertions prove no JSON/text unmarshal methods. |
| S2 | Pointer-boxed `StringList` | `internal/secret/stringlist.go`, `stringlist_test.go` | S | medium | `TestStringListOwnership`, `TestStringListRendering`, `TestStringListSerialization`, `TestStringListNegativeControl`; constructor/accessor copies are independent and no member canary renders. |
| C1 | Attribution vocabulary | `internal/credentials/attribution.go`, `attribution_test.go` | M | medium | `TestAttributionEnumsAndRefs`, `TestAttributionAttemptOrder`, `TestAttributionSourceCoherence`, `TestAttributionOwnershipAndLogValue`; closed `Field` values drive the chain, ACP-env/shared-env yields exactly two sorted tiers, returned slices are independent, and logs contain only closed metadata. |
| C2 | Composite errors | `internal/credentials/errors.go`, `errors_test.go` | S | medium | `TestResolveErrorClassification`, `TestResolveErrorMessageFromAttempts`, `TestResolveErrorRedaction`, `TestResolveErrorCancellationIdentity`; one value matches its credential sentinel and `apperr.ErrConfig`, and cancellation additionally matches the exact context error. |
| C3 | Result and source contract | `internal/credentials/source.go`, `source_test.go` | S | low | `TestResultConstructors`, `TestSourceSurface`; zero and every constructor state are exact, empty hit becomes empty miss, and reflection sees exactly one `Source` method. |
| C3b | Environment adapter | `internal/credentials/source_env.go`, `source_env_test.go` | S | low | `TestEnvSourceLookup`; injected unset, empty, and hit rows return the exact result and never `Unavailable`. |
| C4a | Pin keychain dependency | `go.mod`, `go.sum` | XS | low | `go mod download github.com/zalando/go-keyring@v0.2.8` succeeds and `go list -m` shows the audited version; tidy is deliberately deferred until C4c creates the import. |
| C4b | Keychain adapter red harness | `internal/credentials/source_keychain_test.go` | S | medium | The four named C4c tests are written first and fail because `newKeychainSourceFor`/classification are absent; no production constructor is called. |
| C4c | Read-only keychain adapter | `internal/credentials/source_keychain.go` | M | medium | `TestKeychainSourceClassification`, `TestKeychainSourceCancellation`, `TestKeychainSourceTimeoutAndBusy`, `TestKeychainSourceRedaction` turn green; pre-cancel calls `get` zero times, timeout then cross-service busy creates one worker, and `go mod tidy` retains the dependency with no further diff. |
| C5 | Deterministic external fake | `internal/credentials/credtest/fake.go`, `fake_test.go` | S | low | `TestFakeSourceProgramming`, `TestFakeSourceConcurrentCallLog`; scripted states, miss default, and synchronized exact call order pass. |
| C6 | Source assembly and resolver lifecycle | `internal/credentials/sources.go`, `sources_test.go` | S | medium | `TestNewResolverRejectsNilSources`, `TestResolverSingleUse`; each nil-source case returns nil resolver and matches `ErrInvalidSources` plus `apperr.ErrConfig`, and only one concurrent caller enters resolution. `DefaultSources` is compiled but deliberately not invoked by tests; E1b pins its declaration and bans test call sites. |
| D3 | Redacted `Resolved` type | `internal/credentials/resolved.go`, `resolved_test.go` | M | medium | `TestResolvedOwnership`, `TestResolvedRendering`, `TestResolvedSerialization`, `TestResolvedNegativeControl`; all fields are unexported, accessors return values/copies, the full matrix hides both canaries, JSON/slog expose only count, TeamID stays visible, and gob fails. |
| D1a | Four-tier chain data | `internal/credentials/chain.go`, `chain_test.go` | S | medium | `TestCredentialChainDefinition`, `TestCredentialChainIdentifiers`; all four `step{tier,src,ref}` entries and all three field specs are exact and are the canonical identifier source. |
| D1b | Resolver orchestration | `internal/credentials/resolve.go`, `resolve_test.go` | M | medium | `TestCredentialPrecedence`, `TestCredentialFallbackAndAttribution`, `TestCredentialRequiredOptional`, `TestResolverCancellation`; every row asserts exact value, attempt order, call log, partial attribution, error identities, and cancellation performs no later lookup. |
| D2 | Authorized-user resolver | `internal/credentials/authusers.go`, `authusers_test.go` | S | low | `TestAuthorizedUsersPrecedence`, `TestAuthorizedUsersCSVParity`, `TestAuthorizedUsersRequired`, `TestAuthorizedUsersNoKeychain`; only absent ACP falls back, present-empty ACP fails without shared lookup, duplicates/order remain, and the local keychain stub count is zero. |
| E1a | Secret-free config guard | `internal/config/credential_boundary_test.go` | S | medium | `TestConfigCredentialBoundary`, `TestConfigCredentialBoundaryCycles`; recursive type/name/tag checks reject synthetic forbidden fields and terminate on self-reference. |
| E1b | Dependency containment guard | `tests/arch/credentials_imports_test.go` | M | medium | `TestKeyringContainment`, `TestCredtestImportContainment`, `TestTestsCannotUseRealCredentialSources`, `TestDependencyOrderIsTopological`; positive-root walk and every parsed mutation fixture produce the specified result. It also permits production `Expose` only in `resolve.go`/`authusers.go` and `StringList.Values` only in `resolved.go`. |
| E2 | Unwired startup-order integration | `tests/integration/credentials_startup_test.go` | S | medium | `TestCredentialStartupOrder`, `TestCredentialStartupFailureAttribution`; invalid config yields zero source calls, valid config resolves once, and credential failure returns nil resolved plus partial attribution. |
| E6a | Credential-doc contract harness | `internal/credentials/docs_coupling_test.go` | S | medium | `TestCredentialReferenceIdentifiers`, `TestCredentialReferenceSemantics`; in-package tests enumerate actual chain refs and authorized-user env constants, assert cardinality eight/two/three only as completeness checks, and reject unknown exact code-span identifiers. |
| E4 | Canonical credential reference | `docs/credentials-reference.md` | S | low | E6a turns green with the per-field table, parity rules, trust model, fake/real boundary, and explicit “implemented but not wired into `RunE`” status. |
| E5a | README and example surfaces | `README.md`, `config.toml.example` | S | low | E6a turns green; both link to E4, retain current CLI behavior without adding an unrelated flag table, and state config files remain secret-free. |
| E5b | Config reference migration | `docs/config-reference.md` | XS | low | E6a turns green; all four omitted credential fields, `_ACP` precedence migration, and the unwired P3 status are accurate. |

### Required mid-slice gate

Execution order contains this explicit stop:

```text
S1 -> S2 -> [GATE-SECRET]
```

`[GATE-SECRET]` runs the smallest existing gates that cover the primitive:
`gofmt -l internal/secret`, `go vet ./internal/secret`, and
`go test -race ./internal/secret`. Ship halts if any rendering row or leaky-control expectation
fails. Dependency, vulnerability, cross-compile, and documentation gates are intentionally not
run here because their relevant files do not exist yet.

## Dependency Graph and Execution Order

```text
S1 -> S2 -> [GATE-SECRET]
S1 -> S1b
S1 -> C1 -> C2
C1,S1 -> C3 -> C3b
C4a -> C4b
C1,C3,C4b -> C4c
S1,C1,C3 -> C5
S1,S2,C1 -> D3
C3b,C4c -> C6
C1,C3,C3b -> D1a
C1,C2,C3,C3b -> D2
C2,C5,C6,D3,D1a,D2 -> D1b
S1,S2 -> E1a
C4a,C4c,C5,C6,D1b,D2,D3 -> E1b
C5,D1b,D2,D3 -> E2
D1a,D2 -> E6a -> E4 -> E5a -> E5b
```

Serial order:
`S1, S1b, S2, GATE-SECRET, C1, C2, C3, C3b, C4a, C4b, C4c, C5, C6, D3,
D1a, D2, D1b, E1a, E1b, E2, E6a, E4, E5a, E5b`.
`TestDependencyOrderIsTopological` in `tests/arch/credentials_imports_test.go` encodes these edges
as fixture data and requires the serial list to be a valid topological linearization.
One implementation branch and one worktree; no parallel Ship execution.

## Documentation Contract

`docs/credentials-reference.md` is canonical and must include:

* a four-column tier table for `app_token`, `bot_token`, and `team_id`;
* an explicit two-tier env-only row for `authorized_user_ids`, with keychain cells marked `n/a`;
* all eight full env names, both service names, and all three keychain accounts;
* `SLACK_APP_TOKEN_ACP` beating a shared keychain entry as the precedence counterexample;
* required/optional semantics, empty-as-miss behavior, and exact ACL CSV behavior;
* the trust boundary: any process running as the same OS user may influence that user's keychain;
* the environment trust boundary: the parent process, service manager, service-definition owner,
  and any environment-file writer can choose tokens and the future authorization ACL; deployed
  service definitions and environment files require restrictive ownership and permissions;
* `Miss` versus `Unavailable`, including that fallback continues and raw backend text is dropped;
* the fake/real boundary: production uses `go-keyring`; tests use injected sources and never touch
  an OS keychain;
* memory-residency limits: no zeroization guarantee for Go strings;
* startup status: the packages and their required load/validate/resolve order are implemented and
  integration-tested, but P3 deliberately does not call them from `RunE`; current CLI startup
  behavior remains unchanged until P8 introduces the first Slack-consuming startup pipeline.

The coupling test enumerates the actual production refs in the three chain specs plus the two
authorized-user env constants; it does not reconstruct names from a second base/suffix convention.
It asserts the resulting cardinalities (eight env names, two services, three accounts) as
completeness checks. Matching uses exact Markdown code spans, not the broad `agent-intercom*`
pattern, so `ipc_name`, database filenames, and repository names cannot collide. P3 adds no
unrelated CLI-flag documentation contract.

No provisioning command is claimed verified by this slice. Credential writes are out of scope.
If documentation mentions external OS tooling, it links to platform documentation without
claiming a tested round trip.

## Security and Correctness Invariants

| ID | Invariant | Enforcement |
|---|---|---|
| PI-1 | Token and ACL values never appear through standard fmt/JSON/gob/template/slog/error rendering, including ordinary reflection fallback | Pointer-boxed storage, total value-receiver render methods, complete matrix, leaky controls |
| PI-2 | `config.Config` remains structurally credential-free | E1 recursive reflection guard |
| PI-3 | Non-hit provider results carry no value | Constructor-only result states and C3 tests |
| PI-4 | `Unavailable` never synthesizes a value and falls through except caller cancellation | C4/D1 tests |
| PI-5 | Tests never access an OS keychain | No P3 `RunE` wiring, injected `get`, fake call logs, AST rejection of test `DefaultSources`/`keyring.Get` |
| PI-6 | Value selection is oracle-equivalent | D1/D2 matrices |
| PI-7 | Credential errors preserve the existing `apperr` display and kind contract | C2 composite tests against real `apperr` |
| PI-8 | Product code cannot call the dependency's write APIs unnoticed | E1 AST guard |
| PI-9 | Partial attribution remains available on failure without containing values | `Resolve` triple return and error-path tests |
| PI-10 | At most one keychain worker per source can remain in flight | One-slot semaphore and repeated-timeout test |

Accepted residual risks:

* **Precedence poisoning:** a process with the same OS-user authority can provision a higher-tier
  keychain entry, and a principal controlling the service environment can choose env-tier values
  including the future ACL. P3 documents these trust properties and reports the winning tier. A
  policy that rejects mixed or unexpected sources belongs to P8, where a real Slack identity is
  available.
* **Cross-tier identity mixing:** preserved for oracle parity and made visible by
  `SourceCoherence`. P8 must verify the resolved Slack tenant before using both tokens.
* **Memory residency:** Go strings cannot be reliably zeroized. Process dump controls and child
  environment minimization belong to deployment/P6.
* **Blocked keychain worker:** the dependency cannot be canceled in-flight. The semaphore bounds
  the leak to one worker per source and env fallback remains available.

## Plan Hardening

Hardening is required because this slice handles live credentials, adds a third-party OS adapter,
and introduces a goroutine around an API that cannot be canceled.

### Threat model

| Threat | Boundary | Prevention | Detection / residual |
|---|---|---|---|
| Accidental token or ACL disclosure | Standard formatting, logging, marshaling, templates, error composition, ordinary reflection fallback | Pointer-boxed storage; total render methods; values excluded from attribution and errors | Exact-output leak matrices plus deliberately leaky test fixtures |
| Hostile or compromised keychain backend text | OS adapter → application error/log | Closed `Reason`; raw errors discarded after `errors.Is` classification | Backend canary tests |
| Higher-tier credential poisoning | Same-user keychain and service environment | No complete P3 prevention without changing oracle precedence | Winning `Tier` attribution; trust model documented; P8 owns source/tenant policy |
| Keychain hang | Non-cancelable dependency call | One-slot semaphore, buffer-one completion channel, five-second caller wait | Timeout/busy attribution; at most one blocked worker per source |
| Test access to real keychain | Test → `DefaultSources` or direct dependency | All adapter tests inject `get`; integration uses `credtest`; AST guard rejects real-source construction from tests | Mutation fixtures for direct, alias, and function-value forms |
| Config boundary erosion | TOML schema → credential storage | No decode methods; recursive config type/name/tag guard | Synthetic forbidden-field tests |
| Error-taxonomy breakage | Credentials errors → existing `apperr` callers | Multi-child `Unwrap`; display delegated to real `*apperr.Error` | Real `errors.Is`/`errors.As` tests |

### Failure and lifecycle rules

* Invalid sources fail construction before resolution and match `ErrInvalidSources` plus
  `apperr.ErrConfig`.
* Caller cancellation performs no new provider I/O, returns nil `*Resolved`, retains partial
  attribution, and matches both `ErrResolutionCanceled` and the exact context cause.
* Required-field exhaustion returns no partial credential aggregate. Optional `team_id`
  exhaustion is the only empty-value success.
* Internal timeout, unsupported platform, backend error, and busy source are `Unavailable` and
  continue. A true miss is distinct and does not set `AnyUnavailable`.
* A timed-out worker owns the semaphore until it returns; repeated calls do not create more
  workers. Buffered completion prevents a returning abandoned worker from blocking on send.
* The resolver's single-use mutex makes concurrent calls deterministic and race-safe.

### Risky actions

`strict_safety.enabled` is false. Adding `go-keyring` and reading an OS credential store are
moderate, reversible actions. There are no writes, migrations, destructive actions, or production
deployment changes. The real adapter is reachable only through explicit future use of
`DefaultSources`; P3 does not wire it into the current executable.

### Hardening verification

`GATE-SECRET` isolates the redaction primitive before credential work. The final suite adds
race-enabled concurrency tests, existing secret scanning, standard-library dependency containment,
CGO-disabled cross-compiles, vulnerability scanning, and operator-doc coupling. Any canary
appearance, live keychain test path, unbounded worker count, write-API reference, or config
credential field is a release blocker.

## Constitution Check

| Principle | Compliance |
|---|---|
| I Safety-First Go | Closed enums, explicit errors, no panic/unsafe, bounded concurrency |
| II Test-First | Every unit names its red test file and exact expected behavior |
| III/IV Workspace boundaries | No path expansion; oracle remains read-only; no secrets in artifacts |
| V Observability | Safe structured attribution is returned for the first real composition root; P3 remains deliberately unwired |
| VI Single Responsibility | One justified dependency; `secret`, `credentials`, and CLI composition remain separate |
| IX Git-friendly persistence | Plan, backlog, and review remain human-readable tracked artifacts |
| X Context efficiency | One canonical credential reference and one bounded hierarchy |

No constitutional violation is requested.

## Verification and Closure Requirements

Ship must run the existing gates only:

1. `gofmt -l .`
2. `goimports -l .`
3. `go vet ./...`
4. `go test -race ./...`
5. `go build ./...`
6. `golangci-lint run ./...`
7. `staticcheck ./...`
8. `govulncheck ./...`
9. `gitleaks detect --source . --no-git -v`
10. `go mod tidy` with no subsequent diff
11. existing CGO-disabled cross-compiles for linux/amd64, windows/amd64, darwin/amd64, darwin/arm64
12. the repository's existing local pre-commit Markdown/documentation checks (no new CI workflow)

Closure records the exact `go-keyring` version, cross-platform build results, provider
classification results, redaction matrix result, and the accepted risks above. Rollback is a
single merge revert; no schema, persisted state, keychain write, or migration exists.

## Harvest Shape

After a PASS review, harvest exactly one P3 feature. Each implementation unit above becomes one
task with one or two file/test subtasks. Every task is `XS`, `S`, or `M`,
`complexity: low|medium`, and
must retain the plan and deliberation references. Add a dependency from the new feature to
`003-F`, and link the new queued shipment to shipped predecessor `003-S`. Do not include any P4+
work from `4989A42D`.

## Gate Status

Revision 6 failed the final bounded escalation-route adversarial re-review. Harvest is forbidden.
The operator must choose whether to preserve exact oracle empty-ACL fallback and authorize a
targeted revision, or reopen deliberation to accept first-present ACL behavior as an explicit
divergence. The other P1s are mechanical but remain blocked behind that decision.

```text
dispatch_mode: multi-agent
decision: FAIL
```

## Plan Review — Escalation Attempt 1 (Revision 4)

`dispatch_mode: multi-agent`

`decision: FAIL`

Five independent passes completed: Security (`gpt-5.6-sol`), Go/Correctness
(`claude-sonnet-5`), Schema-CLI-Docs (`gemini-3.8-flash`), Architecture
(`claude-opus-5`), and Scope (`gpt-5.4-mini`). Consolidated result: 0 P0, 8 P1, 6 P2, 2 P3.

| ID | Severity | Finding | Revision-5 disposition |
|---|---|---|---|
| AR-01 | P1 | Revision 4 wired `RunE` despite accepted P3-D8(b) | Restored composed-but-unwired integration; removed all `cmd/intercom` production changes. |
| AR-02 | P1 | Resolver preceded the `Resolved` declaration | `D3` now precedes `D1a`/`D1b`; graph carries the compile edge. |
| AR-03 | P1 | Acceptance remained prose rather than exact tests | Every unit now names at most four test functions with exact inputs/outcomes. |
| AR-04 | P1 | External doc test could not read runtime identifiers | Coupling test is in-package and derives identifiers from production chain data. |
| AR-06 | P1 | Required hardening section absent | Added full `## Plan Hardening` threat, lifecycle, risk, and verification contract. |
| AR-07 | P1 | Formatter-removal control could pass | Direct verbs require exact markers; unformatted fixture must fail those exact expectations. |
| AR-08 | P1 | Cancellation lost `context` identity | `resolveError` has an optional third unwrap child and exact cancellation/deadline tests. |
| AR-09 | P1 | Pre-canceled lookup could still start keychain I/O | Added cancellation checks before and after semaphore acquisition plus zero-call assertion. |
| AR-05 | P2 | Dependency graph omitted compile/test-first edges | Rebuilt graph and serial order, including doc harnesses before docs. |
| AR-10 | P2 | `stringlist_test.go` package unspecified | Declared `package secret`. |
| AR-11 | P2 | Verification omitted real gates | Added `goimports`, `staticcheck`, `gitleaks`; correctly labeled Markdown checks local. |
| AR-12 | P2 | Tests could still call `DefaultSources` | AST guard now rejects it and production keychain constructors from tests. |
| AR-13 | P2 | Unavailable warning behavior unspecified | P3 remains unwired; P8 is named as the first consumer and must log one safe WARN. |
| AR-14 | P2 | ACL environment trust boundary absent | Canonical docs now require service-manager/environment-file trust guidance. |
| AR-15 | P3 | `NewResolver` error contract unspecified | Added `ErrInvalidSources`, nil-resolver rule, and config-kind matching. |
| AR-16 | P3 | AST walk boundary unspecified | Added module-root discovery and explicit owned-tree exclusions. |

## Plan Review — Escalation Attempt 2 (Revision 5)

`dispatch_mode: multi-agent`

`decision: FAIL`

Five independent passes completed: Security (`gpt-5.6-sol`), Go/Correctness
(`claude-sonnet-5`), Architecture (`claude-opus-5`), Scope (`gemini-3.8-flash`), and
Schema-CLI-Docs (`grok-4.6`). Consolidated result: 0 P0, 4 P1, 12 P2, 3 P3.

| ID | Severity | Finding | Revision-6 disposition |
|---|---|---|---|
| AR5-01 | P1 | Serial order placed D1b before its D2 dependency | Reordered D2 before D1b and added a topological-order contract test. |
| AR5-02 | P1 | External ACL test could not reach the unexported resolver | Made it in-package with a local counting source; no `credtest` cycle. |
| AR5-03 | P1 | Present-empty ACP ACL incorrectly fell through | Pinned first-present semantics: only absence falls back; present-empty fails without shared lookup. |
| AR5-04 | P1 | Cancellation was not tested through `Resolver.Resolve` | Added resolver-level cancel/deadline rows, exact identities/attribution, no-later-call and zero-keychain-I/O assertions. |
| AR5-05 | P2 | Source assembly and constructor contract lacked an owner | Added C6 with exact default, nil-source, and single-use tests. |
| AR5-06 | P2 | Keychain constructors/timeout seam unnamed | Added exact production/test constructors and guard identifiers. |
| AR5-07 | P2 | Tidy would remove the dependency before its import | Split pin, red harness, and adapter implementation; tidy runs only after the import exists. |
| AR5-08 | P2 | `source_test.go` and `chain_test.go` packages absent | Declared both as in-package tests. |
| AR5-09 | P2 | Docs test still rebuilt identifiers | It now enumerates actual production refs/constants; cardinality is only a check. |
| AR5-10 | P2 | `Resolved` exported mutable credential fields | All fields are unexported with value/copy accessors and ownership tests. |
| AR5-11 | P2 | Attribution slices externally mutable | Closed `Field`, unexported storage, nested deep-copy accessors, isolation/log tests. |
| AR5-12 | P2 | Graph omitted compile/test edges | Added S1→C5, C2→D2, C5→E2, C6, and guard edges; rebuilt serial order. |
| AR5-13 | P2 | Credential field identity had competing owners | C1 now owns closed `Field`; chain and docs consume it. |
| AR5-14 | P2 | CLI flag documentation was unrelated scope | Removed E6b and the new flag-table requirement. |
| AR5-15 | P2 | Unknown-enum rejection was unreachable | Removed the unreachable branch; constructors make states closed. |
| AR5-16 | P2 | Repository guard was negatively bounded | Moved to `tests/arch` and positively enumerated owned Go roots. |
| AR5-17 | P3 | Scope omitted config boundary test | Added `internal/config` test-only scope. |
| AR5-18 | P3 | Plaintext accessor sites were not bounded | Named and guard-enforced the three permitted P3 call sites. |
| AR5-19 | P3 | Shared semaphore cross-tier busy behavior unstated | Documented and added exact timeout-then-busy test. |

## Plan Review — Escalation Attempt 3 (Revision 6, Final Bounded Gate)

`dispatch_mode: multi-agent`

`decision: FAIL`

Five independent passes completed: Security (`gpt-5.6-sol`), Go/Correctness
(`claude-sonnet-5`), Architecture (`claude-opus-5`), Scope (`gemini-3.8-flash`), and
Schema-CLI-Docs (`grok-4.6`). Consolidated result: 1 P0, 3 P1, 6 P2, 2 P3.

| ID | Severity | Finding | Required resolution |
|---|---|---|---|
| AR6-01 | P0 | Revision 6's present-empty ACP ACL rule contradicts oracle `config.rs:489-493`, which filters empty ACP before shared fallback | Operator decision: restore exact empty-as-absent fallback, or reopen deliberation and record first-present behavior as a divergence. |
| AR6-02 | P1 | C6 calls `Resolve` before D1b/D3 exist and `sources_test.go` lacks package placement | Restrict C6 to construction/nil validation; move single-use execution to D1b; declare `package credentials`. |
| AR6-03 | P1 | Cancel-during-lookup and deadline-during-lookup orchestration remain unspecified | Define `CanceledResult` as `Unavailable/ReasonCanceled`, make it an unconditional abort, check context after lookups/before success, and add resolver-level race tests. |
| AR6-04 | P1 | Claimed total fmt matrix omits standard verbs and representative flags/width/precision/index forms | Expand the table to every standard operand-consuming verb and representative formatting forms for values/pointers. |
| AR6-05 | P2 | One timed-out worker suppresses later fields' keychain attempts | Document as a timeout divergence and test later-field env fallback, or redesign the worker bound. |
| AR6-06 | P2 | `Resolved` JSON map-key behavior is unspecified without `MarshalText` | Specify fail-closed `UnsupportedTypeError` with no canary, or add a redacting text marshaler. |
| AR6-07 | P2 | `go mod download` does not pin the module | Use `go get ...@v0.2.8`, verify exact module diff, defer tidy until import exists. |
| AR6-08 | P2 | `AuthorizedUserIDs()` is an unguarded plaintext boundary | Mark/rename it as explicit exposure and add accessor containment. |
| AR6-09 | P2 | Table order/graph/test topology still conflict | Reorder D2 before D1b in the table, add missing edges, and remove plan-ID topology from permanent product tests. |
| AR6-10 | P2 | E5a/E5b docs are not covered by E6a | Extend the red doc harness across README, example, and config reference. |
| AR6-11 | P3 | New package dependency directions are absent from architecture docs | Add a bounded domain-map update. |
| AR6-12 | P3 | Credentials package-ordering documentation has no file owner | Assign it to `sources.go` or a dedicated `doc.go` unit. |

This is the second consecutive Stage failure (normal route, then configured escalation route).
Per the operator's stop condition, no further revision, harvest, shipment, or stash archival is
authorized in this session.
