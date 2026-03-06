---
type: network
id: md-compat-demo
name: MD 兼容示例
version: 1.0.0
includes:
  - objects.md
---

# MD 兼容示例

本示例演示 BKN 以 `.md` 为载体时的加载与校验。根文件与 includes 均可使用 `.md`。

## 验证

```bash
python .cursor/skills/bkn-creator/scripts/validate.py examples/md-compat/index.md
```
