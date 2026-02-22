```
                                            _
  ___ _ ____   __ __ _ _   _  __ _ _ __ __| |
 / _ \ '_ \ \ / // _` | | | |/ _` | '__/ _` |
|  __/ | | \ V /| (_| | |_| | (_| | | | (_| |
 \___|_| |_|\_/  \__, |\__,_|\__,_|_|  \__,_|
                  |___/
```

# envguard

**Your .env files are broken. You just don't know it yet.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![CI](https://github.com/ruidosujeira/envguard/actions/workflows/ci.yml/badge.svg)](https://github.com/romano/envguard/actions)

A linter, validator, and synchronizer for `.env` files. Catches duplicate keys, typos, exposed secrets, sync drift, and formatting issues — before they break production.

## Install

```bash
# Go
go install github.com/ruidosujeira/envguard/cmd/envguard@latest

# Homebrew (coming soon)
# brew install envguard

# Or download a binary from Releases
```

## Quick Start

```bash
# Lint your .env files
envguard

# Detect exposed secrets
envguard secrets

# Check if .env and .env.example are in sync
envguard sync
```

## What It Catches

### Lint Rules

```bash
$ envguard lint --no-color

.env
  3:1    error    Duplicate key: DATABASE_URL (first defined at line 1)
  5:1    warning  Possible typo: DATBASE_URL (did you mean DATABASE_URL?)
  7:1    info     Empty value: REDIS_URL
  9:4    error    Invalid character "-" in key "FOO-BAR" at position 4
  11:1   warning  Key "API KEY" has surrounding whitespace
  0:0    warning  .env is not in .gitignore

✖ 2 errors, 3 warnings, 1 info
```

- **Duplicate keys** — the second silently overwrites the first
- **Typo detection** — Levenshtein distance against known env var names
- **Invalid characters** — keys must be `[A-Za-z0-9_]`, starting with a letter or `_`
- **Formatting** — trailing whitespace, spaces in keys, excessive blank lines
- **Empty values** — variables declared but never assigned
- **Gitignore check** — warns if `.env` is not in `.gitignore`

### Secret Detection

```bash
$ envguard secrets

.env
  2:1    error    Possible secret detected: AWS_KEY (matches AWS Access Key ID pattern)
  5:1    error    Possible secret detected: SECRET_KEY (high entropy value: 4.8)
  8:1    error    Possible secret detected: DATABASE_URL (matches Generic password in URL pattern)
```

- Pattern matching for AWS keys, JWTs, Stripe keys, GitHub tokens, private keys, passwords in URLs
- Shannon entropy analysis for values that look like random secrets
- Only flags values that actually look dangerous

### Sync

```bash
$ envguard sync

.env.example
  Missing keys (present in .env but not in .env.example):
    - NEW_API_KEY
    - FEATURE_FLAG_X

$ envguard sync --fix
Added 2 missing key(s) to .env.example
```

## Commands

| Command | Description |
|---------|-------------|
| `envguard` | Lint `.env` files in current directory (alias for `lint`) |
| `envguard lint [files...]` | Lint specific files |
| `envguard lint --fix` | Auto-fix formatting issues |
| `envguard secrets [files...]` | Detect exposed secrets |
| `envguard sync` | Check `.env` ↔ `.env.example` sync |
| `envguard sync --fix` | Auto-add missing keys to `.env.example` |
| `envguard check [files...]` | Run all checks (lint + secrets + sync) |

## Flags

| Flag | Description |
|------|-------------|
| `--format json` | Output as JSON (for CI integration) |
| `--no-color` | Disable colored output |
| `--strict` | Exit non-zero on any finding (not just errors) |
| `--fix` | Auto-fix issues where possible |

## CI Integration

```yaml
# GitHub Actions
- name: Lint .env files
  run: envguard check --format json --strict
```

Exit codes:
- `0` — no errors (warnings/info don't cause failure unless `--strict`)
- `1` — errors found (or any findings with `--strict`)

## envguard vs dotenv-linter

| Feature | envguard | dotenv-linter |
|---------|----------|---------------|
| Duplicate keys | ✅ | ✅ |
| Formatting checks | ✅ | ✅ |
| Typo detection (Levenshtein) | ✅ | ❌ |
| Secret detection (patterns) | ✅ | ❌ |
| Secret detection (entropy) | ✅ | ❌ |
| .gitignore check | ✅ | ❌ |
| Sync .env ↔ .env.example | ✅ | ❌ |
| Auto-fix sync | ✅ | ❌ |
| JSON output | ✅ | ❌ |
| Zero core dependencies | ✅ | N/A (Rust) |

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b feature/amazing`)
3. Run tests (`go test ./...`)
4. Commit your changes
5. Open a PR

## License

[MIT](LICENSE)
