# BKN Python SDK

Parse, validate, and transform [BKN (Business Knowledge Network)](../../docs/SPECIFICATION.md) files.

- **中文** [README.zh.md](README.zh.md)

## Install

```bash
# From repo root
cd sdk/python
pip install -e .

# Or with path
pip install -e path/to/bkn-specification/sdk/python

# With kweaver API support
pip install -e ".[api]"
```

Requires Python 3.9+.

## Usage

### 1. Parse a single file

```python
from bkn import load

doc = load("examples/supplychain-hd/objects.bkn")
print(doc.frontmatter.type)   # fragment
print(len(doc.objects))      # 12
for e in doc.objects:
    print(e.id, e.name, len(e.data_properties), "properties")
```

### 2. Load a network (with includes)

The root file references sub-files via `includes`; `load_network` resolves them recursively:

```python
from bkn import load_network

network = load_network("examples/supplychain-hd/supplychain.bkn")

print(network.root.frontmatter.name)   # HD供应链业务知识网络_v2
print(len(network.all_objects))       # 12
print(len(network.all_relations))     # 14
print(len(network.all_actions))       # 0
```

### 3. Transform to kweaver JSON

Convert BKN models to JSON for ontology-manager API (see `ref/ontology_import_openapi_v2.json`):

```python
from bkn import load_network
from bkn.transformers import KweaverTransformer

network = load_network("examples/supplychain-hd/supplychain.bkn")

transformer = KweaverTransformer(
    branch="main",
    base_version="",
    id_prefix="supplychain_",   # Object/relation ID prefix, e.g. po -> supplychain_po
)

# Get JSON dict
payload = transformer.to_json(network)
# payload["knowledge_network"]  - Create knowledge network request body
# payload["object_types"]      - Object types list
# payload["relation_types"]    - Relation types list

# Or write to files
transformer.to_files(network, "output/")
# Creates: output/knowledge_network.json, object_types.json, relation_types.json
```

### 4. Import to kweaver via API

Import a BKN network directly to kweaver ontology-manager. Requires `pip install -e ".[api]"`.

Two API modes: **internal** (default, no token) and **external** (Bearer token).

**Internal API** (inside cluster, account headers only):
```python
from bkn import load_network
from bkn.transformers import KweaverClient, KweaverTransformer

network = load_network("examples/supplychain-hd/supplychain.bkn")
client = KweaverClient(
    base_url="http://ontology-manager-svc:13014",  # or KWEAVER_BASE_URL
    account_id="your_account_id",
    account_type="your_account_type",
    business_domain="your_domain_id",
    internal=True,  # default
)
transformer = KweaverTransformer(id_prefix="supplychain_")
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
# result.knowledge_network_id   -> created knowledge network ID
# result.object_types_created   -> number of object types created
# result.relation_types_created -> number of relation types created
# result.errors                 -> list of errors (if any)
# result.success                -> True if no errors
```

Dry-run mode (transform only, no API calls):

```python
result = client.import_network(network, transformer, dry_run=True)
```

### 5. Parse from string

```python
from bkn import parse, parse_frontmatter, parse_body

text = """
---
type: object
id: my_object
name: My Object
---

## Object: my_object
**My Object** - Example object

### Data Source
| Type | ID | Name |
|------|-----|------|
| data_view | 123 | my_view |

### Data Properties
| Property | Display Name | Type | Primary Key | Display Key |
|----------|--------------|------|:-----------:|:-----------:|
| id | ID | int64 | YES | |
| name | Name | VARCHAR | | YES |
"""
doc = parse(text)
```

### 6. Access object and relation structure

```python
obj = network.all_objects[0]
print(obj.data_source.type, obj.data_source.id)
for dp in obj.data_properties:
    print(dp.property, dp.type, dp.primary_key)
for po in obj.property_overrides:
    print(po.property, po.index_config)   # fulltext(standard) + vector(id)

relation = network.all_relations[0]
ep = relation.endpoints[0]
print(ep.source, "->", ep.target, ep.type)
for mr in relation.mapping_rules:
    print(mr.source_property, "->", mr.target_property)
```

### 6.1 Parse `.bknd` data files

