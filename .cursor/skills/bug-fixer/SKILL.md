<!-- markdownlint-disable MD060 -->
---
name: bug-fixer
description: Systematically diagnose and fix bugs from error messages, stack traces, test failures, or user-reported issues. Use when the user reports a bug, pastes an error, asks to fix a failing test, or mentions debugging.
---

# Bug Fixer

Systematically diagnose and fix bugs. Follow the workflow below; do not skip steps.

## Workflow

1. **Classify** — Identify bug source type and extract key signals
2. **Reproduce** — Locate the failing test or minimal reproduction steps
3. **Root cause analysis** — Find and understand the relevant code
4. **Fix** — Apply minimal, targeted change
5. **Verify** — Run tests and linters to confirm fix and no regressions

## Classification

| Source Type | Signals | Action |
|-------------|---------|--------|
| Error message / stack trace | File path, line number, exception type | Jump to location, read surrounding code |
| Failing test | Test name, test file, assertion message | Run the specific test, read test and target code |
| Linter error | File, rule ID, message | Read linter output, fix at reported location |
| User description | Feature, behavior, "X doesn't work" | Search for related code, ask for steps if vague |

Extract: file path, line number, function/class name, error type. Use these to narrow search.

## Reproduce

- **Test failure**: Run the failing test in isolation (e.g. `pytest path/to/test.py::test_name`, `go test -run TestName ./...`)
- **Runtime error**: If user provided steps, try to reproduce; if not, rely on stack trace location
- **Linter**: Re-run linter after fix to confirm

Auto-detect test runner from project:

| Indicator | Command |
|-----------|---------|
| `pytest`, `pyproject.toml` with pytest | `pytest` or `uv run pytest` |
| `go.mod`, `*_test.go` | `go test ./...` or `go test ./path/...` |
| `package.json`, `jest`/`vitest`/`mocha` | `npm test` or `npx vitest` |
| `Cargo.toml`, `#[test]` | `cargo test` |

## Root Cause Analysis

1. **Locate** — Use file:line from error, or Grep for function/class/module name
2. **Read** — Read the failing code and its callers; understand data flow
3. **Identify** — Find the root cause (logic error, wrong assumption, edge case), not just the symptom

Avoid fixing symptoms (e.g. adding null checks) without addressing the underlying cause.

## Fix

- **Minimal change** — Change only what is necessary to fix the bug
- **No unrelated edits** — Do not refactor, rename, or "improve" unrelated code
- **Preserve behavior** — Do not change intended behavior elsewhere
- **Follow project style** — Match existing patterns, indentation, naming

## Verify

1. **Run tests** — Execute the previously failing test, then the full test suite for the affected area
2. **Check lints** — Run linter; fix any new issues introduced by the fix
3. **Confirm** — Ensure the original error is gone and no new failures appear

For this project (BKN specification):

- Python SDK: `cd sdk/python && uv run pytest`
- Golang SDK: `cd sdk/golang && go test ./bkn/...`

## Guard Rules

| Condition | Action |
|-----------|--------|
| Bug description is vague, no error message | Ask user for steps to reproduce or paste the error |
| Multiple plausible causes | Analyze each; prefer the simplest fix that matches the error |
| Fix would require large refactor | Propose minimal fix first; suggest refactor as follow-up if needed |
| Tests are missing for the buggy code | Fix the bug; optionally suggest adding a test |
