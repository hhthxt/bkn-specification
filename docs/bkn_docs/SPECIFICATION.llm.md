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

正文结构（与 network/fragment 内嵌定义保持一致）：
- `## Entity: {entity_id}`
- `**{显示名称}**` + 简短描述
- `### Data Source`：表格，列 Type | ID | Name，行 data_view | {view_id} | {view_name}
- `### Data Properties`（必须）：表格，列 Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index
  - 至少标记一个 `Primary Key: YES`（主键）和一个 `Display Key: YES`（展示键）
  - 最简形式可只含 Property | Primary Key | Display Key 三列
- Type 列标准类型：int32, int64, float32, float64, decimal(p,s), bool, VARCHAR, TEXT, DATE, TIME, TIMESTAMP, JSON, BINARY；不在列表中的类型透传
- `### Property Override`（可选）：表格，列 Property | Display Name | Index Config | Constraint | Description
- Constraint 列语法：`operator` / `operator(args)` / `operator value`；多条用 `; ` 组合（AND）
  - 比较：`== v`、`!= v`、`> v`、`< v`、`>= v`、`<= v`
  - 范围：`range(min,max)`
  - 枚举：`in(v1,v2,...)`、`not_in(v1,v2,...)`
  - 存在性：`not_null`、`exist`、`not_exist`
  - 正则：`regex:pattern`
  - 示例：`not_null; >= 0`、`in(Running,Pending,Failed)`、`regex:^[a-z0-9_]+$`
- `### Logic Properties`（可选）：`#### {property_name}`，含 Type/Source/Description、Parameter 表
- `### Business Semantics`（可选）：业务说明

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
- `## Relation: {relation_id}`
- `**{显示名称}**` + 简短描述
- `### Endpoints`：表格 Source | Target | Type（direct 或 data_view），可选列 Required | Min | Max（基数约束）
- `### Mapping Rules`：表格 Source Property | Target Property
- `data_view` 类型时还需 `### Mapping View`、`### Source Mapping`、`### Target Mapping`

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
- `## Action: {action_id}`
- `**{显示名称}**` + 简短描述
- `### Bound Entity`：表格 Bound Entity | Action Type
- `### Trigger Condition`：YAML 块，含 condition.object_type_id、field、operation、value
- `### Pre-conditions`（可选）：表格 Entity | Check | Condition | Message，Check 为 property:{name} 或 relation:{id}
- `### Tool Configuration`（可选）：Type | Toolbox ID | Tool ID
- `### Parameter Binding`（可选）：Parameter | Source | Binding | Description，Source 为 property/input/const
- `### Scope of Impact`（可选）：Object | Impact Description

## 输出规则（必须遵守）

1. **仅输出 BKN Markdown**：含 frontmatter 和 body，无多余说明
2. **不要包裹代码块**：不要用 \`\`\`markdown 包裹整体输出
3. **引用已存在的 ID**：entity/relation 引用时，使用项目中已有的 id
4. **表格格式**：按上述列名严格对齐
5. **命名**：ID 使用小写字母、数字、下划线；显示名和描述用中文（除非另有要求）
6. **必填字段**：type、id、name、network；Entity 需 Data Source 和 Primary Key/Display Key；Relation 需 Endpoints 和 Mapping Rules；Action 需 Bound Entity 和 Trigger Condition
