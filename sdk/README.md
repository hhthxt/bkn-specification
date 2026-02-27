# BKN SDK

Official SDKs for parsing, validating, and transforming [BKN (Business Knowledge Network)](../docs/SPECIFICATION.md) files.

- **中文** [README.zh.md](README.zh.md)

## Available SDKs

| Language | Status | Path |
|----------|--------|------|
| **Python** | Available | [sdk/python/](python/) |
| **Golang** | Planned | [sdk/golang/](golang/) |

## Python SDK Usage

```bash
# Install
cd sdk/python
pip install -e .

# Run tests
python -m pytest tests/ -v
```

### Quick Example

```python
from bkn import load_network
from bkn.transformers import KweaverTransformer

network = load_network("docs/examples/supplychain-hd/supplychain.bkn")

transformer = KweaverTransformer(id_prefix="supplychain_")
payload = transformer.to_json(network)   # Get kweaver import JSON
transformer.to_files(network, "output/")  # Or write to files
```

See [python/README.md](python/README.md) for detailed usage and API.

## Golang SDK

Coming soon. See [golang/README.md](golang/README.md).
