# PRD: ras-utils

## Problem

Clinical+ services duplicate common patterns: caching, validation, HTTP helpers, NATS publishing, timezone-aware scheduling. Each service implements its own version, leading to inconsistent behavior and repeated bugs.

## Solution

Shared Go utility library (`github.com/transactrx/ras-utils`) providing tested, documented implementations of common patterns. Services import what they need.

## User Stories

**As a service developer**, I want to:
- Add TTL caching without implementing cache eviction logic
- Validate NPIs, emails, phones with correct algorithms (Luhn checksum, etc.)
- Convert between Go types and pgtype without null-handling bugs
- Publish NATS events with consistent format and error handling
- Check location operating hours across timezones
- Write HTTP handlers with consistent JSON response format

**As a platform team**, I want to:
- Fix a bug once and have all services get the fix
- Enforce consistent patterns across the ecosystem
- Reduce onboarding time for new developers

## Success Metrics

| Metric | Target |
|--------|--------|
| Adoption | Used by all Clinical+ Go services |
| Bug rate | <1 bug report per quarter |
| Breaking changes | 0 unintentional per year |
| Test coverage | >80% on exported functions |

## Non-Goals

- CLI tools (this is a library)
- Service-specific business logic
- Database migrations or schemas
- Configuration management beyond env vars

## Constraints

- Go 1.21+
- No CGO dependencies
- Packages must be independently importable (no monolithic import)
- Semver with auto-tagging on merge

## Stakeholders

- **Consumers**: clinicalPlus, clinicalPlusPortal, clinicalDocumentsRepo, notificationEngine
- **Maintainers**: Platform team
