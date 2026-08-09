# OpsFlow — AI Agent Contract

## Purpose

This file tells coding agents how to operate inside the repository.

## Mandatory first step

Before changing code, read:

1. `rules.md`
2. `prd.md`
3. `architecture.md`
4. `design.md`
5. `schema.md`
6. `IMPLEMENTATION-PLAN.md`
7. `TASKS.md`

Only read the relevant sections when the repository becomes large.

## Never

- invent a requirement
- silently change architecture
- add a technology because it looks impressive
- bypass the API Gateway for public traffic
- call LLM providers directly from business services
- expose secrets
- skip tests
- modify database schema without a migration
- introduce breaking API changes without an ADR
- give AI autonomous production mutation access
- rewrite unrelated code

## Before implementation

The agent must identify:

```text
Task
 ├── Requirement
 ├── Affected service
 ├── Affected API
 ├── Affected schema
 ├── Events
 ├── Security impact
 ├── Tests
 └── Documentation
```

## After implementation

The agent must verify:

```text
Implementation
 ↓
Unit tests
 ↓
Integration tests
 ↓
Lint/static analysis
 ↓
Security checks
 ↓
Documentation
 ↓
Task status
```

## Scope discipline

If a task reveals an architectural problem, stop and report it rather than silently redesigning the platform.

## Commit convention

Use:

```text
feat:
fix:
refactor:
test:
docs:
chore:
ci:
security:
```

Example:

```text
feat(incident): add incident lifecycle API
```
