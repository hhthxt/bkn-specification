---
name: bkn-creator
description: Generate BKN (Business Knowledge Network) files for modeling objects, relations, and actions. Optionally import to kweaver via API. Use when the user asks to create a BKN file, define a knowledge network, model objects/relations, generate .bkn files, import to kweaver, or run BKN/SDK scripts.
---

# BKN Creator

Generate `.bkn` files conforming to the BKN specification, and optionally import them to kweaver. BKN files are the single source of truth for domain knowledge; this Skill orchestrates loading, validation, and execution.

**File extensions** — Supported at runtime: `.bkn`, `.bknd`, `.md`. Content must satisfy BKN frontmatter/type/structure regardless of extension. Recommended: schema `.bkn`, data `.bknd`; use `.md` when coexisting with generic Markdown tooling.

## BKN-Driven Execution Protocol

**Request classification** — Route user intent into one of:

| 类型 | 识别特征 | 行为 |
|------|----------|------|
| `create/update model` | 创建/修改/定义 对象/关系/行动/网络 | 生成或改写 `.bkn/.bknd/.md`，走生成模式 |
| `validate/transform/import` | 校验/转换/导入 | 直接进入脚本链路 |
| `operate action` | 执行某操作、调用某工具、对某对象做某事 | 从 Action 定义解析工具与参数，走执行模式 |

**Load flow** — For `operate action` or when context needs domain knowledge:

1. Load `network.bkn` (preferred) or `index.bkn` (compatible) to get network overview and `includes`. Can also pass a directory — SDK auto-discovers the root.
2. Load relevant includes (objects/relations/actions) by task
3. Do not load entire network if only a subset is needed

**Action selection** — When multiple Actions match user intent:

- Use `Bound Object` + `Trigger Condition` to narrow candidates
- Prefer `risk_level: low` and `enabled: true`
- If ambiguous or high-risk, ask user to confirm before executing

**Guard rules** (before executing any Action):

| 条件 | 行为 |
|------|------|
| `enabled != true` | 拒绝执行，返回「需先启用该行动」 |
| `requires_approval == true` | 进入审批等待，不直接执行 |
| `risk_level == high` | 强制二次确认，建议先 `import --dry-run` |
| 字段缺失 | 按保守策略处理（视为高风险待确认） |

**Parameter resolution** — From `Parameter Binding`:

- `property`: resolve from target object instance
- `input`: collect from user if missing
- `const`: use as-is

**Output modes**:

- **生成模式**：仅输出 BKN Markdown，无多余说明
- **执行模式**：返回操作结果、风险/审批状态、以及审计提示（如 dry-run 建议）

## Workflow

**顺序**：先读规范 → 再选模板 → 再生成输出。

1. **理解需求**：识别业务域、对象、关系、行动
2. **读规范**：加载 [references/specification.md](references/specification.md)，按格式规则生成
3. **选模板**：从 [assets/](assets/) 选取对应类型（object、relation、action、network、data）
4. **生成 `.bkn` 文件**：按下方 Output Rules 输出
5. **（可选）校验/导入**：按 Script Chain 顺序运行脚本

## Scripts

Scripts live in [scripts/](scripts/). Run from repo root. Install first: `pip install -e sdk/python` or `pip install -e "sdk/python[api]"` for import.

**Script chain** — Recommended order for model changes and imports:

1. **validate** — After any BKN edit, run first to ensure load succeeds
2. **transform** or **import --dry-run** — Before real import, verify transform
3. **import** — Only after dry-run passes

| 触发条件 | 执行 |
|----------|------|
| 用户修改/新增了 `.bkn` / `.bknd` / `.md` | `validate.py <path>` |
| 用户要导入到 kweaver | `validate` → `import --dry-run` → `import` |
| 用户仅要导出 JSON | `validate` → `transform.py` |

**validate.py** — Check BKN loads (accepts file or directory):
```bash
python .cursor/skills/bkn-creator/scripts/validate.py <path-or-dir>
# e.g. python .cursor/skills/bkn-creator/scripts/validate.py examples/k8s-modular/index.bkn
# e.g. python .cursor/skills/bkn-creator/scripts/validate.py examples/k8s-network
```

**transform.py** — Export to kweaver JSON (no API):
```bash
python .cursor/skills/bkn-creator/scripts/transform.py <path> -o <output_dir> [--id-prefix PREFIX]
# e.g. python .cursor/skills/bkn-creator/scripts/transform.py examples/k8s-modular/index.bkn -o output
```

**import_to_kweaver.py** — Import via API:
```bash
# Internal mode (default, account headers)
python .cursor/skills/bkn-creator/scripts/import_to_kweaver.py <path> --account-id X --account-type Y [--base-url URL] [--id-prefix PREFIX]

# External mode (Bearer token from KWEAVER_TOKEN or --token)
python .cursor/skills/bkn-creator/scripts/import_to_kweaver.py <path> --external [--base-url URL] [--id-prefix PREFIX]

# Dry-run (transform only)
python .cursor/skills/bkn-creator/scripts/import_to_kweaver.py <path> --dry-run --account-id X --account-type Y
```

