# BKN 语言规范

版本: 2.0.0

## 概述

BKN (Business Knowledge Network) 是一种基于 Markdown 的声明式建模语言，用于定义业务知识网络中的对象、关系和行动。BKN 只负责描述模型结构与语义，不包含执行逻辑——校验引擎、数据管道、工作流等运行时能力由消费 BKN 模型的平台实现。

本文档定义了 BKN 的完整语法规范。

### 术语表（Glossary）

**核心概念**

| 术语 | 含义 |
|------|------|
| BKN | Business Knowledge Network，业务知识网络 |
| knowledge_network | 一个业务知识网络的整体集合 |
| object_type | 业务对象类型（例如 Pod/Node/Service） |
| relation_type | 连接两个 object_type 的关系类型（例如 belongs_to/routes_to） |
| action_type | 对 object_type 执行的操作定义（可绑定 tool 或 mcp） |
| risk_type | 风险类，对行动类和对象类的执行风险进行结构化建模 |
| concept_group | 概念分组，用于将相关对象类型组织在一起 |

**对象结构**

| 术语 | 含义 |
|------|------|
| data_view | 数据视图，对象/关系可直接映射的数据来源 |
| connection | 可复用的数据源连接定义；多个对象可引用同一 connection，共享连接信息 |
| data_properties | 对象的属性定义表，声明字段名称、类型、描述 |
| keys | 键定义，声明主键、展示键、增量键 |
| logic_properties | 逻辑属性，基于其他数据源的派生字段（metric / operator） |
| primary_key | 主键字段，用于唯一定位实例（Keys 小节中声明） |
| display_key | 展示字段，用于 UI 显示和检索（Keys 小节中声明） |
| metric | 逻辑属性类型：指标，从外部数据源获取的度量值 |
| operator | 逻辑属性类型：算子，基于输入参数的计算逻辑 |

**行动结构**

| 术语 | 含义 |
|------|------|
| trigger_condition | 触发条件，定义 action 自动执行的条件 |
| pre-conditions | 前置条件，执行前必须满足的数据检查（不满足则阻止执行） |
| tool | 行动绑定的外部工具 |
| mcp | Model Context Protocol，行动绑定的 MCP 工具 |
| schedule | 定时配置（FIX_RATE 或 CRON），用于周期性执行 |
| scope_of_impact | 影响范围，声明行动影响的对象 |

**文件组织**

| 术语 | 含义 |
|------|------|
| frontmatter | YAML 元数据区（`---` 包裹），每个 .bkn 文件的头部 |
| network | 文件类型 `type: network`，完整知识网络的顶层容器 |
| data | 文件类型 `type: data`，实例数据文件（建议使用 `.bknd` 扩展名） |

### 标准原语表 (Primitives)

Section 标题和表格列名的规范形式，建议使用英文。解析器应同时支持英文与中文以便兼容。

下表按 **统一标题层级** 组织，适用于所有文件类型（network / object_type / relation_type / action_type / risk_type / concept_group / data）。

