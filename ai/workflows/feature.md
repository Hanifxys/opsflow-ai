# AI Feature Workflow

Use this workflow for every non-trivial feature.

```text
1. Select TASK
2. Read source-of-truth documents
3. Analyse impact
4. Product check
5. Architecture check
6. API/schema design
7. Security review
8. Implementation
9. Unit tests
10. Integration tests
11. Static analysis
12. Documentation
13. Review
14. Mark task complete
```

## Gate 1 — Requirements

Feature must map to a requirement or explicitly approved task.

## Gate 2 — Architecture

Feature must not violate service ownership or communication rules.

## Gate 3 — Security

Feature must identify authentication, authorisation and secret impact.

## Gate 4 — Implementation

Only affected files should change.

## Gate 5 — Verification

Tests and checks must pass.

## Gate 6 — Documentation

Update affected `.md`/OpenAPI/schema/ADR artifacts.

## Gate 7 — Review

A reviewer agent checks the final diff.
