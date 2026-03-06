# BKN `.md` 兼容示例

本目录演示 BKN 以 `.md` 为载体时的运行时支持。根文件与 includes 均可使用 `.md`，检测与使用必须满足 BKN frontmatter/type/结构约束。

## 正向验证

### 1. index.md + includes .md

```bash
python .cursor/skills/bkn-creator/scripts/validate.py examples/md-compat/index.md
```

预期：成功加载，输出 type、id、name 及 objects/relations/actions 数量。

### 2. .bkn root include .md

将 `index.md` 重命名为 `index.bkn`，或新建 `index.bkn` 其 includes 含 `objects.md`，同样可成功。

### 3. type: data 的 .md

若 includes 中含 `data/scenario.md`（frontmatter 含 `type: data`、`object` 或 `relation`，正文为标题+表格），可成功加载。

## 负向验证（预期失败）

### 1. .md 无 frontmatter

创建仅含普通 Markdown 的 `.md`，无 `---` 包裹的 YAML：

```bash
echo "# 普通文档" > /tmp/plain.md
python .cursor/skills/bkn-creator/scripts/validate.py /tmp/plain.md
```

预期：`BKN file must have YAML frontmatter with a valid type`

### 2. .md 有 frontmatter 但 type 缺失

```bash
echo "---
id: x
name: 测试
---
## Object: x" > /tmp/no_type.md
python .cursor/skills/bkn-creator/scripts/validate.py /tmp/no_type.md
```

预期：`BKN frontmatter must include a valid 'type' field`

### 3. type: data 的 .md 不满足数据表结构

frontmatter 含 `type: data` 但缺少 `object`/`relation`，或正文无有效表格：

预期：`type: data frontmatter must have exactly one of object or relation` 或 `type: data body must have a valid GFM table`

### 4. 不支持的文件扩展名

```bash
python .cursor/skills/bkn-creator/scripts/validate.py some_file.txt
```

预期：`Unsupported file extension: '.txt'. BKN supports: .bknd, .bkn, .md`

### 5. .md 有扩展名但无 BKN 内容

```bash
python .cursor/skills/bkn-creator/scripts/validate.py README.md
```

若 `README.md` 无 BKN frontmatter，预期：`BKN file must have YAML frontmatter with a valid type`

## 推荐实践

- schema 优先 `.bkn`
- data 优先 `.bknd`
- `.md` 用于需与通用 Markdown 工具共存的场景
