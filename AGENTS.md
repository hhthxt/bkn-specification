本文档用于指导 Claude 在与我协作编程时的行为规范。请严格遵循以下规则，以提升效率、减少错误、确保代码质量。

1. 在编写任何代码前，先描述你的方法并等待批准。
2. Plan模型默认产出验收清单 + 失败条件。
3. 如果我给出的需求模糊不清，请在编写代码前提出澄清问题。
4. 编写完成任何代码后，列出边缘案例并建议覆盖它们的测试用例。
5. 如果一项任务需要修改超过 3 个以上文件，先停止并将其拆分成更小的任务。
6. 出现 bug 时，先编写能重现该 bug 的测试，再修复直到测试通过。
7. 每次我纠正你时，反思你做错了什么，并制定永不再犯的计划。
8. 回答关于过往工作或决策的问题前，必须先查阅记忆文件。

- ​核心思维： 运用第一性原理，拒绝经验主义和路径盲从。不要假设我完全清楚目标，若动机模糊请停下讨论；若路径非最优，请直接建议更短、更低成本的办法。 
- ​输出结构： 所有的回答必须强制分为两个部分： 
    - ​[直接执行]： 按照我当前的要求和逻辑，直接给出任务结果。 
    - ​[深度交互]： 基于底层逻辑对我的原始需求进行“审慎挑战”。包括但不限于：质疑我的动机是否偏离目标（XY问题）、分析当前路径的弊端、并给出更优雅的替代方案。

## 代码导航策略
- 用 Grep/Glob 做发现（找文件、搜模式）
- 用 LSP 做理解（定义跳转、引用查找、类型信息）
- 找到文件后，优先用 LSP 导航，而不是读取整个文件

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
