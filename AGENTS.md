# AGENTS.md

Behavioral constraints for LLM agents working in this repository.

## Scope

This is a **shared utility library** imported by other services. Changes here affect the entire Clinical+ ecosystem.

## Code Style

- Match existing patterns exactly — no "improvements" to adjacent code
- Godoc comments on all exported functions, types, and packages
- No comments explaining what code does; only why (hidden constraints, workarounds)
- Minimum code that solves the problem — no speculative features

## Change Rules

- Touch only what the task requires
- Remove imports/variables YOUR changes made unused
- Do NOT remove pre-existing dead code unless asked
- Do NOT refactor working code unless asked

## Testing

- Run `go test ./...` before considering work complete
- New exported functions require tests
- Bug fixes require a regression test

## Security

- No secrets in code or comments
- Validate at trust boundaries only (this lib is internal)
- No `//nolint` without justification

## File Access

- All packages are independent except `raslocation` depends on `rastime`
- No circular dependencies
- No new packages without explicit request

## Prohibited

- Breaking changes to exported APIs without major version bump
- Adding dependencies for functionality achievable in <50 lines
- Generic "helper" functions without concrete use case