| Level | English (canonical) | Definition | 中文 | Syntax |
|:-----:|---------------------|------------|------|--------|
| `#` | {Network Name} | Network title | 网络标题 | `# {name}` |
| `##` | Network Overview | Network topology overview | 网络概览 | `## Network Overview` |
| `##` | ObjectType | Individual object type definition | 对象类型 | `## ObjectType: {name}` |
| `##` | RelationType | Individual relation type definition | 关系类型 | `## RelationType: {name}` |
| `##` | ActionType | Individual action type definition | 行动类型 | `## ActionType: {name}` |
| `##` | RiskType | Individual risk type definition | 风险类型 | `## RiskType: {name}` |
| `##` | ConceptGroup | Concept group definition | 概念分组 | `## ConceptGroup: {name}` |
| `###` | Data Source | The data view this object maps from | 数据来源 | `### Data Source` |
| `###` | Data Properties | Explicit list of fields (name, type, description) | 数据属性 | `### Data Properties` |
| `###` | Keys | Primary key, display key, incremental key | 键定义 | `### Keys` |
| `###` | Logic Properties | Derived fields: metric, operator | 逻辑属性 | `### Logic Properties` |
| `###` | Endpoint | Relation endpoint: source, target, type | 关联定义 | `### Endpoint` |
| `###` | Mapping Rules | How source/target properties map | 映射规则 | `### Mapping Rules` |
| `###` | Mapping View | For data_view relations: the join view | 映射视图 | `### Mapping View` |
| `###` | Source Mapping | Map source object props to view | 起点映射 | `### Source Mapping` |
| `###` | Target Mapping | Map view to target object props | 终点映射 | `### Target Mapping` |
| `###` | Bound Object | Object this action operates on | 绑定对象 | `### Bound Object` |
| `###` | Trigger Condition | When to run (YAML condition) | 触发条件 | `### Trigger Condition` |
| `###` | Pre-conditions | Data conditions required before action execution | 前置条件 | `### Pre-conditions` |
| `###` | Tool Configuration | tool or MCP binding | 工具配置 | `### Tool Configuration` |
| `###` | Parameter Binding | param name, source, binding | 参数绑定 | `### Parameter Binding` |
| `###` | Schedule | FIX_RATE or CRON | 调度配置 | `### Schedule` |
| `###` | Scope of Impact | What objects are affected | 影响范围 | `### Scope of Impact` |
| `###` | Object Types | Object types in a concept group | 对象类型列表 | `### Object Types` |
| `###` | Control Scope | Risk control scope | 管控范围 | `### Control Scope` |
| `###` | Control Policy | Risk control policy | 管控策略 | `### Control Policy` |
| `###` | Pre-checks | Risk pre-checks | 前置检查 | `### Pre-checks` |
| `###` | Rollback Plan | Risk rollback plan | 回滚方案 | `### Rollback Plan` |
| `###` | Audit Requirements | Risk audit requirements | 审计要求 | `### Audit Requirements` |
| `####` | {property_name} | Individual logic property sub-section | — | `#### {name}` |

表格列名（canonical）：Name, Display Name, Type, Description; Source, Target; Source Property, Target Property; Parameter, Source, Binding, Description; Bound Object, Action Type; Object, Check, Condition, Message; Object, Impact Description。解析器同时接受中文列名。

## 文件格式

### 文件扩展名

- `.bkn` - BKN 定义文件（schema），推荐
- `.bknd` - BKN 数据文件（instance data），推荐
- `.md` - 兼容载体，运行时支持；内容必须满足 BKN frontmatter/type/结构约束

**`.md` 兼容模式**：可用 `.md` 保存 BKN 内容，便于跨平台文档化与协作。运行时加载时，`.md` 与 `.bkn/.bknd` 走同一解析与校验路径；若缺少 frontmatter、`type` 或结构不符合要求，将直接报错。推荐实践：schema 优先 `.bkn`，data 优先 `.bknd`，`.md` 用于需与通用 Markdown 工具共存的场景。

### 文件编码

- UTF-8

### 基本结构

每个 BKN 文件由两部分组成：

1. **YAML Frontmatter**: 文件元数据
2. **Markdown Body**: 知识网络定义内容

```markdown
---
type: network
id: example-network
name: Example Network
tags: [example]
---

# Example Network

Network description...

## Network Overview

...
```

---

## Frontmatter 规范

### 文件类型 (type)

| type | 说明 | 用途 |
|------|------|------|
| `network` | 完整知识网络 | 网络顶层容器文件 |
| `object_type` | 单个对象类型定义 | 独立的对象类型文件，可直接导入 |
| `relation_type` | 单个关系类型定义 | 独立的关系类型文件，可直接导入 |
| `action_type` | 单个行动类型定义 | 独立的行动类型文件，可直接导入 |
| `risk_type` | 单个风险类型定义 | 独立的风险类文件，可直接导入 |
| `concept_group` | 概念分组 | 将相关对象类型组织在一起 |
| `data` | 数据文件 | 承载 object/relation 的实例数据行（建议 `.bknd`） |

### 网络文件 (type: network)

```yaml
---
type: network                    # 完整知识网络
id: string                       # 网络ID，唯一标识
name: string                     # 网络显示名称
tags: [string]                   # 可选，标签列表
business_domain: string          # 可选，业务领域
---
```

描述放在 body 中，`# {name}` 标题之后。

### 对象类型文件 (type: object_type)

