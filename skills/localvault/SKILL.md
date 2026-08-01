---
name: localvault
description: "Manage secrets using the localvault CLI (Doppler-like, OS keyring backed). Use when storing, retrieving, listing, importing, or exporting secrets for projects. Trigger on 'secret', 'localvault', 'vault', 'env vars', 'set secret', 'get secret', 'import .env', 'export secrets', 'run with secrets'."
---

# LocalVault

CLI secrets manager backed by the OS keyring. Secrets are encrypted at rest by the OS (Secret Service / Keychain / Credential Manager). Metadata lives in `~/.localvault/vault.json`.

## Concepts

- **Project** — named collection of configs (e.g. `myapp`)
- **Config** — environment within a project (e.g. `dev`, `prd`)
- **Scope** — active project+config pair, resolved from flags or `configure`

## Setup

```bash
localvault setup                              # once, creates ~/.localvault/
localvault projects create <name>
localvault configure --project <name>
localvault configs create <env>
localvault configure --config <env>
```

## Core Operations

```bash
# Secrets
localvault set KEY=VALUE [KEY2=V2 ...]        # set (hidden prompt if no =)
localvault get KEY [--plain]                  # get value
localvault secrets [--values]                 # list (masked by default)
localvault unset KEY [KEY2 ...]               # delete

# Scope override (any command)
localvault <cmd> --project X --config Y

# Run with secrets as env vars
localvault run -- <command> [args]

# Import/Export
localvault import <file.env>
localvault import --stdin
localvault export [--format dotenv|json|docker]

# Configs
localvault configs copy <src> <dst>           # duplicate secrets across envs
```

## Environment Variables

| Var | Purpose |
|-----|---------|
| `LOCALVAULT_DIR` | Override vault directory (default `~/.localvault`) |
| `LOCALVAULT_PASSPHRASE` | Passphrase for file-based keyring backend (headless/CI) |

## Naming Rules

- Projects/configs: `[a-z0-9_-]`, 1-64 chars
- Secret keys: `[A-Z0-9_]`, 1-256 chars (env var convention)
- Secret values: any UTF-8, max 64KB

## Common Workflows

**New project from scratch:**
```bash
localvault setup
localvault projects create api
localvault configure --project api
localvault configs create dev
localvault configure --config dev
localvault set DATABASE_URL=postgres://localhost/api
localvault run -- go run .
```

**Migrate from .env file:**
```bash
localvault import .env
localvault secrets --values   # verify
```

**Promote dev → prd:**
```bash
localvault configs copy dev prd
localvault configure --config prd
```

**CI/headless usage:**
```bash
export LOCALVAULT_PASSPHRASE="ci-secret"
localvault setup
localvault import secrets.env
localvault run -- make deploy
```
