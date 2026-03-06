# BKN 规范（LLM 生成用）

你负责生成符合 BKN 格式的 Markdown，供业务知识网络建模使用。

## 文件扩展名与载体

- `.bkn` / `.bknd`：推荐扩展名
- `.md`：兼容载体，可用 `.md` 保存；内容必须满足 BKN frontmatter、`type` 及结构约束，否则加载时报错

## 文件结构

每个 BKN 文件（`.bkn` / `.bknd` / `.md`）由两部分组成：
1. **YAML Frontmatter**（元数据，以 `---` 包裹）
2. **Markdown Body**（定义内容）

## Frontmatter 类型

| type | 说明 |
|------|------|
| object | 单个对象定义 |
| relation | 单个关系定义 |
| action | 单个行动定义 |
| network | 含多个定义的网络文件 |
| fragment | 混合片段 |
| data | 数据文件（建议 `.bknd`，承载对象/关系实例行） |

## `.bknd` 数据文件（type: data）

用于承载**知识原生**实例数据，不承载 schema 定义。仅当 Object 的 Data Source 为 `bknd` 或未指定 data_view 时，才使用 `.bknd` 维护数据；Data Source 为 `data_view` 的对象数据来自外部系统，不可用 `.bknd` 编辑。

Frontmatter 示例：

```yaml
---
type: data
network: recoverable-network
object: scenario   # 或 relation: rs_under_scenario（二选一）
source: PFMEA模板.xlsx   # 可选，数据来源
---
```

正文使用「一个标题 + 一个 Markdown 表格」，列名需与目标对象 Data Properties（或关系映射字段）保持一致。

## 对象 (Object)

```yaml
---
type: object
id: {object_id}        # 小写+下划线
name: {显示名称}
network: {network_id}
---
```

正文结构（与 network/fragment 内嵌定义保持一致）：
- `## Object: {object_id}`
- `**{显示名称}**` + 简短描述
- （可选）定义级元数据：`- **Tags**: tag1, tag2`、`- **Owner**: owner`
- `### Data Source`：表格，列 Type | ID | Name，行 data_view | {view_id} | {view_name}
- `### Data Properties`（必须）：表格，列 Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index
  - 至少标记一个 `Primary Key: YES`（主键）和一个 `Display Key: YES`（展示键）
  - 最简形式可只含 Property | Primary Key | Display Key 三列
- Type 列标准类型：int32, int64, integer, float32, float64, float, decimal(p,s), decimal, bool, VARCHAR, TEXT, DATE, TIME, TIMESTAMP, JSON, BINARY；不在列表中的类型透传
- `### Property Override`（可选）：表格，列 Property | Display Name | Index Config | Constraint | Description
  - Index Config 语法：`keyword`、`keyword(max_len)`、`fulltext`、`fulltext(analyzer)`、`vector`、`vector(model_id)`，可组合如 `keyword(1024) + fulltext(standard) + vector(model_id)`
- Constraint 列语法：`operator` / `operator(args)` / `operator value`；多条用 `; ` 组合（AND）
  - 比较：`== v`、`!= v`、`> v`、`< v`、`>= v`、`<= v`
  - 范围：`range(min,max)`
  - 枚举：`in(v1,v2,...)`、`not_in(v1,v2,...)`
  - 存在性：`not_null`、`exist`、`not_exist`
  - 正则：`regex:pattern`
  - 示例：`not_null; >= 0`、`in(Running,Pending,Failed)`、`regex:^[a-z0-9_]+$`
- `### Logic Properties`（可选）：`#### {property_name}`，含 Type/Source/Description、Parameter 表；Parameter 表列为 Parameter | Type | Source | Binding | Description
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
- （可选）定义级元数据：`- **Tags**: tag1, tag2`、`- **Owner**: owner`
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
# 治理字段（可选）
enabled: false          # 是否启用，默认 false；导入不等于启用
risk_level: low | medium | high
requires_approval: true  # 是否需要审批才能启用/执行
---
```

**治理字段复用策略**（不改 BKN 元素类型，仅消费 Action 现有字段）：

| 字段 | 执行时含义 |
|------|------------|
| `enabled != true` | 默认不可执行，返回「需先启用」 |
| `risk_level == high` | 强制二次确认，建议先 dry-run |
| `requires_approval == true` | 进入审批路径，不直接执行 |
| 字段缺失 | 按保守策略处理（视为高风险待确认） |

运行时属性 `risk`（`allow` | `not_allow` | `unknown`）为计算属性，由风险评估函数根据场景与带 `__risk__` tag 的数据计算，**不写入 BKN 文件**。

正文：
- `## Action: {action_id}`
- `**{显示名称}**` + 简短描述
- `### Bound Object`：表格 Bound Object | Action Type
- `### Trigger Condition`：YAML 块，含 condition.object_type_id、field、operation、value
- `### Pre-conditions`（可选）：表格 Object | Check | Condition | Message，Check 为 property:{name} 或 relation:{id}
- `### Tool Configuration`（可选）：Type | Toolbox ID | Tool ID
- `### Parameter Binding`（可选）：Parameter | Type | Source | Binding | Description，Source 为 property/input/const
- `### Scope of Impact`（可选）：Impacted Object | Impact Description

## 输出规则（必须遵守）

1. **仅输出 BKN Markdown**：含 frontmatter 和 body，无多余说明
2. **不要包裹代码块**：不要用 \`\`\`markdown 包裹整体输出
3. **引用已存在的 ID**：object/relation 引用时，使用项目中已有的 id
4. **表格格式**：按上述列名严格对齐
5. **命名**：ID 使用小写字母、数字、下划线；显示名和描述用中文（除非另有要求）
6. **必填字段**：type、id、name、network；Object 需 Data Source 和 Primary Key/Display Key；Relation 需 Endpoints 和 Mapping Rules；Action 需 Bound Object 和 Trigger Condition
