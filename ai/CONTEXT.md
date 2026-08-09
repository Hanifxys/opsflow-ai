# AI Context Loading Strategy

Do not load the entire repository into every prompt.

## Always load

- AGENTS.md
- rules.md

## Product work

Load:

- prd.md
- TASKS.md

## Architecture work

Load:

- architecture.md
- design.md
- relevant ADRs

## Database work

Load:

- schema.md
- design.md
- relevant service design

## AI work

Load:

- architecture.md
- design.md
- rules.md
- AI-specific prompts

## Coding task

Load:

- selected task
- relevant service files
- relevant design sections
- relevant schema/API sections

The goal is to provide enough context for correctness without flooding the model with unrelated information.
