# Agent Guidelines

This file defines how agents should operate in this repository. Domain and specification details live in the canonical project docs (e.g. `docs/`), not here.

---

## Core Principles

- **Read before editing** — Identify and read the relevant source-of-truth documentation before changing related files.
- **Minimize scope** — Prefer the smallest safe change; avoid touching unrelated files.
- **Preserve consistency** — Keep coupled areas in sync: `docs/`, `examples/`, and `sdk/` depend on each other.
- **Verify after changes** — Run the smallest relevant validation for touched areas.

---

## Working Process

1. **Identify affected areas** — Determine whether the change impacts docs, examples, tests, schemas, or SDK.
2. **Inspect context** — Review nearby code and documentation before editing.
3. **Make the change** — Apply edits with minimal scope.
4. **Verify touched surfaces** — Run relevant tests, lints, or validation for the changed areas.
5. **Check dependents** — If examples or generated artifacts depend on the change, verify they still conform.

---

## Subagent Usage

Use subagents as follows:

- **Design** — When the task involves design, solution shaping, or feature-level planning, **use the `designer` subagent** to produce a structured design before implementation.
- **Exploration** — Use `explore` for broad codebase search and consistency checks across docs/examples/sdk.
- **Implementation** — Use `generalPurpose` for multi-file changes and complex refactoring.
- **Verification** — Use `shell` for running tests, lint checks, and debugging support.

After any change that affects `examples/` or `sdk/`, run SDK tests via a subagent to confirm nothing is broken.

---

## Verification

- Run the smallest relevant checks for touched files or areas.
- When changing docs, verify examples still conform and SDK behavior matches.
- When changing examples, verify they still match the spec and pass SDK validation.
- When changing SDK, confirm it still supports patterns described in docs and exercised in examples.
- Typical commands: `cd sdk/python && uv run pytest` and `cd sdk/golang && go test ./bkn/...`.

---

## Communication

- Summarize what changed, what was verified, and any remaining risks or assumptions.
- Call out incomplete verification or uncertain dependencies.