```yaml
---
type: object_type                # 对象类型定义
id: string                       # 对象ID，唯一标识
name: string                     # 对象显示名称
tags: [string]                   # 可选，标签列表
---
```

### 关系类型文件 (type: relation_type)

```yaml
---
type: relation_type              # 关系类型定义
id: string                       # 关系ID，唯一标识
name: string                     # 关系显示名称
tags: [string]                   # 可选，标签列表
---
```

### 行动类型文件 (type: action_type)

```yaml
---
type: action_type                # 行动类型定义
id: string                       # 行动ID，唯一标识
name: string                     # 行动显示名称
tags: [string]                   # 可选，标签列表
action_type: add | modify | delete | query  # 可选，行动类型
enabled: boolean                 # 可选，是否启用（建议默认 false）
risk_level: low | medium | high  # 可选，静态风险等级
requires_approval: boolean       # 可选，是否需要审批
---
```

### 风险类型文件 (type: risk_type)

```yaml
---
type: risk_type                  # 风险类型定义
id: string                       # 风险类ID，唯一标识
name: string                     # 风险类显示名称
tags: [string]                   # 可选，标签列表
---
```

### 概念分组文件 (type: concept_group)

```yaml
---
type: concept_group              # 概念分组
id: string                       # 分组ID，唯一标识
name: string                     # 分组显示名称
tags: [string]                   # 可选，标签列表
---
```

---

## 数据文件规范（.bknd / type: data）

`.bknd` 文件使用与 `.bkn` 相同的 Markdown 语法，但正文承载的是实例数据表，而非对象/关系/行动定义。

### Frontmatter

```yaml
---
type: data
network: recoverable-network
object: scenario            # 与 relation 二选一
source: PFMEA模板.xlsx      # 可选，数据来源
---
```

- `type` 必须为 `data`
- `object` 或 `relation` 二选一，用于指向对应 `.bkn` 中定义的 ID
- `network` 建议填写，保持与 schema 文件一致
- `source` 为可选字段，用于记录数据来源

### Body

正文由一个标题（`#` 或 `##`）+ 一个 GFM 表格组成。表头应与目标 object 的 Data Properties 列名（或 relation 的映射字段）一致。

```markdown
# scenario

| scenario_id | name | category | primary_object | description |
|-------------|------|----------|----------------|-------------|
| ops-rm-rf | rm -rf 删除备份存储 | integrity | backup_system | 直接销毁备份数据 |
```

### 约束

- 列名应与 schema 定义保持一致，避免隐式字段
- 每个 `type: data` 文件建议只包含一个数据表，便于版本化和审计

### Data Source 与可编辑性

对象的 Data Source 决定其数据是否可写入 `.bknd`：

| Data Source Type | 数据来源 | `.bknd` 是否允许 |
|------------------|----------|------------------|
| `data_view` | 外部系统（ERP、数据库、API） | **否**，数据由外部系统维护，`.bknd` 不适用 |
| `connection` | 通过 connection 定义连接的外部系统 | **否**，数据由外部系统提供，`.bknd` 不适用 |
| `bknd` | BKN 知识原生数据 | **是**，`.bknd` 为数据源，可读写 |
| 无 Data Source | 平台默认 | 视平台实现而定 |

- 当 Object 的 Data Source 为 `data_view` 或 `connection` 时，不应为该对象创建 `.bknd` 文件，数据由外部系统提供
- 当 Object 的 Data Source 为 `bknd` 时，数据存储在 `.bknd` 文件中，可编辑、可版本化

---

## 对象类型定义规范

### 语法

```markdown
## ObjectType: {name}

{description}

### Data Properties

| Name | Display Name | Type | Description |
|------|--------------|------|-------------|
| {prop} | {display_name} | {type} | {desc} |

### Keys

Primary Keys: {key_name}
Display Key: {key_name}
Incremental Key: {key_name}

### Logic Properties

#### {property_name}

- **Type**: metric | operator
- **Source**: {source_id} ({source_type})
- **Description**: {description}

| Parameter | Type | Source | Binding | Description |
|-----------|------|--------|---------|-------------|
| ... | string | property | {property_name} | 从对象属性绑定 |
| ... | array | input | - | 运行时用户输入 |
| ... | string | const | {value} | 常量值 |

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | {view_id} | {view_name} |
| bknd | {object_id} | {display_name} |
```

