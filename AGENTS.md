# AgentDock Agent Instructions

This file is the required entry point for coding agents working in this
repository. Read it at the start of every goal before editing files.

## Project Boundary

- AgentDock is a greenfield project. `multica-upstream/` is reference material
  only and must not be committed, copied from blindly, or treated as the
  implementation base.
- Product feature decisions belong to the user. Agents may recommend options and
  tradeoffs, but must not silently decide product scope, MVP behavior, or UX
  direction.
- Engineering execution is delegated to the agent: planning, implementation,
  verification, atomic commits, and repository memory updates.

## Required Startup

At the start of each goal:

1. Read this file.
2. Read [ENGINEERING.md](ENGINEERING.md).
3. Read the relevant memory:
   - [CONTEXT.md](CONTEXT.md) for terms and durable project context.
   - [SCOPE.md](SCOPE.md) for product scope and non-goals.
   - `docs/adr/` for accepted architectural decisions.
4. Check `git status --short` before editing.
5. Identify applicable skills before acting.

## Skill Use

- Use relevant skills proactively.
- Use `planning-with-files` for complex, multi-step, cross-module, research-heavy
  work, or any task likely to require 5+ tool calls. Keep plans under
  `.planning/<goal>/`.
- Use brainstorming/design flow for new behavior, product decisions, or broad
  feature design. Do not bypass approval gates when a skill requires them.
- Use debugging and TDD-oriented workflows when fixing bugs or implementing
  risky behavior.
- If a skill clearly applies but is skipped, state why.

## HITL Threshold

Default to autonomous execution. Stop and ask the user only for:

- Architecture boundary changes.
- Irreversible or destructive operations.
- External services, credentials, paid resources, or networked side effects.
- Security, privacy, or data-retention risk.
- Changes that conflict with existing ADRs or durable memory.
- Product feature scope or UX decisions.

Do not ask for routine implementation choices, small refactors, formatting,
focused tests, or local verification.

## Work Loop

- Work in small, verifiable increments.
- Prefer existing patterns over new abstractions.
- Keep changes scoped to the current goal.
- Before editing, state what will change.
- After editing, run the relevant verification from [ENGINEERING.md](ENGINEERING.md).
- Commit in the smallest coherent atomic units. A single goal may produce many
  commits.
- Update repository memory when durable facts, decisions, workflows, or pitfalls
  change.

## Git Rules

- Never revert user changes unless explicitly asked.
- Never commit `multica-upstream/`.
- Never use destructive git commands unless the user explicitly requests them.
- Commit messages should be concise conventional-style subjects when possible:
  `docs: ...`, `feat: ...`, `fix: ...`, `test: ...`, `refactor: ...`.
- Commit bodies should record goal context, verification, and important
  decisions when the subject is not enough.

## Final Response

When finishing a goal, report:

- What changed.
- Commit hash(es), if commits were made.
- Verification run and anything not run.
- Memory updates made, if any.
