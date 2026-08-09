# AI Quality Gates

A change can move forward only when the relevant gates pass.

## Gate A — Requirement

- [ ] mapped to task
- [ ] acceptance criteria defined

## Gate B — Design

- [ ] architecture compliant
- [ ] data ownership defined
- [ ] API/event contract defined

## Gate C — Security

- [ ] auth considered
- [ ] permissions considered
- [ ] secrets considered
- [ ] input validation considered

## Gate D — Implementation

- [ ] code compiles
- [ ] error handling exists
- [ ] no unrelated changes

## Gate E — Tests

- [ ] unit
- [ ] integration where relevant
- [ ] regression
- [ ] E2E for critical flow

## Gate F — Operations

- [ ] logs
- [ ] metrics
- [ ] health checks
- [ ] failure behaviour

## Gate G — Documentation

- [ ] API
- [ ] schema
- [ ] architecture/ADR
- [ ] runbook if operationally relevant

## Gate H — Review

- [ ] reviewer agent passed
- [ ] no blocker/high findings
