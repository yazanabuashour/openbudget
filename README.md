# openbudget

`openbudget` is now bootstrapped as a Go module with a minimal CLI entrypoint, local SQLite runtime foundations, and CI checks. The repository is still early-stage scaffolding and does not publish runnable artifacts yet.

## Repository contents

- [CONTRIBUTING.md](CONTRIBUTING.md) explains how outside contributors should propose changes.
- [SECURITY.md](SECURITY.md) explains how to report vulnerabilities privately and what response timing to expect.
- [docs/maintainers.md](docs/maintainers.md) documents Beads-based maintainer workflow and repo administration notes.
- [cmd/openbudget](cmd/openbudget) contains the initial CLI entrypoint.
- [internal/app](internal/app) contains the first internal package and unit tests.
- [internal/localruntime](internal/localruntime) resolves the local database path and opens the internal SQLite runtime.
- [internal/storage/sqlite](internal/storage/sqlite) contains the internal SQLite migration and config repository foundation.
- [internal/agenteval](internal/agenteval) contains helpers for isolated Codex eval homes.
- [LICENSE](LICENSE) defines the project license.

## Release contract

The initial release surface is GitHub Releases with semantic version tags in the `0.y.z` range. Release notes are generated from protected tags. This repository does not currently publish packages or downloadable build artifacts.

## Development

Install the pinned toolchain with:

```bash
mise install
```

Run the CLI with:

```bash
mise exec -- go run ./cmd/openbudget
```

The default SQLite path is `${XDG_DATA_HOME:-~/.local/share}/openbudget/openbudget.db`.
Override it with `OPENBUDGET_DATABASE_PATH` or `localruntime.Config{DatabasePath: "..."}`.
`OPENBUDGET_DATA_DIR`, `DATA_DIR`, and data-directory config fields are not supported.
Runtime settings loaded after SQLite opens should be stored in the database; the database path itself must remain outside SQLite because the runtime needs the file path before it can open the database.

Run the local quality gates with:

```bash
make check
```

## Contributing

Outside contributors can work entirely through GitHub issues and pull requests. Beads is maintainer-only workflow tooling and is not required for community contributions.

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution expectations and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for community standards.