When the user asks to validate, convert, or import, run the corresponding script directly.

**Validation scenarios** — Use these to verify BKN-Skill integration:

| 场景 | Action 配置 | 预期行为 |
|------|-------------|----------|
| 低风险直接执行 | `risk_level: low`, `enabled: true`, `requires_approval: false` | 可直接执行，无需审批 |
| 高风险拦截 | `risk_level: high` 或 `requires_approval: true` | 强制二次确认，建议 dry-run，不直接执行 |
| 未启用不可执行 | `enabled: false` | 拒绝执行，返回「需先启用该行动」 |

## File Organization

Choose an organization style based on network size:

**Modular** (recommended for large networks, team collaboration):
Each object/relation/action in its own file, with a `network.bkn` (preferred) or `index.bkn` root.
See `examples/k8s-modular/` for this pattern.

```
my-network/
├── network.bkn                  # or index.bkn (compatible)
├── objects/
│   ├── order.bkn
│   └── customer.bkn
├── relations/
│   └── order_belongs_customer.bkn
└── actions/
    └── cancel_order.bkn
```

**By-type split** (suitable for medium networks):
Group all objects, relations, actions into separate fragment files.
See `examples/k8s-network/` for this pattern.

```
my-network/
├── network.bkn                  # or index.bkn (compatible)
├── objects.bkn
├── relations.bkn
└── actions.bkn
```

## Templates (assets/)

Read the appropriate template before generating:

- `assets/object.bkn.template` — object with Data Properties, Property Override, Logic Properties
- `assets/relation.bkn.template` — relation with Endpoints, Mapping Rules (direct and data_view)
- `assets/action.bkn.template` — action with Trigger, Pre-conditions, Tool Configuration, Schedule
- `assets/network.bkn.template` — network index with inline object/relation/action definitions
- `assets/data.bknd.template` — instance data file for objects with Data Source Type `bknd`

Fill in `{placeholders}`, remove unused optional sections, and remove template comments.

## Output Rules

**Generate mode** (create/update model):

1. Output **only** valid BKN Markdown (frontmatter + body). No extra explanation around the file content.
2. Do **not** wrap the entire output in a code fence.
3. Use existing object/relation IDs when referencing other definitions in the same network.
4. Follow table column names exactly as defined in the spec.
5. IDs: lowercase letters, digits, underscores. Display names and descriptions in Chinese unless specified otherwise.
6. Required fields: `type`, `id`, `name`, `network` in frontmatter.
   - Object: must have Data Source + at least one Primary Key and one Display Key.
   - Relation: must have Endpoints + Mapping Rules.
   - Action: must have Bound Object + Trigger Condition.

**Execute mode** (operate action):

- Return operation result, risk/approval status, and audit hints (e.g. suggest dry-run for high-risk).

## Kweaver Import

To import the generated BKN network to kweaver via API, use the Python SDK (`pip install -e "sdk/python[api]"`).

Two API modes: **internal** (default, no token, uses account headers) and **external** (Bearer token).

**Internal API** (inside cluster, no auth token):
```python
from bkn import load_network
from bkn.transformers import KweaverClient, KweaverTransformer

network = load_network("path/to/index.bkn")
client = KweaverClient(
    base_url="http://ontology-manager-svc:13014",  # or KWEAVER_BASE_URL
    account_id="your_account_id",
    account_type="your_account_type",
    business_domain="your_domain_id",
    internal=True,  # default
)
transformer = KweaverTransformer(id_prefix="my_prefix_")
result = client.import_network(network, transformer)
```

**External API** (Bearer token):
```python
client = KweaverClient(
    base_url="https://your-gateway/api",
    token="your_bearer_token",  # or KWEAVER_TOKEN
    account_id="your_account_id",
    account_type="your_account_type",
    business_domain="your_domain_id",
    internal=False,
)
result = client.import_network(network, transformer)
```

Dry-run (transform only, no API calls):

```python
result = client.import_network(network, transformer, dry_run=True)
```

For step-by-step control:

```python
payload = transformer.to_json(network)
kn_id = client.create_knowledge_network(payload["knowledge_network"])
client.create_object_types(kn_id, payload["object_types"])
client.create_relation_types(kn_id, payload["relation_types"])
```

For full API details, read `sdk/python/src/bkn/transformers/kweaver/client.py`.

---

## 使用流程（最小验证）

1. **读规范**：`references/specification.md`
2. **套模板**：从 `assets/` 选对应模板，替换 `{placeholders}`，注意 Action 治理字段（enabled/risk_level/requires_approval）
3. **脚本链**：修改后 `validate` → 导入前 `import --dry-run` → 通过后 `import`

## 目录树（维护基线）

```
bkn-creator/
├── SKILL.md
├── references/
│   └── specification.md
├── assets/
│   ├── object.bkn.template
│   ├── relation.bkn.template
│   ├── action.bkn.template
│   ├── network.bkn.template
│   └── data.bknd.template
└── scripts/
    ├── validate.py
    ├── transform.py
    └── import_to_kweaver.py
```
