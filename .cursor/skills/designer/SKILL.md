---
name: designer
description: Produce structured designs for features, architecture changes, and solution shaping before implementation. Use when the task involves design, planning, feature scoping, API design, spec changes, or any work that benefits from a design-first approach before coding.
---

# Designer

Produce structured, actionable design documents before implementation begins. This skill ensures that non-trivial changes are thought through — covering scope, trade-offs, affected areas, and verification strategy — before any code is written.

## When to Use

- New features or capabilities (SDK, CLI, spec additions)
- Architectural changes (parser pipeline, loader refactoring, new transformers)
- Spec changes that ripple into examples and SDK
- API or data model design (new object properties, relation types, action schemas)
- Cross-cutting concerns (new validation rules, new file formats, risk model changes)

Do **not** use for trivial fixes, typo corrections, or single-line changes.

## Workflow

1. **Understand the request** — Clarify what the user wants to achieve, not just what they asked for.
2. **Survey the landscape** — Read relevant docs, code, and examples to understand current state.
3. **Identify affected areas** — Map which parts of the codebase are impacted: `docs/`, `examples/`, `sdk/`, `cli/`.
4. **Design the solution** — Produce a structured design (see Output Format below).
5. **Review trade-offs** — Explicitly state alternatives considered and why the chosen approach wins.
6. **Define verification** — Specify how to confirm the implementation is correct.

## Context Loading

Before designing, load the relevant sources of truth:

| Area | Source |
|------|--------|
| BKN spec | `docs/SPECIFICATION.md` or `docs/SPECIFICATION.en.md` |
| Architecture | `docs/ARCHITECTURE.md` |
| Templates | `docs/templates/` |
| Examples | `examples/` (pick relevant ones) |
| Python SDK | `sdk/python/src/bkn/` |
| Golang SDK | `sdk/golang/bkn/` |
| CLI | `cli/` |

Only load what is relevant to the design task. Do not read the entire codebase.

## Output Format

Produce a design document with these sections:

```markdown
## Goal

One-sentence summary of what this design achieves.

## Background

Brief context: current state, why the change is needed, any prior art.

## Scope

- **In scope**: What this design covers
- **Out of scope**: What it explicitly does not cover

## Design

Detailed description of the approach:
- Data model changes (if any)
- API / interface changes (if any)
- File format changes (if any)
- Behavioral changes (if any)

Use tables, diagrams (mermaid), or code snippets where they clarify.

## Affected Areas

| Area | Impact | Files / Directories |
|------|--------|---------------------|
| Spec | ... | `docs/SPECIFICATION.md` |
| SDK (Python) | ... | `sdk/python/src/bkn/...` |
| SDK (Golang) | ... | `sdk/golang/bkn/...` |
| Examples | ... | `examples/...` |
| CLI | ... | `cli/...` |
| Tests | ... | ... |

## Alternatives Considered

| Alternative | Pros | Cons | Why not chosen |
|-------------|------|------|----------------|
| ... | ... | ... | ... |

## Migration / Compatibility

- Breaking changes (if any) and migration path
- Backward compatibility considerations

## Implementation Plan

Ordered list of implementation steps, each small enough for a single PR or commit.

## Verification

- What tests to add or update
- What manual checks to perform
- Commands to run: `cd sdk/python && uv run pytest`, `cd sdk/golang && go test ./bkn/...`
```

Sections may be omitted if truly not applicable, but prefer explicit "N/A" over silent omission.

## Guard Rules

| Condition | Action |
|-----------|--------|
| Request is vague or ambiguous | Ask clarifying questions before designing |
| Design would require spec changes | Flag explicitly; spec changes need careful review |
| Multiple valid approaches exist | Present top 2-3 with trade-offs; recommend one |
| Design affects public API surface | Call out backward compatibility impact |
| Scope is very large | Suggest phasing; design phase 1 in detail, outline later phases |

## Principles

- **Spec is the source of truth** — Designs must conform to or explicitly propose changes to `docs/SPECIFICATION.md`.
- **Consistency across layers** — If the design changes one layer (spec, SDK, examples), identify ripple effects to other layers.
- **Minimal blast radius** — Prefer approaches that change fewer files and touch fewer subsystems.
- **Testability** — Every design should be verifiable; if it cannot be tested, reconsider the approach.
- **Incremental delivery** — Prefer designs that can be implemented and verified in small steps.
