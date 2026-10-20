# intercom

## Configuration

`intercom` uses `config.toml` (override with `--config <path>`) as its
configuration format, `internal/config` implements the schema, defaults,
and validation. See `config.toml.example` for a complete, loadable example
covering every configuration key, and `docs/config-reference.md` for the
full key reference, default values, and validation rules enforced at load
time. The configuration is loaded and validated at process start by
`cmd/intercom` (wired in 062-S).

## Running intercom

`intercom --workspace <dir> --config <config.toml> --port <port>` starts one
Copilot session for one workspace, with a local terminal UI and a loopback-only
WebSocket listener for authenticated remote surfaces. Credentials for remote
clients are shown once on the local TUI. The operator runbook covers starting,
the `[remote]` keys, credential handling, exposing the port with a Microsoft
Dev Tunnel, limits, shutdown and troubleshooting:
[`docs/runbooks/intercom-operator-runbook.md`](docs/runbooks/intercom-operator-runbook.md).
Tests against a real Copilot runtime are opt-in; see
[`docs/runbooks/live-sdk-tests.md`](docs/runbooks/live-sdk-tests.md).