```python
from bkn import load

doc = load("examples/risk/data/risk_scenario.bknd")
table = doc.data_tables[0]
print(table.object_or_relation)  # risk_scenario
print(table.columns)             # ["scenario_id", "name", ...]
print(len(table.rows))           # number of data rows
```

### 6.2 Serialize to `.bknd` (to_bknd)

```python
from bkn import to_bknd, to_bknd_from_table, load

# From structured data
md = to_bknd(object_id="risk_scenario", rows=[{"scenario_id": "s1", "name": "Test", ...}], network="recoverable-network")

# From a parsed DataTable (round-trip)
doc = load("examples/risk/data/risk_scenario.bknd")
table = doc.data_tables[0]
md = table.to_bknd()
```

`load_network()` loads only files listed in frontmatter `includes`; to load `.bknd` files, add them explicitly (e.g. `includes: [data/risk_scenario.bknd]`). Data tables are aggregated in `network.all_data_tables`.

### 7. Risk assessment

Objects and relations tagged with the reserved **`__risk__`** tag in BKN participate in the built-in risk evaluation. The Action model has a runtime/computed property **`risk`**, filled by the risk assessment module based on the current scenario and risk-tagged knowledge.

`evaluate_risk` returns a **`RiskResult`** with three fields: `decision`, `risk_level`, and `reason`.

```python
from bkn import load_network, evaluate_risk, RiskResult

network = load_network("examples/risk/index.bkn")

# Pass risk_rules when you have instance data (e.g. from a graph or API)
rules = [
    {"scenario_id": "sec_t_01", "action_id": "restart_erp", "allowed": False, "risk_level": 5, "reason": "月末封网"},
]
result = evaluate_risk(network, "restart_erp", {"scenario_id": "sec_t_01"}, risk_rules=rules)
print(result.decision)    # "not_allow"
print(result.risk_level)  # 5
print(result.reason)      # "月末封网"
```

**Custom evaluator** — inject your own logic:

```python
def my_evaluator(network, action_id, context, risk_rules=None, **kwargs):
    if action_id == "grant_root_admin":
        return RiskResult(decision="not_allow", risk_level=5, reason="全局禁止提权")
    return RiskResult(decision="unknown")

result = evaluate_risk(network, "grant_root_admin", {}, evaluator=my_evaluator)
print(result.decision)   # "not_allow"
print(result.risk_level) # 5
```

- **Tagging**: In BKN, add `- **Tags**: __risk__` to definitions that participate in the built-in risk evaluation (reserved tag; users must not use it for custom purposes).
- **RiskResult**: `decision` ("allow" | "not_allow" | "unknown"), `risk_level` (int | None, 0–5 recommended), `reason` (str).
- **Custom evaluator**: Pass `evaluator=my_func` to fully replace the built-in logic; the evaluator must return `RiskResult`.

### Update model (no-patch)

- **Add/modify**: Import `.bkn` files; each definition is upserted by `(network, type, id)`.
- **Delete**: Use the SDK delete API (see plan for `delete` API); deletion is not expressed in BKN files.

## Modules

| Module | Description |
|--------|-------------|
| `bkn.models` | Dataclass models: BknDocument, BknObject, Relation, Action, DataProperty, PropertyOverride, etc. |
| `bkn.parser` | Parsing: parse(), parse_frontmatter(), parse_body(); supports EN/CN table headers |
| `bkn.loader` | Loading: load(path), load_network(root_path); auto-resolves includes |
| `bkn.risk` | Risk assessment: evaluate_risk(...) -> RiskResult; RiskResult(decision, risk_level, reason) |
| `bkn.delete` | Delete API: DeleteTarget, plan_delete(), network_without() |
| `bkn.checksum` | Checksum: generate_checksum_file(), verify_checksum_file() |
| `bkn.transformers.base` | Abstract `Transformer` base class with `to_json()` and `to_files()` interface |
| `bkn.transformers.kweaver` | KweaverTransformer, KweaverClient; outputs kweaver import JSON |

## KweaverTransformer parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `branch` | Branch name | `"main"` |
| `base_version` | Base version string | `""` |
| `id_prefix` | Object/relation ID prefix (e.g. `supplychain_` makes `po` -> `supplychain_po`) | `""` |

## Testing

```bash
cd sdk/python
pip install -e ".[dev]"
python -m pytest tests/ -v
```