- `Type`：参数数据类型，如 string、number、boolean、array
- `Source`：值来源，`property`（对象属性）/ `input`（用户输入）/ `const`（常量）
- `Binding`：当 Source 为 property 时为属性名，为 const 时为常量值，为 input 时为 `-`

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| {name} | YES | 对象类型显示名称 |
| Data Properties | YES | 属性定义表 |
| Keys | YES | 主键、展示键声明 |
| Logic Properties | NO | 指标、算子等扩展属性 |
| Data Source | NO | 映射的数据视图，未设定时由平台自动管理 |

### 数据类型

Data Properties 表的 `Type` 列使用以下标准类型。类型名称大小写不敏感，推荐使用下表中的规范形式。

| 类型 | 说明 | JSON 对应 | SQL 对应 |
|------|------|-----------|----------|
| int32 | 32 位有符号整数 | number | INT / INTEGER |
| int64 | 64 位有符号整数 | number | BIGINT |
| integer | 泛型整数（精度未指定） | number | 平台相关（通常 int64） |
| float32 | 32 位浮点数 | number | FLOAT / REAL |
| float64 | 64 位浮点数 | number | DOUBLE / DOUBLE PRECISION |
| float | 泛型浮点数（精度未指定） | number | 平台相关（通常 float64） |
| decimal(p,s) | 精确十进制数，p 为精度，s 为小数位 | string / number | DECIMAL(p,s) / NUMERIC(p,s) |
| decimal | 泛型精确十进制（精度未指定） | string / number | 平台相关 |
| bool | 布尔值 | boolean | BOOLEAN |
| VARCHAR | 变长字符串 | string | VARCHAR / TEXT |
| TEXT | 长文本 | string | TEXT / CLOB |
| DATE | 日期（无时间） | string (ISO 8601) | DATE |
| TIME | 时间（无日期） | string (ISO 8601) | TIME |
| TIMESTAMP | 日期时间（含时区） | string (ISO 8601) | TIMESTAMP |
| JSON | JSON 结构数据 | object / array | JSON / JSONB |
| BINARY | 二进制数据 | string (base64) | BLOB / BYTEA |

> 当数据源使用的类型不在上表中时，可直接使用数据源原生类型名称（如 `ARRAY<VARCHAR>`），解析器应透传不识别的类型。

---

## 关系类型定义规范

### 语法

```markdown
## RelationType: {name}

{description}

### Endpoint

| Source | Target | Type |
|--------|--------|------|
| {source_object_type_id} | {target_object_type_id} | direct | data_view |

### Mapping Rules

| Source Property | Target Property |
|------------------|-----------------|
| {source_prop} | {target_prop} |
```

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| {name} | YES | 关系类型显示名称 |
| Endpoint | YES | 关系端点定义（Source、Target、Type） |
| Source | YES | 起点对象类型 ID |
| Target | YES | 终点对象类型 ID |
| Type | YES | `direct` (直接映射) 或 `data_view` (视图映射) |
| Mapping Rules | YES | 属性映射关系 |

### 关系类型

#### 直接映射 (direct)

通过属性值匹配建立关联：

```markdown
## RelationType: Pod属于Node

Pod 实例与其所属 Node 的归属关系。

### Endpoint

| Source | Target | Type |
|--------|--------|------|
| pod | node | direct |

### Mapping Rules

| Source Property | Target Property |
|------------------|-----------------|
| pod_node_name | node_name |
```

#### 视图映射 (data_view)

通过中间视图建立关联：

```markdown
## RelationType: 用户点赞帖子

用户与帖子之间的点赞关系。

### Endpoint

| Source | Target | Type |
|--------|--------|------|
| user | post | data_view |

### Mapping View

| Type | ID |
|------|-----|
| data_view | user_post_likes_view |

### Source Mapping

| Source Property | View Property |
|-----------------|----------------|
| user_id | uid |

### Target Mapping

| View Property | Target Property |
|---------------|-----------------|
| pid | post_id |
```

---

## 行动类型定义规范

### 语法

