# LocalVault — Implementation Plan

A Doppler-like CLI secrets manager backed by the OS keyring (Secret Service / Keychain / Credential Manager).

## 1. Core Concepts

| Concept | Description | Example |
|---------|-------------|---------|
| **Project** | Named collection of configs | `myapp`, `api`, `infra` |
| **Config** | Named environment within a project | `dev`, `stg`, `prd` |
| **Secret** | Key-value pair stored in a config | `DATABASE_URL=postgres://...` |
| **Active scope** | Currently selected project+config | `myapp.dev` |

## 2. Storage Architecture

```
┌─────────────────────────────────────────────────────┐
│                   OS Keyring                         │
│  (Secret Service / Keychain / Credential Manager)   │
│                                                     │
│  One entry per secret:                              │
│    service: "localvault"                            │
│    key:     "<project>.<config>.<SECRET_NAME>"      │
│    value:   "secret_value"                          │
│                                                     │
│  Example entries:                                   │
│    localvault / myapp.dev.DATABASE_URL = pg://...   │
│    localvault / myapp.dev.API_KEY      = sk-abc    │
│    localvault / myapp.prd.DATABASE_URL = pg://...   │
│                                                     │
│  Fallback (headless/CI):                            │
│    99designs/keyring encrypted file backend         │
│    ~/.localvault/keyring.enc (AES, passphrase)      │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  ~/.localvault/vault.json  (metadata only)          │
│                                                     │
│  {                                                  │
│    "version": 1,                                    │
│    "active_project": "myapp",                       │
│    "active_config": "dev",                          │
│    "projects": {                                    │
│      "myapp": {                                     │
│        "created_at": "...",                         │
│        "configs": {                                 │
│          "dev": {                                   │
│            "created_at": "...",                     │
│            "secrets": ["DATABASE_URL", "API_KEY"]   │
│          },                                         │
│          "prd": {                                   │
│            "created_at": "...",                     │
│            "secrets": ["DATABASE_URL"]              │
│          }                                          │
│        }                                            │
│      }                                              │
│    }                                                │
│  }                                                  │
└─────────────────────────────────────────────────────┘
```

**Why this design?**
- One keyring entry per secret → `set`/`unset` are atomic, no read-modify-write
- Secrets in OS keyring → encrypted at rest by the OS, never on disk in plaintext
- Metadata in JSON → fast listing, project/config structure without keyring round-trips
- The `secrets` array in metadata is an index (key names only) for fast listing
- 99designs/keyring provides `Keys()` for cross-validation and encrypted-file fallback for headless/CI

## 3. CLI Commands (Doppler-like UX)

```
localvault                              # Show help
localvault setup                        # Initialize ~/.localvault/
localvault status                       # Show active project/config + vault health

# Projects
localvault projects                     # List all projects
localvault projects create <name>       # Create project
localvault projects delete <name>       # Delete project + all configs + secrets
localvault projects info <name>         # Show project details

# Configs (environments)
localvault configs                      # List configs for active project
localvault configs create <name>        # Create config in active project
localvault configs delete <name>        # Delete config + its secrets
localvault configs copy <src> <dst>     # Copy all secrets between configs

# Scope
localvault configure                    # Interactive: set active project+config
localvault configure --project myapp    # Set active project
localvault configure --config prd       # Set active config
localvault configure --project X --config Y  # Set both

# Secrets
localvault set KEY=VALUE [KEY2=V2 ...]  # Set one or more secrets
localvault set KEY                      # Set secret (prompts for value, hidden input)
localvault get KEY                      # Get single secret value
localvault get KEY --plain              # Raw value (no formatting, for piping)
localvault secrets                      # List all secrets (keys + masked values)
localvault secrets --values             # List with actual values
localvault unset KEY [KEY2 ...]         # Delete one or more secrets

# Run (inject secrets as env vars)
localvault run -- <command> [args]      # Run command with secrets as env vars
localvault run --project X --config Y -- cmd  # Override scope

# Import/Export
localvault import <file.env>            # Import from .env file
localvault import --stdin               # Import from stdin
localvault export                       # Export as KEY=VALUE lines
localvault export --format json         # Export as JSON
localvault export --format docker       # Export as docker --env-file format
```

### Flags (global)
```
--project, -p    Override active project for this command
--config, -c     Override active config for this command
--output, -o     Output format: table (default), json, plain
--no-color       Disable colored output
--verbose, -v    Verbose logging
```

