# BKN SDK

解析、校验与转换 [BKN 业务知识网络](../docs/SPECIFICATION.md) 的官方 SDK。

- **English** [README.md](README.md)

## 可用 SDK

| 语言 | 状态 | 路径 |
|------|------|------|
| **Python** | 可用 | [sdk/python/](python/) |
| **Golang** | 可用 | [sdk/golang/](golang/) |
| **TypeScript** | 可用 | [sdk/typescript/](typescript/) |

## Python SDK 用法

```bash
# 从 PyPI 安装（发行包名 kweaver-bkn，import 仍为 bkn）
pip install kweaver-bkn

# 或从本仓库可编辑安装
cd sdk/python
pip install -e .

# 运行测试
python -m pytest tests/ -v
```

### 快速示例

```python
from bkn import load_network
from bkn.transformers import KweaverTransformer

network = load_network("examples/supplychain-hd/supplychain.bkn")

transformer = KweaverTransformer(id_prefix="supplychain_")
payload = transformer.to_json(network)   # 获取 kweaver 导入 JSON
transformer.to_files(network, "output/")  # 或写入文件
```

详细用法与 API 见 [python/README.md](python/README.md)。

## TypeScript SDK 用法

```bash
npm install @kweaver-ai/bkn
```

详见 [typescript/README.md](typescript/README.md)。

## Golang SDK

```bash
cd sdk/golang
go test ./bkn/... -v
```

详见 [golang/README.md](golang/README.md)。

PyPI / npm 发布流程见仓库 [.github/workflows](../.github/workflows)（`publish-pypi.yml`、`publish-npm.yml`）。Go 仍使用 `go get github.com/kweaver-ai/bkn-specification/sdk/golang`。
