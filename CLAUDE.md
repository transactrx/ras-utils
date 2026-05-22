# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go utility library (`github.com/transactrx/ras-utils`) providing shared helper functions for the Clinical+ ecosystem. Designed to be imported by other repositories.

## Build Commands

```bash
go mod tidy          # Install/sync dependencies
go build ./...       # Build all packages
go test ./...        # Run all tests
go test -v ./...     # Run tests with verbose output
go test -run TestName ./rascache  # Run specific test in package
```

## Architecture

Utility library organized by functional domain - each package is independent with no cross-dependencies except `raslocation` depends on `rastime`:

- **rascache/** - Generic in-memory cache with TTL, thread-safe, supports UTC or local time expiration
- **rasconfig/** - Database pool setup (pgxpool) and environment variable helpers
- **rasconversion/** - Nullable Go types → pgtype conversions for PostgreSQL
- **rasevents/** - NATS event publishing with sync/async modes and worker pools
- **rashttp/** - HTTP request parsing, JSON responses, status helpers
- **raslocation/** - Location operating hours, timezone-aware scheduling (uses rastime)
- **raslogging/** - HTTP middleware with request logging and panic recovery
- **rasstack/** - HTTP middleware composition
- **rastime/** - TimeOfDay, TimeRange, DayOfWeek types for schedule management
- **rasworker/** - Generic worker pool with graceful shutdown

## Versioning

Auto-tagged on PR merge to main based on branch prefix:
- `major/*` → v1.2.3 → v2.0.0
- `minor/*` → v1.2.3 → v1.3.0  
- `build/*` or `feature/*` → v1.2.3 → v1.2.4

## Standards

**Think Before Coding** - State assumptions explicitly. Surface tradeoffs. Push back when simpler approaches exist.

**Simplicity First** - Minimum code that solves the problem. No speculative features, no abstractions for single-use code, no "flexibility" that wasn't requested.

**Surgical Changes** - Touch only what you must.

**Documentation** - Use Godoc style comments on all exported functions, types, and packages.

## Code Search

```bash
# Find function definitions
ast-grep --lang go -p 'func $NAME($$$) $$$'

# Find struct definitions
ast-grep --lang go -p 'type $NAME struct { $$$ }'
```

## Tooling

- TEXT/strings → `rg`
- CODE STRUCTURE → `ast-grep`
- Multiple results selection → pipe to `fzf`
- JSON → `jq`
- YAML/XML → `yq`
