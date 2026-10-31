# intercom operator runbook (MVP)

How to start `intercom`, hand out remote credentials safely, expose the
loopback port through a Microsoft Dev Tunnel, and shut down. Design
references are to
[`docs/design-docs/intercom-go-backend-architecture.md`](../design-docs/intercom-go-backend-architecture.md);
the controls are decided in
[ADR MVP-1](../decisions/2026-10-19-intercom-go-mvp-adr-1-scope-and-session.md),
[ADR MVP-2](../decisions/2026-10-19-intercom-go-mvp-adr-2-credentials-and-origin.md)
and [ADR MVP-3](../decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md).

## 1. What you are running

One `intercom` process serves **one** Copilot session for **one** workspace. It
owns a local terminal UI (full control, no credential) and a loopback-only
WebSocket listener for remote surfaces (an observer role that watches, an
approver role that can prompt, abort and answer permission requests).
Remote access is reached only through a tunnel you create yourself.

## 2. Start

```sh
intercom --workspace /path/to/workspace --config /path/to/config.toml --port 3000
```

PowerShell:

```powershell
intercom --workspace C:\src\my-workspace --config C:\etc\intercom\config.toml --port 3000
```

From a checkout, `go run ./cmd/intercom --workspace ... --config ... --port ...`
accepts the same flags. These are all of them (`intercom --help`):

| Flag | Default | Meaning |
|---|---|---|
| `--config` | `config.toml` | Path of the config file. |
| `--workspace` | the config's `default_workspace_root` | Workspace directory; it overrides `default_workspace_root`. It must exist and lie inside an authorized root (`default_workspace_root` or a `[[workspace]].path`, checked through `pathsafe`, which resolves symlinks): the root itself or any subdirectory of it is accepted; anything outside every root stops startup. |
| `--port` | the config's `http_port` (3000) | Overrides `http_port`. Loopback TCP port for the WebSocket listener. `--port 0` picks a free port (the chosen address is in the `intercom ready` log line and on the credential panel). |
| `--log-level` | `info` | `debug`, `info`, `warn` or `error`. |

The process then starts, in order: config and workspace checks, the hub, the
bound (not yet serving) listener, the Copilot SDK client (`Start` is the
readiness gate; a missing, too-old or unauthenticated CLI fails here, before
anything is exposed), the listener, the credentials, and the local TUI. Any
failure unwinds everything already started and exits non-zero.

Prerequisites: the Copilot CLI installed and authenticated for the user that
runs `intercom`, a `config.toml` with `default_workspace_root` (see
`config.toml.example`), and an interactive terminal: the local TUI draws the
credential panel on stdout, so `intercom` refuses to start when stdout is
redirected to a file or pipe (a credential must never be written to one).

### Headless and service use

The command-line binary is TUI-only. ADR MVP-2 forbids delivering a bearer
credential to stdout pipes, logs or files, so there is no flag that disables the
TUI, and a non-interactive launch (a service, a scheduled task, a redirected
stdout) is refused at start with "stdout is not a terminal". This is a decision,
not a gap.

* In-process, programmatic headless operation is supported: a Go program
  that embeds `internal/app` sets `app.TUIOptions{Disabled: true}`, then owns
  credential delivery itself. The `intercom` binary does not expose it.
* **Unattended service mode** for the binary waits on identity-provider
  authentication (ADR MVP-2, option b), which is post-MVP.

## 3. Configure `[remote]`

Secure defaults: no browser origin and no extra host is allowed, and remote
approval of destructive requests is off. **Never put a credential in
`config.toml`**: any credential-shaped key under `[remote]` fails validation.