```markdown
## ActionType: {name}

{description}

### Bound Object

| Bound Object | Action Type |
|--------------|-------------|
| {object_type_id} | add | modify | delete | query |

### Trigger Condition

```yaml
field: {property_name}
operation: == | != | > | < | >= | <= | in | not_in | exist | not_exist
value: {value}
```

### Pre-conditions

(optional) 执行前的数据前置条件，不满足则阻止行动执行

| Object | Check | Condition | Message |
|--------|-------|-----------|---------|
| {object_type_id} | relation:{relation_id} | exist | 违反时的说明 |
| {object_type_id} | property:{property_name} | {op} {value} | 违反时的说明 |

### Scope of Impact

| Object | Impact Description |
|--------|--------------------|
| {object_type_id} | {影响说明} |

### Tool Configuration

| Type | Toolbox ID | Tool ID |
|------|------------|---------|
| tool | {toolbox_id} | {tool_id} |

or

| Type | MCP ID | Tool Name |
|------|--------|-----------|
| mcp | {mcp_id} | {tool_name} |

### Parameter Binding

| Parameter | Source | Binding | Description |
|-----------|--------|---------|-------------|
| {param_name} | property | {property_name} | {说明} |
| {param_name} | input | - | {说明} |
| {param_name} | const | {value} | {说明} |

### Schedule

(optional)

| Type | Expression |
|------|------------|
| FIX_RATE | {interval} |
| CRON | {cron_expr} |

### Execution Description

(optional) Detailed execution flow...
```

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| {name} | YES | 行动类型显示名称 |
| Bound Object | YES | 目标对象类型 ID |
| Action Type | YES | `add` / `modify` / `delete` / `query` |
| Trigger Condition | NO | 自动触发的条件 |
| Pre-conditions | NO | 执行前的数据前置条件 |
| Scope of Impact | NO | 影响范围声明 |
| Tool Configuration | YES | 执行的工具或 MCP |
| Parameter Binding | YES | 参数来源配置 |
| Schedule | NO | 定时执行配置 |

### 触发条件操作符

以下操作符适用于 Trigger Condition 和 Pre-conditions：

| 操作符 | 说明 | 示例 |
|--------|------|------|
| == | 等于 | `value: Running` |
| != | 不等于 | `value: Running` |
| > | 大于 | `value: 100` |
| < | 小于 | `value: 100` |
| >= | 大于等于 | `value: 100` |
| <= | 小于等于 | `value: 100` |
| in | 包含于 | `value: [A, B, C]` |
| not_in | 不包含于 | `value: [A, B, C]` |
| exist | 存在 | (无需 value) |
| not_exist | 不存在 | (无需 value) |
| range | 范围内 | `value: [0, 100]` |

### 参数来源

| 来源 | 说明 |
|------|------|
| property | 从对象属性获取 |
| input | 运行时用户输入 |
| const | 常量值 |

---

## 风险类型定义规范

风险类型（RiskType）用于对行动类型和对象类型的执行风险进行结构化建模。风险类型是独立类型，不是行动类型的附属字段；ActionType 的 `risk_level` 声明「多危险」，RiskType 声明「如何管控」。

### 语法

```markdown
## RiskType: {name}

{description}

### Control Scope

{管控范围的描述文字}

### Control Policy

- {策略描述1}
- {策略描述2}

### Pre-checks

(optional)

| Object | Check | Condition | Message |
|--------|-------|-----------|---------|
| {object_type_id} | {check_type} | {condition} | {message} |

### Rollback Plan

1. {回滚步骤1}
2. {回滚步骤2}

### Audit Requirements

- {审计要求1}
- {审计要求2}
```

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| {name} | YES | 风险类型显示名称 |
| Control Scope | YES | 管控范围描述 |
| Control Policy | YES | 管控策略（至少一条） |
| Pre-checks | NO | 执行前检查项列表 |
| Rollback Plan | NO | 失败恢复策略 |
| Audit Requirements | NO | 审计日志与告警配置 |

---

## 概念分组定义规范

概念分组（ConceptGroup）用于将相关的对象类型组织在一起，便于理解和管理。

### 语法

```markdown
## ConceptGroup: {name}

{description}

### Object Types

| ID | Name | Description |
|----|------|-------------|
| {object_type_id} | {name} | {description} |
```

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| {name} | YES | 分组显示名称 |
| Object Types | YES | 包含的对象类型列表 |