## 4. Package Structure

```
localvault/
├── main.go                     # Entry point
├── go.mod                      # module: github.com/xilistudios/localvault
├── go.sum
├── LICENSE
├── Makefile
├── README.md
├── PLAN.md
│
├── cmd/                        # Cobra commands (thin layer, calls internal/)
│   ├── root.go                 # Root command, global flags, version
│   ├── setup.go                # localvault setup
│   ├── status.go               # localvault status
│   ├── projects.go             # localvault projects [create|delete|info]
│   ├── configs.go              # localvault configs [create|delete|copy]
│   ├── configure.go            # localvault configure
│   ├── secrets.go              # localvault set/get/secrets/unset
│   ├── run.go                  # localvault run -- <cmd>
│   └── importexport.go         # localvault import/export
│
├── internal/
│   ├── vault/
│   │   ├── vault.go            # Vault struct: orchestrates metadata + keyring
│   │   ├── vault_test.go
│   │   ├── metadata.go         # Metadata file read/write/lock
│   │   ├── metadata_test.go
│   │   ├── keyring.go          # Keyring adapter (wraps 99designs/keyring)
│   │   ├── keyring_test.go
│   │   ├── scope.go            # Project/config scope resolution
│   │   └── scope_test.go
│   │
│   ├── model/
│   │   ├── model.go            # VaultFile, Project, Config structs
│   │   └── model_test.go
│   │
│   ├── envfile/
│   │   ├── parser.go           # .env file parser
│   │   ├── parser_test.go
│   │   ├── exporter.go         # Export to .env / JSON / docker format
│   │   └── exporter_test.go
│   │
│   └── output/
│       ├── table.go            # Table formatter (aligned columns)
│       ├── json.go             # JSON formatter
│       └── plain.go            # Plain formatter (for piping)
│
└── .github/
    └── workflows/
        └── ci.yml              # Build + test on Linux/macOS/Windows
```

## 5. Key Design Decisions

### 5.1 Keyring Key Format — One entry per secret
```
keyring item:
  Key:         "localvault.<project>.<config>.<SECRET_NAME>"
  Data:        []byte("secret_value")
  Label:       "localvault: <project>.<config>.<SECRET_NAME>"
  Description: "Managed by localvault"
```

Using `99designs/keyring` — each secret is an independent keyring item:
- `Set KEY=V` → one keyring write, no read-modify-write
- `unset KEY` → one keyring delete
- `get KEY` → one keyring read
- `secrets` / `run` → iterate metadata index, one read per secret
- `Keys()` available for cross-validation with metadata index

**Backend resolution (99designs/keyring):**
1. macOS → Keychain
2. Linux desktop → Secret Service (D-Bus) or KWallet
3. Windows → Credential Manager
4. Headless/CI → Encrypted file backend (`~/.localvault/keyring.enc`, passphrase via env var `LOCALVAULT_PASSPHRASE` or prompt)

This means `localvault` works everywhere: laptops, Docker, CI runners.

### 5.2 Concurrency
- File lock on vault.json using `flock` (syscall on Linux/macOS) to prevent corruption
- Keyring operations are per-item atomic (no read-modify-write needed)
- Metadata index update is the only shared state → flock covers it

### 5.3 Security
- Secrets NEVER written to disk in plaintext (only in keyring)
- Metadata file (vault.json) contains only key names, no values
- `localvault secrets` masks values by default (`DATABASE_URL = pg****://...`)
- `localvault run` passes secrets via env vars (child process only, not persisted)
- No telemetry, no network calls, fully offline

### 5.4 Validation Rules
- Project names: `[a-z0-9_-]`, 1-64 chars
- Config names: `[a-z0-9_-]`, 1-64 chars
- Secret keys: `[A-Z0-9_]`, 1-256 chars (env var convention)
- Secret values: any UTF-8 string, max 64KB

## 6. Dependencies

```
github.com/spf13/cobra           # CLI framework
github.com/99designs/keyring     # OS keyring (Secret Service, Keychain, WinCred, encrypted file fallback)
github.com/fatih/color           # Colored output
golang.org/x/term                # Hidden input for secret values
```

Minimal dependency footprint. No database, no server, no network.
`99designs/keyring` pulls in D-Bus bindings on Linux but is well-maintained (used by aws-vault, 1Password CLI).

## 7. Implementation Phases

