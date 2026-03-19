# BKN SDK

Official SDKs for parsing, validating, and transforming [BKN (Business Knowledge Network)](../docs/SPECIFICATION.md) files.

- **中文** [README.zh.md](README.zh.md)

## Available SDKs

| Language | Status | Path |
|----------|--------|------|
| **Python** | Available | [sdk/python/](python/) |
| **Golang** | Available | [sdk/golang/](golang/) |
| **TypeScript** | Available | [sdk/typescript/](typescript/) |

## Python SDK Usage

```bash
# Install from PyPI (distribution name kweaver-bkn; import package bkn)
pip install kweaver-bkn

# Or editable from this repo
cd sdk/python
pip install -e .

# Run tests
python -m pytest tests/ -v
```

### Quick Example

```python
from bkn import load_network
from bkn.transformers import KweaverTransformer

network = load_network("examples/supplychain-hd/supplychain.bkn")

transformer = KweaverTransformer(id_prefix="supplychain_")
payload = transformer.to_json(network)   # Get kweaver import JSON
transformer.to_files(network, "output/")  # Or write to files
```

See [python/README.md](python/README.md) for detailed usage and API.

## TypeScript SDK Usage

```bash
npm install @kweaver-ai/bkn
```

See [typescript/README.md](typescript/README.md).

## Golang SDK

```bash
cd sdk/golang
go test ./bkn/... -v
```

See [golang/README.md](golang/README.md) for usage and API.

PyPI / npm release workflows live under [.github/workflows](../.github/workflows) (`publish-pypi.yml`, `publish-npm.yml`). Go remains `go get github.com/kweaver-ai/bkn-specification/sdk/golang`.