---

## 通用语法元素

### 表格格式

使用标准 Markdown 表格：

```markdown
| 列1 | 列2 | 列3 |
|-----|-----|-----|
| 值1 | 值2 | 值3 |
```

居中对齐（用于布尔值）：

```markdown
| 列1 | 列2 |
|-----|:---:|
| 值1 | YES |
```

### YAML 代码块

用于复杂结构（如条件表达式）：

```markdown
```yaml
condition:
  operation: and
  sub_conditions:
    - field: status
      operation: ==
      value: Failed
    - field: retry_count
      operation: <
      value: 3
`` `
```

### Mermaid 图表

用于可视化关系：

```markdown
```mermaid
graph LR
    A --> B
    B --> C
`` `
```

### 引用块

用于关键信息高亮：

```markdown
> **注意**: 该对象变更需要审批流程
```

### 标题层级

标题层级在所有文件类型中保持一致：

- `#` - 网络标题（`# {Network Name}`）
- `##` - 类型定义（`## ObjectType:` / `## RelationType:` / `## ActionType:` / `## RiskType:` / `## ConceptGroup:`）或网络概览（`## Network Overview`）
- `###` - 定义内 section（Data Properties, Keys, Endpoint, Mapping Rules, Trigger Condition 等）
- `####` - 子项（例如逻辑属性名）

---

## 文件组织

### 根文件发现与目录加载

当传入**目录**作为网络入口时（如 `validate network <dir>`、`load_network(dir)`），SDK/CLI 按以下顺序自动发现根文件：

1. `network.bkn`（推荐）
2. `network.md`
3. `index.bkn`（兼容）
4. `index.md`
5. 若以上均不存在，且目录中恰好只有一个 `type: network` 文件，则使用该文件
6. 否则报错「根文件不唯一/无法确定」

**无 includes 时的默认输入**：当根文件为 `type: network` 且未声明 `includes` 时，默认将同目录下所有可解析为合法 BKN 的文件（`.bkn`、`.bknd`、`.md`）视为同一网络输入；仅扫描同目录，不递归子目录。若根文件已声明 `includes`，则完全按 `includes` 加载，不做隐式目录发现。此规则仅对 `type: network` 生效，不对 `fragment` 生效。

### 模式一：单文件（小型网络）

所有定义在一个 `.bkn` 文件中：

```markdown
---
type: network
id: my-network
name: My Network
---

# My Network

Network description...

## Network Overview

...
```

### 模式二：按类型拆分（中型网络）

使用 `network.bkn` 或 `index.bkn` 引用其他文件（推荐 `network.bkn`）：

```markdown
---
type: network
id: my-network
includes:
  - objects.bkn
  - relations.bkn
  - actions.bkn
---

# My Network

网络描述...
```

### 模式三：每定义一文件（大型网络，推荐）

每个对象/关系/行动/风险独立一个文件。支持两种编排入口：

**模式 A：SKILL.md 编排**（Agent Skill 标准）

```
{business_dir}/
├── SKILL.md                     # agentskills.io 标准入口，含网络拓扑、索引、使用指南
├── network.bkn                  # 推荐根文件（也可用 index.bkn 兼容）
├── checksum.txt                 # 可选，目录级一致性校验（SDK generate_checksum_file 生成）
├── object_types/
│   ├── material.bkn             # type: object_type
│   └── inventory.bkn            # type: object_type
├── relation_types/
│   └── material_to_inventory.bkn # type: relation_type
├── action_types/
│   ├── check_inventory.bkn      # type: action_type
│   └── adjust_inventory.bkn     # type: action_type
├── risk_types/
│   └── inventory_adjustment_risk.bkn  # type: risk_type
├── concept_groups/
│   └── supply_chain.bkn         # type: concept_group
└── data/                        # 可选，.bknd 实例数据
    └── scenario.bknd
```

### SKILL.md 与 BKN 兼容性

`SKILL.md` 是 agentskills.io 定义的 Agent Skill 入口文件，与 BKN 的目录组织互补使用：