### Phase 1 — Foundation (cmd/root, internal/model, internal/vault/metadata)
- [ ] `go mod init`, Makefile, main.go
- [ ] Model structs (VaultFile, Project, Config)
- [ ] Metadata read/write with file locking
- [ ] Root command with global flags + version
- [ ] `localvault setup` — create ~/.localvault/vault.json
- [ ] `localvault status` — show active scope + health check

### Phase 2 — Keyring + Core CRUD (internal/vault/keyring, vault.go)
- [ ] Keyring adapter: Get/Set/Remove per-secret items via 99designs/keyring
- [ ] Backend detection: OS keyring vs encrypted file fallback (headless/CI)
- [ ] Vault orchestrator: ties metadata + keyring together
- [ ] `localvault projects create/delete/info`
- [ ] `localvault configs create/delete/copy`
- [ ] `localvault configure` (interactive + flags)

### Phase 3 — Secrets Operations (cmd/secrets)
- [ ] `localvault set KEY=VALUE` (single + multi + hidden prompt)
- [ ] `localvault get KEY` (+ --plain)
- [ ] `localvault secrets` (masked + --values)
- [ ] `localvault unset KEY`

### Phase 4 — Run + Import/Export (cmd/run, internal/envfile)
- [ ] `localvault run -- <command>` (inject env vars)
- [ ] .env parser (handles quotes, comments, multiline)
- [ ] `localvault import <file>` / `--stdin`
- [ ] `localvault export` (plain / json / docker formats)

### Phase 5 — Polish
- [ ] Output formatters (table/json/plain)
- [ ] Colored output + --no-color
- [ ] Shell completions (bash/zsh/fish)
- [ ] CI workflow (Linux/macOS/Windows)
- [ ] README with usage examples
- [ ] Error messages: helpful, suggest fixes

## 8. Example Usage Flow

```bash
# First time
$ localvault setup
✓ Vault initialized at ~/.localvault/

# Create project structure
$ localvault projects create myapp
✓ Created project "myapp"

$ localvault configs create dev
✓ Created config "dev" in project "myapp"

$ localvault configs create prd
✓ Created config "prd" in project "myapp"

# Set secrets
$ localvault configure --project myapp --config dev
✓ Active scope: myapp.dev

$ localvault set DATABASE_URL=postgres://localhost:5432/myapp API_KEY=sk-abc123
✓ Set 2 secrets in myapp.dev

$ localvault set STRIPE_SECRET
? Enter value for STRIPE_SECRET: ********
✓ Set STRIPE_SECRET in myapp.dev

# List secrets
$ localvault secrets
  KEY              VALUE
  DATABASE_URL     pg****://localhost:5432/myapp
  API_KEY          sk****23
  STRIPE_SECRET    sk****...

# Run with secrets injected
$ localvault run -- node server.js
# server.js sees process.env.DATABASE_URL, etc.

# Import from .env
$ localvault import .env
✓ Imported 12 secrets into myapp.dev

# Export
$ localvault export --format json
{"DATABASE_URL":"postgres://...","API_KEY":"sk-abc123"}

# Copy config
$ localvault configs copy dev prd
✓ Copied 3 secrets from myapp.dev → myapp.prd
```

## 9. Testing Strategy

- **Unit tests**: model validation, metadata CRUD, env parser, exporters
- **Integration tests**: full vault operations with mock keyring (in-memory map)
- **Keyring mock**: `internal/vault/keyring.go` uses an interface, tests inject a map-based mock
- **CLI tests**: execute commands via cobra's `Execute()` with captured stdout
- Target: >80% coverage on internal/ packages

```go
// SecretStore interface for testability
type SecretStore interface {
    Get(key string) ([]byte, error)
    Set(key string, data []byte) error
    Remove(key string) error
    Keys() ([]string, error)
}

// Production: OSKeyring (99designs/keyring, auto-detects backend)
// Tests:      MockKeyring (in-memory map[string][]byte)
```

### Keyring item key format in code:
```go
func secretKey(project, config, name string) string {
    return fmt.Sprintf("localvault.%s.%s.%s", project, config, name)
}
// e.g. "localvault.myapp.dev.DATABASE_URL"
```

## 10. Makefile Targets

```makefile
build       # go build -o bin/localvault
test        # go test ./... -v -count=1
lint        # golangci-lint run
fmt         # gofmt + goimports
install     # go install
clean       # rm -rf bin/
ci          # fmt + lint + test + build
```
