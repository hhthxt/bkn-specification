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

doc = load("docs/examples/supplychain-hd/entities.bkn")
print(doc.frontmatter.type)   # fragment
print(len(doc.entities))      # 12
for e in doc.entities:
    print(e.id, e.name, len(e.data_properties), "properties")
```

### 2. Load a network (with includes)

The root file references sub-files via `includes`; `load_network` resolves them recursively:

```python
from bkn import load_network

network = load_network("docs/examples/supplychain-hd/supplychain.bkn")

print(network.root.frontmatter.name)   # HD供应链业务知识网络_v2
print(len(network.all_entities))      # 12
print(len(network.all_relations))     # 14
print(len(network.all_actions))       # 0
```

### 3. Transform to kweaver JSON

Convert BKN models to JSON for ontology-manager API (see `ref/ontology_import_openapi_v2.json`):

```python
from bkn import load_network
from bkn.transformers import KweaverTransformer

network = load_network("docs/examples/supplychain-hd/supplychain.bkn")

transformer = KweaverTransformer(
    branch="main",
    base_version="",
    id_prefix="supplychain_",   # Entity/relation ID prefix, e.g. po -> supplychain_po
)

# Get JSON dict
payload = transformer.to_json(network)
# payload["knowledge_network"]  - Create knowledge network request body
# payload["object_types"]      - Object types (entities) list
# payload["relation_types"]    - Relation types list

# Or write to files
transformer.to_files(network, "output/")
# Creates: output/knowledge_network.json, object_types.json, relation_types.json
```

### 4. Import to kweaver via API

Import a BKN network directly to kweaver ontology-manager. Requires `pip install -e ".[api]"`:

```python
from bkn import load_network
from bkn.transformers import KweaverClient, KweaverTransformer

network = load_network("docs/examples/supplychain-hd/supplychain.bkn")

client = KweaverClient(
    base_url="http://ontology-manager-svc:13014",
    account_id="your_account_id",
    account_type="your_account_type",
    business_domain="your_domain_id",
)

transformer = KweaverTransformer(id_prefix="supplychain_")
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
type: entity
id: my_entity
name: My Entity
---

## Entity: my_entity
**My Entity** - Example entity

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

### 6. Access entity and relation structure

```python
entity = network.all_entities[0]
print(entity.data_source.type, entity.data_source.id)
for dp in entity.data_properties:
    print(dp.property, dp.type, dp.primary_key)
for po in entity.property_overrides:
    print(po.property, po.index_config)   # fulltext(standard) + vector(id)

relation = network.all_relations[0]
ep = relation.endpoints[0]
print(ep.source, "->", ep.target, ep.type)
for mr in relation.mapping_rules:
    print(mr.source_property, "->", mr.target_property)
```

## Modules

| Module | Description |
|--------|-------------|
| `bkn.models` | Dataclass models: BknDocument, Entity, Relation, Action, DataProperty, PropertyOverride, etc. |
| `bkn.parser` | Parsing: parse(), parse_frontmatter(), parse_body(); supports EN/CN table headers |
| `bkn.loader` | Loading: load(path), load_network(root_path); auto-resolves includes |
| `bkn.transformers.base` | Abstract `Transformer` base class with `to_json()` and `to_files()` interface |
| `bkn.transformers.kweaver` | KweaverTransformer, KweaverClient; outputs kweaver import JSON |

## KweaverTransformer parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `branch` | Branch name | `"main"` |
| `base_version` | Base version string | `""` |
| `id_prefix` | Entity/relation ID prefix (e.g. `supplychain_` makes `po` -> `supplychain_po`) | `""` |

## Testing

```bash
cd sdk/python
pip install -e ".[dev]"
python -m pytest tests/ -v
```