| Key | Default | Meaning |
|---|---|---|
| `token_lifetime_seconds` | `14400` (4 h) | Fixed lifetime of a minted credential, `60` to `86400`. No sliding expiry and no refresh: at expiry the credential is dead and the socket is closed. |
| `allowed_origins` | `[]` | Exact, canonical, lowercase `http(s)://host[:port]` origins a **browser** may connect from (default port omitted; no wildcard, path or userinfo). Empty means every browser origin gets 403 even with a valid credential. Native clients send no `Origin` and are unaffected. |
| `allowed_hosts` | `[]` | Extra `Host` header values accepted besides the loopback names for the bound port: `host[:port]`, lowercase, no wildcard or scheme. A tunnel deployment must list the tunnel hostname here or every request is refused 403. An entry without a port matches the default port only. |
| `max_connections` | `8` | Cap on concurrent remote connections, `1` to `1024`. Over the cap gets 503 before any credential work. |
| `remote_approve_destructive` | `false` | When `false`, a remote approver can approve only `read` requests: approving shell, write, URL, other or unclassified kinds is refused (`destructive_remote_refused`). When `true`, a remote approver may also approve `shell`, `write` and `url`; `other` and unclassified (unknown) kinds are always local-TUI-only and stay refused. Denying is always allowed. The local TUI can approve everything. |

