# BKN 规范（LLM 生成用）

你负责生成符合 BKN 格式的 Markdown，供业务知识网络建模使用。

## 文件扩展名与载体

- `.bkn` / `.bknd`：推荐扩展名
- `.md`：兼容载体，可用 `.md` 保存；内容必须满足 BKN frontmatter、`type` 及结构约束，否则加载时报错

## 根文件与目录加载

- **根文件命名**：推荐 `network.bkn` 作为网络入口；`index.bkn` 兼容，优先级低于 `network.bkn`
- **目录输入**：`validate network <dir>`、`load_network(dir)` 等支持传入目录，自动发现根文件（顺序：network.bkn > network.md > index.bkn > index.md）
- **无 includes**：当根文件为 `type: network` 且未声明 `includes` 时，同目录下所有 BKN 文件视为同一网络输入；有 `includes` 则按 `includes` 加载

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
| risk | 单个风险定义 |
| network | 含多个定义的网络文件 |
| fragment | 混合片段 |
| data | 数据文件，承载 object/relation 的实例数据行（建议 `.bknd`） |
| connection | 可复用的数据源连接定义（可选；多对象共享同一数据源时使用） |

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
- `### Data Source`：表格，列 Type | ID | Name，行 data_view | {view_id} | {view_name}、connection | {connection_id} | {display_name} 或 bknd | {object_id} | {display_name}。仅当 Data Source 为 `bknd` 时可为该对象创建 `.bknd` 数据文件；`data_view` 或 `connection` 时不可用 `.bknd`
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
---
```

正文：
- `## Action: {action_id}`
- `**{显示名称}**` + 简短描述
- `### Bound Object`：表格 Bound Object | Action Type
- `### Trigger Condition`：YAML 块，含 field、operation、value（field 为属性名，operation 为 ==/!=/>/</>=/<=/in/not_in/exist/not_exist，value 为比较值）
- `### Pre-conditions`（可选）：表格 Object | Check | Condition | Message，Check 为 property:{name} 或 relation:{id}
- `### Tool Configuration`（必须）：Type | Tool ID 或 Type | MCP。tool 时列 Type | Tool ID；mcp 时列 Type | MCP
- `### Parameter Binding`（必须）：Parameter | Type | Source | Binding | Description，Source 为 property/input/const
- `### Scope of Impact`（可选）：Object | Impact Description

## 风险 (Risk)

风险定义独立于 action。正文结构：`## Risk: {risk_id}`、`### 管控范围`、`### 管控策略`、`### 前置检查`（可选）、`### 回滚方案`（可选）、`### 审计要求`（可选）。

## 连接定义 (type: connection)（可选）

当多个对象共享同一数据源连接时，可定义 `type: connection` 文件。正文结构：`## Connection: {connection_id}`、`### Connection` 表格（列 Type | Endpoint | Secret Ref）。**不得**在 BKN 中写入明文凭据，仅使用 `secret_ref` 或环境变量引用。Object 的 Data Source 可写 `connection | {connection_id}` 引用该连接。

## 数据文件 (type: data / .bknd)

仅当 Object 的 Data Source 为 `bknd` 时，可为该对象创建 `.bknd` 数据文件。Data Source 为 `data_view` 或 `connection` 的对象数据来自外部系统，**不要**为其生成 `.bknd`。Frontmatter 示例：`type: data`、`network`、`object` 或 `relation`（二选一）。正文为标题 + 一个表格，列名与目标 object 的 Data Properties 一致。

## 更新与删除（无 patch 模型）

- 定义文件导入 = add/modify（upsert）；修改即编辑文件后重新导入
- 删除元素通过 SDK/CLI delete API 执行，不通过 BKN 文件；**不要生成 type: delete 或 type: patch 文件**

## 输出规则（必须遵守）

1. **仅输出 BKN Markdown**：含 frontmatter 和 body，无多余说明
2. **不要包裹代码块**：不要用 \`\`\`markdown 包裹整体输出
3. **引用已存在的 ID**：object/relation 引用时，使用项目中已有的 id
4. **表格格式**：按上述列名严格对齐
5. **命名**：ID 使用小写字母、数字、下划线；显示名和描述用中文（除非另有要求）
6. **必填字段**：type、id、name、network；Object 需 Data Source 和 Primary Key/Display Key；Relation 需 Endpoints 和 Mapping Rules；Action 需 Bound Object、Trigger Condition、Tool Configuration 和 Parameter Binding