- **SKILL.md 管职责**：描述 Skill 的能力、脚本入口、工作流、模板和输出规则，供 AI Agent 解读。
- **network.bkn / index.bkn 管结构**：通过 frontmatter `type: network` + `includes` 声明网络拓扑和文件编排。
- **互不替代**：SKILL.md 不是 BKN 根文件，SDK/CLI 的 `load_network` 和 `validate network` 读取的是 `network.bkn`（或 `index.bkn`），而非 `SKILL.md`。
- **共存推荐**：在模式 A 目录中同时放置 `SKILL.md`（Agent 入口）和 `network.bkn`（SDK/CLI 入口），两者各司其职。
- **checksum 纳入**：`SKILL.md` 被 `checksum generate` 纳入校验和计算（按 `SKILL.md` 全文 normalize 后哈希），确保 Skill 描述变更可被审计追踪。
- **目录校验兼容**：`validate network <dir>` 和 `load_network(dir)` 在 Skill 目录下正常工作——自动发现 `network.bkn`（或回退 `index.bkn`），SKILL.md 不影响网络加载。

**模式 B：network.bkn / index.bkn 编排**（推荐 `network.bkn`，`index.bkn` 兼容）

```
{business_dir}/
├── network.bkn                  # 推荐：type: network，作为网络入口（优先级高于 index.bkn）
├── index.bkn                    # 兼容：type: network，当 network.bkn 不存在时使用
├── checksum.txt                 # 可选，目录级一致性校验（SDK generate_checksum_file 生成）
├── object_types/
│   ├── pod.bkn                  # type: object_type
│   ├── node.bkn                 # type: object_type
│   └── service.bkn              # type: object_type
├── relation_types/
│   ├── pod_belongs_node.bkn     # type: relation_type
│   └── service_routes_pod.bkn   # type: relation_type
├── action_types/
│   ├── restart_pod.bkn          # type: action_type
│   └── cordon_node.bkn          # type: action_type
├── risk_types/
│   └── pod_restart_risk.bkn     # type: risk_type
├── concept_groups/
│   └── k8s.bkn                  # type: concept_group
└── data/                        # 可选，.bknd 实例数据
```

目录名（`object_types/`、`relation_types/`、`action_types/`、`risk_types/`、`concept_groups/`、`data/`）为约定，文件的 `type` 字段为定义类型的权威声明。

**对象类型文件示例** (`pod.bkn`):

```markdown
---
type: object_type
id: pod
name: Pod实例
tags: [k8s, 计算]
---

## ObjectType: Pod实例

Kubernetes 中最小的可部署单元。

### Data Properties

| Name | Display Name | Type | Description |
|------|--------------|------|-------------|
| pod_id | Pod ID | VARCHAR | Pod 唯一标识 |
| pod_name | Pod名称 | VARCHAR | Pod 名称 |
| namespace | 命名空间 | VARCHAR | 所属命名空间 |
| status | 状态 | VARCHAR | Pod 运行状态 |

### Keys

Primary Keys: pod_id
Display Key: pod_name

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | pod_view | Pod数据视图 |
```

---

## 增量导入规范

BKN 支持将任何 `.bkn` 文件动态导入到已有的知识网络。

### 导入器能力要求（工程可控性 9+ 的前提）

建议实现一个 **BKN Importer**，将 BKN 文件转换为系统变更，并提供以下能力（缺一不可）：

| 能力 | 说明 | 目的 |
|------|------|------|
| `validate` | 结构/表格/YAML block 校验，引用完整性校验，参数绑定校验 | 阻止错误进入系统 |
| `diff` | 计算变更集（新增/更新/删除）与影响范围 | 让变更可解释、可审计 |
| `dry_run` | 在不落地的情况下执行 validate + diff | 上线前预演 |
| `apply` | 执行落地（按确定性语义与冲突策略） | 可控执行 |
| `export` | 将线上知识网络状态导出为 BKN（可 round-trip） | 防漂移、可回滚、可复现 |

> 要求：所有导入操作必须记录审计信息（操作者、时间、输入文件指纹、变更集、结果）。

### 导入的确定性（必须保证）

为保证多人协作与可回放性，导入语义必须是**确定性的（deterministic）**：

- 对同一组输入文件（不考虑文件系统顺序）导入结果一致
- 同一文件重复导入结果一致（幂等）
- 冲突可解释：要么明确失败（fail-fast），要么有明确规则（例如 last-wins），不得“隐式合并”

