# localvault

A Doppler-like local secrets manager backed by your OS keyring.

Secrets are stored encrypted by the operating system (Secret Service on Linux, Keychain on macOS, Credential Manager on Windows). They never touch disk in plaintext. No cloud, no server, no network.

## Install

```sh
go install github.com/xilistudios/localvault@latest
```

Or build from source:

```sh
git clone https://github.com/xilistudios/localvault.git
cd localvault
make install
```

## Quick Start

```sh
# Initialize
localvault setup

# Create project and environment
localvault projects create myapp
localvault configure --project myapp
localvault configs create dev
localvault configure --config dev

# Store secrets
localvault set DATABASE_URL=postgres://localhost:5432/mydb
localvault set API_KEY=sk-abc123
localvault set PASSWORD              # prompts for hidden input

# List secrets (masked by default)
localvault secrets
#   API_KEY      = sk*****23
#   DATABASE_URL = po**************************db
#   PASSWORD     = ****

# Get a single value
localvault get DATABASE_URL --plain

# Run a command with secrets injected as env vars
localvault run -- node server.js

# Import from .env file
localvault import .env

# Export
localvault export --format json
```

## Commands

```
localvault setup                          Initialize ~/.localvault/
localvault status                         Show vault health + active scope
localvault configure                      Set active project/config

localvault projects                       List projects
localvault projects create <name>         Create project
localvault projects delete <name>         Delete project + all secrets
localvault projects info <name>           Show project details

localvault configs                        List configs
localvault configs create <name>          Create config
localvault configs delete <name>          Delete config + secrets
localvault configs copy <src> <dst>       Copy secrets between configs

localvault set KEY=VALUE [...]            Set secrets
localvault get KEY [--plain]              Get secret value
localvault secrets [--values]             List secrets (masked by default)
localvault unset KEY [...]                Delete secrets

localvault run -- <cmd> [args]            Run with secrets as env vars
localvault import [file | --stdin]        Import from .env
localvault export [--format F]            Export (dotenv, json, docker)
```

### Global Flags

```
-p, --project string    Override active project
-c, --config string     Override active config
-o, --output string     Output format (default "table")
    --no-color          Disable colored output
-v, --verbose           Verbose output
```

## Storage

| What | Where | Encrypted by |
|------|-------|-------------|
| Secret values | OS keyring (one entry per secret) | OS (AES-256 via Secret Service / Keychain / WinCred) |
| Metadata (project/config names, key index) | `~/.localvault/vault.json` | Filesystem permissions (0600) |

Keyring backends (auto-detected in order):
1. Secret Service (Linux desktop — GNOME/KDE)
2. Keychain (macOS)
3. Credential Manager (Windows)
4. KWallet (Linux/KDE fallback)
5. Encrypted file (`~/.localvault/keyring/`) — headless/CI fallback

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `LOCALVAULT_DIR` | Override vault directory (default `~/.localvault`) |
| `LOCALVAULT_PASSPHRASE` | Passphrase for file-based keyring backend |

## Naming Rules

- Projects/configs: `[a-z0-9_-]`, 1–64 chars
- Secret keys: `[A-Z0-9_]`, 1–256 chars
- Secret values: any UTF-8, max 64KB

## Development

```sh
make build      # Build binary to bin/
make test       # Run all tests
make lint       # golangci-lint
make fmt        # gofmt + goimports
make ci         # fmt + lint + test + build
```

## Architecture

```
cmd/               Cobra commands (thin CLI layer)
internal/model     VaultFile, Project, Config structs + validation
internal/vault     Vault orchestrator, keyring adapter, metadata, scope
internal/envfile   .env parser + exporter (dotenv/json/docker)
```

## License

MIT
