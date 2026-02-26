# BKN Python SDK

Parse, validate, and transform [BKN (Business Knowledge Network)](../../docs/bkn_docs/SPECIFICATION.md) files.

## Install

```bash
pip install -e .
# with kweaver API support:
pip install -e ".[api]"
```

Requires Python 3.9+.

## Quick Start

### Parse a single file

```python
from bkn import load

doc = load("path/to/entities.bkn")
for entity in doc.entities:
    print(entity.id, entity.name, len(entity.data_properties), "properties")
```

### Load a network (with includes)

```python
from bkn import load_network

network = load_network("path/to/supplychain.bkn")
print(f"{len(network.all_entities)} entities, {len(network.all_relations)} relations")
```

### Transform to kweaver JSON

```python
from bkn import load_network
from bkn.transformers import KweaverTransformer

network = load_network("path/to/supplychain.bkn")

transformer = KweaverTransformer(
    branch="main",
    id_prefix="supplychain_hd0202_",
)

# Get JSON dict
payload = transformer.to_json(network)
print(payload["knowledge_network"])
print(f"{len(payload['object_types'])} object types")
print(f"{len(payload['relation_types'])} relation types")

# Or write to files
transformer.to_files(network, "output/")
```

## Modules

| Module | Description |
|--------|-------------|
| `bkn.models` | Dataclass models: `BknDocument`, `Entity`, `Relation`, `Action`, etc. |
| `bkn.parser` | Parse `.bkn` text: frontmatter (YAML) + body (Markdown sections/tables) |
| `bkn.loader` | Load from filesystem, resolve `includes` for network files |
| `bkn.transformers.kweaver` | Convert BKN models to kweaver ontology-manager API JSON |

## Testing

```bash
pip install -e ".[dev]"
python -m pytest tests/ -v
```