### 唯一键与作用域

每个定义的唯一键建议为：

- `key = (network_id, type, id)`

其中 `network_id` 由导入目标网络（导入命令参数或 `type: network` 的 `id`）确定。

### 更新语义（replace vs merge）

默认建议使用 **replace（整段覆盖）**：

- 当 `key` 已存在时，用导入文件中的定义整体替换旧定义
- **缺失字段不代表删除**：仅代表“该字段不在本次定义中”；删除元素应通过 SDK/CLI 的显式删除 API 执行，而非 BKN 文件

如确有需要，可支持受控的 **merge-by-section（按章节合并）**，但必须满足：

- 仅允许合并少数“附加型章节”（例如 `属性覆盖`、`逻辑属性`）
- 冲突必须可控：同名逻辑属性/同名字段配置冲突时 fail-fast 或 last-wins（需配置）
- 合并策略必须在导入器中显式配置并记录到导入审计日志

### 冲突与优先级

当同一个 `key` 在一次导入批次中被多个文件重复声明：

- 默认：**fail-fast**（推荐，保证稳定性）
- 可选：按显式优先级排序（例如命令行顺序或 `priority` 字段），否则不建议支持

### 导入行为

| 场景 | 行为 |
|------|------|
| ID 不存在 | 创建新定义 |
| ID 已存在 | 更新定义（覆盖） |
| 删除元素 | 通过 SDK/CLI 的 delete API 显式执行，不通过 BKN 文件 |

### 导入示例

**场景：向已有网络添加新对象类型**

创建 `deployment.bkn`:

```markdown
---
type: object_type
id: deployment
name: Deployment
tags: [k8s]
---

## ObjectType: Deployment

Kubernetes deployment controller.

### Data Properties

| Name | Display Name | Type | Description |
|------|--------------|------|-------------|
| id | ID | VARCHAR | 唯一标识 |
| deployment_name | Deployment名称 | VARCHAR | Deployment 名称 |

### Keys

Primary Keys: id
Display Key: deployment_name

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | deployment_view | Deployment视图 |
```

导入后，网络将包含新的 `deployment` 对象类型。

**场景：更新已有对象类型**

创建同 ID 的文件，导入后自动覆盖：

```markdown
---
type: object_type
id: pod
name: Pod实例（更新版）
tags: [k8s]
---

## ObjectType: Pod实例（更新版）

更新后的定义...
```

---

## 无 patch 更新模型

BKN 采用**无 patch 的更新模型**：定义文件仅用于新增与修改，删除通过 SDK/CLI API 显式执行。

### 定义文件导入（add/modify）

- 单个 `.bkn` 文件导入时，按 `(network, type, id)` 执行 **upsert**（新增或覆盖）
- 修改：直接编辑对应定义文件，重新导入即可覆盖
- 缺失字段不代表删除：仅表示该字段不在本次定义中

### 删除元素

- 删除应通过 **SDK/CLI 的 delete API** 显式执行，不通过 BKN 文件
- 删除操作要求：显式参数、可审计、支持 dry-run 与批量删除

### 编辑流程

1. **新增**：创建 `.bkn` 文件，导入
2. **修改**：编辑 `.bkn` 文件，重新导入
3. **删除**：调用 SDK/CLI 的 delete 接口

---

## 最佳实践

### 命名规范

- **ID**: 小写字母、数字、下划线，如 `pod_belongs_node`
- **显示名称**: 简洁明确，如 "Pod属于节点"
- **标签**: 使用统一的标签体系

### 文档结构

1. 网络描述放在文件开头
2. 使用 mermaid 图展示整体拓扑
3. 对象定义在前，关系和行动在后
4. 相关定义放在一起

### 简洁原则

- 优先使用完全映射模式
- 只在需要时声明属性覆盖
- 避免重复信息

### 可读性

- 使用表格呈现结构化数据
- 添加业务语义说明
- 必要时使用 mermaid 图

---

## 参考

- [架构设计](./ARCHITECTURE.md)
- 样例：
  - [K8s 网络](../examples/k8s-network/) - K8s 拓扑知识网络
  - [供应链网络](../examples/supplychain-hd/) - 供应链业务知识网络