Related keys that matter to operators: `http_port` (default 3000),
`copilot.cli_path` (see the PATH residual below), `copilot.startup_timeout_seconds`
(default 30) and `timeouts.approval_seconds` (default 3600, at most 86400:
how long a permission request waits before the hub resolves it "user not
available").
Full reference: [`docs/config-reference.md`](../config-reference.md).

## 4. Credentials

At start the process mints two independent 256-bit bearer credentials, one per
role (`approver` and `observer`).

* They are **shown once, on the local TUI only**, in a panel with the WebSocket
  URL, each role, a short fingerprint and the expiry. Copy them then. Press
  **Esc** to dismiss the panel; the TUI model copy of the plaintext is dropped
  when you dismiss it. The process drops its own stored reference to the
  plaintext when it hands the credentials to the TUI, and again on re-issue,
  revoke and shutdown. Only the fingerprint and expiry stay stored.
* **Ctrl+R** re-issues: new credentials are minted and shown, the old ones are
  revoked immediately and sockets that authenticated with them are closed. This
  is also the renewal path (ADR MVP-2): a credential cannot be extended, so
  before `token_lifetime_seconds` runs out, re-issue and hand the new value to
  the remote user out of band.
* They are never in `config.toml`, a URL, the command line, a hub envelope, a
  hydrate payload, a log line or on disk. Logs carry only the role, the
  fingerprint and the expiry.
* A credential is a **bearer secret**: whoever holds it within its lifetime can
  act in that role. Treat it like a password and re-issue if it leaks.

Clients present the credential in the `Authorization: Bearer <credential>`
request header (native clients, preferred) or, from a browser, as the
WebSocket subprotocol pair `intercom.v1, bearer.<credential>`. The server
selects only `intercom.v1` and never echoes the credential entry. Every
credential failure (missing, malformed, wrong, expired, revoked) is the same
bare HTTP 401. The reference client and the complete wire protocol are in
`internal/transport/wsclient`.

**Accepted residual: subprotocol logging.** A browser can only carry the
credential in `Sec-WebSocket-Protocol`, and a tunnel or edge proxy may record
that request header in its own logs. This is accepted for the MVP (ADR MVP-2);
the mitigations are the bounded lifetime, re-issue, and using the
`Authorization` header from any client that can set one. If you suspect the edge
logged a credential, re-issue with Ctrl+R.

## 5. Expose the port with a Microsoft Dev Tunnel

Creating and hosting the tunnel is a **manual step outside intercom**: the
process cannot see or verify the tunnel, and keeping anonymous access off is
**your responsibility**. (The backend still authenticates every upgrade with
the credential above, so a mistake does not expose an unauthenticated agent, but
do not rely on that.)

```sh
devtunnel user login
devtunnel create intercom-demo --allow-anonymous false
devtunnel port create intercom-demo -p 3000 --protocol http
devtunnel host intercom-demo
```

Use the port you passed to `intercom`. Confirm the tunnel is private with
`devtunnel show intercom-demo` (no anonymous access entry). Tunnels are
private by default. The explicit `--allow-anonymous false` syntax was accepted
by devtunnel CLI 1.0.2014 on 2026-10-20 (the tunnel's access control was `{}`,
with no anonymous entry). If another build rejects the explicit `false`, omit it
and verify the same way.

Then:

1. Add the tunnel hostname that `devtunnel host` prints (for example
   `intercom-demo-3000.euw.devtunnels.ms`, no scheme) to `[remote].allowed_hosts`
   and restart. Without it every tunnelled request is refused with 403.
2. Connect clients to `wss://<tunnel-host>/ws` (plain `ws://` is accepted only
   for loopback).
3. A browser SPA also needs its exact origin in `[remote].allowed_origins`.
4. A private tunnel also requires its clients to present a Dev Tunnels access
   token (edge authentication). That is separate from the intercom credential.
   Obtain one with `devtunnel token <tunnel-id> --scopes connect` (about 24
   hours of lifetime) and send it in the `X-Tunnel-Authorization` header as
   `tunnel <token>`. The intercom credential still travels in `Authorization`.
   A connection without the edge token is refused at the edge: it receives the
   Dev Tunnels sign-in page (HTTP 200) and no WebSocket upgrade, and never
   reaches intercom. Treat the edge token like a credential and keep it out of
   logs and the repository.
5. ADR MVP-2 asks for a manual, per-environment check that the `Authorization`
   header reaches intercom unmodified through the tunnel. It is a per-environment
   check, so repeat it for yours; if the header does not arrive, native clients
   need another carriage and the ADR must be amended. See the verified result
   below.

### Verified on 2026-10-20

Scope: devtunnel CLI 1.0.2014, one user account, one machine. The backend ran
behind a private tunnel (host pattern `<id>-<port>.<cluster>.devtunnels.ms`)
with `[remote].allowed_hosts` set to the bare tunnel hostname (no port, which
worked). The client was the Go reference client (`internal/transport/wsclient`)
over `wss://<tunnel-host>/ws` from the same machine through the public edge. The
backend was a temporary harness that wires the real Copilot agent, hub and
transport, because the real binary shows credentials only on its interactive TUI
and could not be driven non-interactively (see the closure evidence). The tunnel
and the harness were deleted afterwards.

* Without the edge token the connection was refused at the edge, so the tunnel
  is private.
* With the edge token and a wrong intercom credential, intercom answered 401.
* An observer and an approver each connected with the credential in the
  `Authorization` header: the header reached intercom unmodified.
* The browser-style carriage (`intercom.v1` plus `bearer.<credential>` in
  `Sec-WebSocket-Protocol`) connected through the tunnel, tested with the
  reference client only.
* A full turn completed: the prompt was accepted, a `read` permission was
  requested, a remote approver approved it, `permission.resolved` and
  `agent.idle` were observed, and shutdown was clean.

Not tested: a real browser, a second machine, the real binary (launched from its
TUI) behind the tunnel, `Origin` handling and `allowed_origins` through the
tunnel, a long-lived connection with reconnect across the tunnel, and permission
kinds other than `read` (such as shell or write) routed live.

## 6. Limits and what to expect

* **One session per process, no persistence.** The transcript lives in memory
  and is gone when the process exits; a restart is a new session and old
  credentials and cursors are dead.
* **Loopback only.** The listener binds `127.0.0.1`; exposure is only via the
  tunnel.
* **Observers see everything.** An observer credential sees the full transcript
  and every permission request's text, but cannot prompt, abort or approve.
* **Connection cap** (`max_connections`, default 8). The cap also counts
  not-yet-authenticated sockets, so someone who can reach the tunnel can hold
  slots for the handshake window (10 s) and delay legitimate reconnects. That
  affects the remote surface only, never the local TUI or the agent.
* **Message limits.** A client frame is at most 16 KiB and must be a single
  JSON object; request headers are limited to 16 KiB; a handshake must finish in
  10 s. Violations close the socket (1008 or 1009).
* **Idle deadline.** The server pings every 30 s; a connection that stops
  answering for 90 s is closed. A reconnect resumes from the client's `(gen, seq)`
  cursor without gaps or duplicates, or re-hydrates when the ring no longer
  covers it.
* **Rate limit.** 5 actions per second per connection, burst 10; a connection
  that keeps exceeding it (20 consecutive refusals) is closed.
* **Destructive approvals are refused remotely by default** (section 3). A
  request nobody answers is resolved "user not available" after
  `timeouts.approval_seconds`, and immediately if no surface is connected.
* **Bare-name CLI PATH residual.** With `copilot.cli_path = ""` or a bare name
  such as `copilot`, the CLI is found on `PATH`, so anyone who can write an
  earlier `PATH` entry can substitute a program that runs as you. Set an
  absolute path for anything beyond a trusted workstation.

## 7. Logs

Structured JSON logs go to **stderr**, except that the local TUI owns the
terminal's alternate screen: while the TUI is enabled (always, in this build)
and stderr is a terminal, the process logger discards its output rather than
corrupt the display. To keep the logs, redirect stderr; they are then written
there as JSON lines:

```sh
intercom --workspace /path/to/workspace --config config.toml 2> intercom.log
```

```powershell
intercom --workspace C:\src\my-workspace --config config.toml 2> intercom.log
```

Fatal startup and exit errors go to stderr, so to the log file when stderr is
redirected, and are still printed to the terminal after the TUI has closed
when stderr is a terminal. Raise detail with `--log-level debug`. Logs never
contain credentials, only roles, fingerprints and expiry times.

### Rejected upgrades on the status line

Because the logger is discarded while the TUI owns the terminal, the TUI status
line reports rejected `/ws` WebSocket upgrades itself, so a refused remote client
is visible without a log:

```text
rejected upgrades: N (last: STATUS REASON)
```

`N` is the running count of rejected upgrades since start and `STATUS REASON` is
the HTTP status and reason of the most recent one (400, 401, 403 or 503, see the
troubleshooting table in section 9). The counter covers handler-level upgrade
refusals only: 400 bad request, 401 credential, 403 host or origin, and the 503
for a draining process or a hub refusal. It does not count over-cap 503s from
the connection limiter (`max_connections`) or 404s for a wrong path. Handshake
refusals generated by the websocket library inside Accept (for example an
unsupported non-empty `Sec-WebSocket-Version`) are also not counted. Updates are
coalesced, so a burst of rejections shows the latest count, not one line per
request. Nothing is shown until the first rejection. The line never carries a
credential.

### Unconsumed config keys

Some config keys are accepted and validated but no component reads them yet.
When any of them is set to a non-default value, startup logs one warning,
`config has accepted-but-unconsumed keys`, with the key names (never the values)
in sorted order. The setting has no effect. The keys that can appear are:

* `max_concurrent_sessions`
* `retention_days`
* `operator_detail_level`
* `database.path`
* `timeouts.prompt_seconds`
* `timeouts.wait_seconds`
* `stall.enabled`
* `stall.inactivity_threshold_seconds`
* `stall.escalation_threshold_seconds`
* `stall.max_retries`
* `stall.default_nudge_message`
* `commands`

The warning sits next to the existing `config warning` and `config has unknown
keys` lines, so it is visible only when stderr is redirected (above).

## 8. Shutdown

Quit with **Ctrl+C in the TUI**, or send `SIGINT`/`SIGTERM` to the process. Either
runs the ordered shutdown of design section 3.1:

1. every pending permission is resolved "user not available" (so the agent can
   finish), operator actions stop being accepted, credentials are revoked and
   the listener closes;
2. `server.shutdown` is broadcast to every surface;
3. the running turn is aborted and the agent is stopped, which terminates the
   Copilot CLI child process;
4. client sockets are drained briefly and closed;
5. the hub stops last.

Stopping the local TUI is bounded. The process asks it to quit and waits up to
3 s; if it has not left, it is killed and waited for another 3 s. A TUI that
still does not exit is reported as an error (`local tui: the program did not exit
within ...`) and surfaces in the process's exit error instead of hanging
shutdown. A TUI that fails on its own also starts a shutdown, and its error is
part of the exit error. A signal received during startup cancels the startup
rather than waiting for it, and a failure to fail pending permissions during
teardown is reported in the exit error.

A **second** signal hurries a stuck stop but does not skip it; an interrupted or
unconfirmed shutdown exits non-zero. After exit no CLI child should remain
(the live test in section 10 checks exactly this); if the log reports a cleanup
problem, look for a stray `copilot` process.

## 9. Troubleshooting

| Symptom | Likely cause |
|---|---|
| Exits at start with a config error | Fix the named key; credential-shaped keys under `[remote]` are rejected. |
| Exits at start with an agent start error | CLI missing, older than the validated version, not authenticated, or slower than `copilot.startup_timeout_seconds`. |
| `bind` error | The port is in use: pick another with `--port`. |
| Remote client gets 401 | Missing, wrong, expired or re-issued credential; ask for the current one. |
| Remote client gets 403 | `Host` not allowed (add the tunnel host to `allowed_hosts`) or browser `Origin` not in `allowed_origins`. |
| Remote client gets 503 | Over `max_connections`, or the process is shutting down. If the status line shows no rejections, check `[remote].max_connections`: over-cap 503s are not counted there (section 7). |
| Approve is refused remotely | `destructive_remote_refused`: approve on the local TUI, or (for `shell`, `write`, `url` only) set `remote_approve_destructive = true` knowingly. Unknown or unmodeled kinds can only be approved on the local TUI, whatever the setting. |
| No logs appear | The TUI is enabled and stderr is a terminal, so logs are discarded; redirect stderr to keep them (section 7). |
| Exits at start: "stdout is not a terminal" | `intercom` was run with stdout redirected or without a terminal; run it in an interactive terminal (section 2, headless and service use). |
| Status line shows `rejected upgrades: N (last: STATUS REASON)` | Remote clients are being refused; map the last status with the 401, 403 and 503 rows above (400 is a request that is not a valid upgrade or carries a bad resume cursor) (section 7). |
| Startup log has `config has accepted-but-unconsumed keys` | A listed key is set but not used yet; it has no effect (section 7). |

## 10. Verification and live tests

The normal test suite includes a headless end-to-end scenario
(`go test -race ./internal/app -run E2E`). Tests against a **real** Copilot
runtime are opt-in:

```sh
INTERCOM_LIVE_SDK_TESTS=1 go test ./internal/app -run Live -v
```

```powershell
$env:INTERCOM_LIVE_SDK_TESTS = "1"; go test ./internal/app -run Live -v
```

Prerequisites, what they cover, the stranded-child check and the SQ-b/SQ-c
re-verification are in [live-sdk-tests.md](live-sdk-tests.md).

The live tests were first run on 2026-10-20 (Copilot CLI 1.0.100, `GH_TOKEN`
authentication, Windows, go1.26.9). `TestLiveCopilotEndToEnd` and
`TestLiveStartAndStop` passed, and SQ-b was verified. SQ-c was **not**
re-verified, so the conservative posture of design section 4.4 stays in force.
The results and the open SQ-c question are in
[live-sdk-tests.md](live-sdk-tests.md#results-of-the-first-live-run-2026-10-20).

## 11. Known residuals

* Credential plaintext may appear in a tunnel or edge log when carried as a
  browser subprotocol (section 4). Accepted.
* Bearer credentials can be replayed within their lifetime; single-use upgrade
  tickets, channel binding, identity-provider validation and automatic rotation
  are deferred (ADR MVP-2).
* The unauthenticated connection cap can be held by a party that reaches the
  tunnel (section 6). Accepted.
* The `Authorization` header reached intercom unmodified through a private
  Dev Tunnel in one environment on 2026-10-20 (section 5). Repeat the check per
  environment; a real browser, a second machine and long-lived reconnects across
  the tunnel are untested.
* Design section 4.4 SQ-c (events keep dispatching while a permission handler is
  parked) is unverified; the conservative posture stays in force (section 10).
* A hard-killed intercom (SIGKILL, power loss) skips the ordered shutdown and
  can leave the CLI child running; use the process list to clean up.
* The live SDK tests ran once against a real runtime (section 10); only the
  `read` permission kind has been exercised live. Shell and write kinds have not.
