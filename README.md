# intercom

## Configuration

`intercom` uses `config.toml` (override with `--config <path>`) as its
configuration format, `internal/config` implements the schema, defaults,
and validation. See `config.toml.example` for a complete, loadable example
covering every configuration key, and `docs/config-reference.md` for the
full key reference, default values, and validation rules enforced at load
time. Config loading is not yet wired into the running server; that lands
in a later phase.
