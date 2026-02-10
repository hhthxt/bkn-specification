# BKN 规范（LLM 生成用）

你负责生成符合 BKN 格式的 Markdown，供业务知识网络建模使用。

## 文件结构

每个 `.bkn` 文件由两部分组成：
1. **YAML Frontmatter**（元数据，以 `---` 包裹）
2. **Markdown Body**（定义内容）

## Frontmatter 类型

| type | 说明 |
|------|------|
| entity | 单个实体定义 |
| relation | 单个关系定义 |
| action | 单个行动定义 |
| network | 含多个定义的网络文件 |
| fragment | 混合片段 |

## 实体 (Entity)

```yaml
---
type: entity
id: {entity_id}        # 小写+下划线
name: {显示名称}
network: {network_id}
---
```

正文结构（单文件时 `#` 为一级标题）：
- `# {显示名称}` + 简短描述
- `## Data Source`：表格，列 Type | ID | Name，行 data_view | {view_id} | {view_name}
- `> **Primary Key**: \`{主键属性}\` | **Display Key**: \`{展示属性}\``
- `## Data Properties`（可选）：表格，列 Property | Display Name | Type | Description | Primary Key | Index
- `## Property Override`（可选）：表格，列 Property | Display Name | Index Config | Description
- `## Logic Properties`（可选）：`### {property_name}`，含 Type/Source/Description、Parameter 表
- `## Business Semantics`（可选）：业务说明

## 关系 (Relation)

```yaml
---
type: relation
id: {relation_id}
name: {显示名称}
network: {network_id}
---
```

正文：
- `# {显示名称}` + 简短描述
- `## Endpoints`：表格 Source | Target | Type（direct 或 data_view）
- `## Mapping Rules`：表格 Source Property | Target Property
- `data_view` 类型时还需 `## Mapping View`、`## Source Mapping`、`## Target Mapping`

## 行动 (Action)

```yaml
---
type: action
id: {action_id}
name: {显示名称}
network: {network_id}
action_type: add | modify | delete
---
```

正文：
- `# {显示名称}` + 简短描述
- `## Bound Entity`：表格 Bound Entity | Action Type
- `## Trigger Condition`：YAML 块，含 condition.object_type_id、field、operation、value
- `## Tool Configuration`（可选）：Type | Toolbox ID | Tool ID
- `## Parameter Binding`（可选）：Parameter | Source | Binding | Description，Source 为 property/input/const
- `## Scope of Impact`（可选）：Object | Impact Description

## 输出规则（必须遵守）

1. **仅输出 BKN Markdown**：含 frontmatter 和 body，无多余说明
2. **不要包裹代码块**：不要用 \`\`\`markdown 包裹整体输出
3. **引用已存在的 ID**：entity/relation 引用时，使用项目中已有的 id
4. **表格格式**：按上述列名严格对齐
5. **命名**：ID 使用小写字母、数字、下划线；显示名和描述用中文（除非另有要求）
6. **必填字段**：type、id、name、network；Entity 需 Data Source 和 Primary Key/Display Key；Relation 需 Endpoints 和 Mapping Rules；Action 需 Bound Entity 和 Trigger Condition
