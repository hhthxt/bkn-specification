---
name: bkn-creator
description: Generate BKN (Business Knowledge Network) files for modeling objects, relations, and actions. Optionally import to kweaver via API. Use when the user asks to create a BKN file, define a knowledge network, model objects/relations, generate .bkn files, import to kweaver, or run BKN/SDK scripts.
---

# BKN Creator

Generate `.bkn` files conforming to the BKN specification, and optionally import them to kweaver.

## Workflow

**顺序**：先读规范 → 再选模板 → 再生成输出。

1. **理解需求**：识别业务域、对象、关系、行动
2. **读规范**：加载 [references/specification.md](references/specification.md)，按格式规则生成
3. **选模板**：从 [assets/](assets/) 选取对应类型（object、relation、action、network、data）
4. **生成 `.bkn` 文件**：按下方 Output Rules 输出
5. **（可选）校验/导入**：运行 `scripts/` 中的 validate、transform 或 import_to_kweaver

## Scripts

Scripts live in [scripts/](scripts/). Run from repo root. Install first: `pip install -e sdk/python` or `pip install -e "sdk/python[api]"` for import.

**validate.py** — Check BKN loads:
```bash
python .cursor/skills/bkn-creator/scripts/validate.py <path>
# e.g. python .cursor/skills/bkn-creator/scripts/validate.py examples/k8s-modular/index.bkn
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

## File Organization

Choose an organization style based on network size:

**Modular** (recommended for large networks, team collaboration):
Each object/relation/action in its own file, with an `index.bkn` root.
See `examples/k8s-modular/` for this pattern.

```
my-network/
├── index.bkn
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
├── index.bkn
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

1. Output **only** valid BKN Markdown (frontmatter + body). No extra explanation around the file content.
2. Do **not** wrap the entire output in a code fence.
3. Use existing object/relation IDs when referencing other definitions in the same network.
4. Follow table column names exactly as defined in the spec.
5. IDs: lowercase letters, digits, underscores. Display names and descriptions in Chinese unless specified otherwise.
6. Required fields: `type`, `id`, `name`, `network` in frontmatter.
   - Object: must have Data Source + at least one Primary Key and one Display Key.
   - Relation: must have Endpoints + Mapping Rules.
   - Action: must have Bound Object + Trigger Condition.

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
2. **套模板**：从 `assets/` 选对应模板，替换 `{placeholders}`
3. **可选脚本**：`scripts/validate.py` 校验、`scripts/transform.py` 导出、`scripts/import_to_kweaver.py` 导入

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
