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
| `remote_approve_destructive` | `false` | When `false`, a remote approver can approve only `read` requests: approving shell, write, URL, other or unclassified kinds is refused (`destructive_remote_refused`). Denying is always allowed. The local TUI can approve everything. |

Related keys that matter to operators: `http_port` (default 3000),
`copilot.cli_path` (see the PATH residual below), `copilot.startup_timeout_seconds`
(default 30) and `timeouts.approval_seconds` (default 3600: how long a
permission request waits before the hub resolves it "user not available").
Full reference: [`docs/config-reference.md`](../config-reference.md).

## 4. Credentials

At start the process mints two independent 256-bit bearer credentials, one per
role (`approver` and `observer`).

* They are **shown once, on the local TUI only**, in a panel with the WebSocket
  URL, each role, a short fingerprint and the expiry. Copy them then. Press
  **Esc** to dismiss the panel; the plaintext is dropped from the UI.
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
private by default; if your `devtunnel` version rejects the explicit `false`,
omit it and verify the same way. The `--allow-anonymous false` syntax is
unverified against a real devtunnel build.

Then:

1. Add the tunnel hostname that `devtunnel host` prints (for example
   `intercom-demo-3000.euw.devtunnels.ms`, no scheme) to `[remote].allowed_hosts`
   and restart. Without it every tunnelled request is refused with 403.
2. Connect clients to `wss://<tunnel-host>/ws` (plain `ws://` is accepted only
   for loopback).
3. A browser SPA also needs its exact origin in `[remote].allowed_origins`.
4. A private tunnel also asks its clients for a Dev Tunnels access token
   (edge authentication). That is separate from the intercom credential. Whether
   the `Authorization` header reaches intercom unmodified through your tunnel is
   a manual, per-environment check that this build has not performed (ADR MVP-2
   asks for it): if it does not, native clients need another carriage and the ADR
   must be amended.

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

Fatal startup and exit errors are still printed to the terminal after the TUI
has closed. Raise detail with `--log-level debug`. Logs never contain
credentials, only roles, fingerprints and expiry times.

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
| Remote client gets 503 | Over `max_connections`, or the process is shutting down. |
| Approve is refused remotely | `destructive_remote_refused`: approve on the local TUI, or set `remote_approve_destructive = true` knowingly. |
| No logs appear | The TUI is enabled and stderr is a terminal, so logs are discarded; redirect stderr to keep them (section 7). |
| Exits at start: "stdout is not a terminal" | `intercom` was run with stdout redirected; run it in an interactive terminal (section 2). |

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

## 11. Known residuals

* Credential plaintext may appear in a tunnel or edge log when carried as a
  browser subprotocol (section 4). Accepted.
* Bearer credentials can be replayed within their lifetime; single-use upgrade
  tickets, channel binding, identity-provider validation and automatic rotation
  are deferred (ADR MVP-2).
* The unauthenticated connection cap can be held by a party that reaches the
  tunnel (section 6). Accepted.
* Whether the tunnel's edge authentication forwards the `Authorization` header
  unmodified is unverified (section 5).
* A hard-killed intercom (SIGKILL, power loss) skips the ordered shutdown and
  can leave the CLI child running; use the process list to clean up.
* The live SDK tests have not yet been run against a real runtime.
