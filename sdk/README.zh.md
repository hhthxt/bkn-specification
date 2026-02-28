# BKN SDK

解析、校验与转换 [BKN 业务知识网络](../docs/SPECIFICATION.md) 的官方 SDK。

- **English** [README.md](README.md)

## 可用 SDK

| 语言 | 状态 | 路径 |
|------|------|------|
| **Python** | 可用 | [sdk/python/](python/) |
| **Golang** | 规划中 | [sdk/golang/](golang/) |

## Python SDK 用法

```bash
# 安装
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

## Golang SDK

规划中，见 [golang/README.md](golang/README.md)。
