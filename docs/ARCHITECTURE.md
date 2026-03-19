# BKN 架构设计

BKN (Business Knowledge Network) 是一种 Markdown-based 的业务知识网络建模语言，用于描述业务知识网络中的对象、关系和行动。

## 设计理念

- **人类可读**: 使用 Markdown 语法，业务人员也能理解和编辑
- **Agent 友好**: 结构化的 YAML frontmatter + 语义化的 section，便于 LLM 解析和生成
- **增量导入**: 任何 `.bkn` 文件可直接导入到已有的知识网络，支持动态更新
- **大规模友好**: 每个定义独立一个文件，100 个对象 = 100 个小文件

## 架构概览

```mermaid
flowchart TB
    subgraph BKN_Language["BKN 业务知识网络建模语言"]
        direction TB

        subgraph Types["六种类型"]
            Object["ObjectType<br/>对象类"]
            Relation["RelationType<br/>关系类"]
            Action["ActionType<br/>行动类"]
            Risk["RiskType<br/>风险类"]
            ConceptGroup["ConceptGroup<br/>概念分组"]
            Network["Network<br/>网络"]
        end

        subgraph Structure["文件结构"]
            Frontmatter["YAML Frontmatter<br/>元数据"]
            Body["Markdown Body<br/>描述 + 定义"]
        end

        Object --> Frontmatter
        Relation --> Frontmatter
        Action --> Frontmatter
        Risk --> Frontmatter
        ConceptGroup --> Frontmatter
        Frontmatter --> Body
    end

    subgraph DataLayer["数据层"]
        DataView["Data View<br/>数据视图"]
        MetricModel["Metric Model<br/>指标模型"]
        Operator["Operator<br/>算子"]
    end

    subgraph ActionLayer["行动层"]
        Tool["Tool<br/>工具"]
        MCP["MCP<br/>模型上下文协议"]
    end

    Object -.->|数据来源| DataView
    Object -.->|逻辑属性| MetricModel
    Object -.->|逻辑属性| Operator
    Relation -.->|映射| Object
    Action -.->|绑定| Object
    Action -.->|执行| Tool
    Action -.->|执行| MCP
    ConceptGroup -.->|组织| Object
```

## 工作流

```mermaid
flowchart LR
    subgraph Editors["编辑者"]
        Human["人类<br/>Markdown编辑"]
        Agent["Agent<br/>LLM生成"]
    end

    subgraph Files["BKN文件"]
        BKNFile[".bkn 文件"]
    end

    subgraph Backend["后端服务"]
        Parser["BKN Parser<br/>解析器"]
        API["Knowledge Network API<br/>知识网络管理"]
        KN["Knowledge Network<br/>知识网络"]
    end

    Human -->|编写/修改| BKNFile
    Agent -->|读取/生成| BKNFile
    BKNFile -->|解析| Parser
    Parser -->|调用| API
    API -->|持久化| KN
```

## 六种类型

### 对象类 (ObjectType)

描述业务对象，如 Pod、Node、Service 等。

**核心特性**:
- 声明数据属性（Data Properties）和键定义（Keys）
- 可选映射数据视图（Data Source）
- 支持逻辑属性（指标、算子等）

```mermaid
classDiagram
    class BknObject {
        +String id
        +String name
        +List~String~ tags
        +DataSource data_source
        +List~String~ primary_keys
        +String display_key
        +List~Property~ data_properties
        +List~LogicProperty~ logic_properties
    }

    class DataSource {
        +String type
        +String id
        +String name
    }

    class LogicProperty {
        +String name
        +String display
        +String type
        +String source
        +List~Parameter~ parameters
    }

    BknObject --> DataSource
    BknObject --> LogicProperty
```

### 关系类 (RelationType)

描述两个对象之间的关联关系。

**核心特性**:
- 定义起点和终点对象（Endpoint）
- 支持直接映射（direct）和视图映射（data_view）
- 声明属性映射规则

```mermaid
classDiagram
    class Relation {
        +String id
        +String name
        +ObjectRef source
        +ObjectRef target
        +String type
        +List~MappingRule~ mapping_rules
    }

    class MappingRule {
        +String source_property
        +String target_property
    }

    Relation --> MappingRule
```

### 行动类 (ActionType)

描述可执行的操作，绑定工具或 MCP。

**核心特性**:
- 绑定目标对象（Bound Object）
- 定义触发条件（Trigger Condition）
- 配置执行工具和参数（Tool Configuration / Parameter Binding）
- 支持调度配置（Schedule）
- 可声明前置条件（Pre-conditions）和影响范围（Scope of Impact）

```mermaid
classDiagram
    class Action {
        +String id
        +String name
        +Boolean enabled
        +String risk_level
        +Boolean requires_approval
        +String action_type
        +String bound_object
        +String affect_object
        +Condition condition
        +List~PreCondition~ pre_conditions
        +ActionSource action_source
        +List~Parameter~ parameters
        +Schedule schedule
        +List~Impact~ scope_of_impact
    }

    class Condition {
        +String object_type_id
        +String field
        +String operation
        +Any value
    }

    class ToolSource {
        +String type
        +String toolbox_id
        +String tool_id
    }

    class MCPSource {
        +String type
        +String mcp_id
        +String tool_name
    }

    Action --> Condition
    Action --> ToolSource
    Action --> MCPSource
```

### 风险类 (RiskType)

对行动类和对象类的执行风险进行结构化建模。风险类型是独立类型，不是行动类型的附属字段。

**核心特性**:
- 管控范围（Control Scope）
- 管控策略（Control Policy）
- 前置检查（Pre-checks）
- 回滚方案（Rollback Plan）
- 审计要求（Audit Requirements）

### 概念分组 (ConceptGroup)

将相关的对象类型组织在一起，便于理解和管理。

**核心特性**:
- 包含的对象类型列表（Object Types）
- 提供业务领域的逻辑分组

## 文件组织

每个对象/关系/行动/风险/概念分组独立一个文件，以 `network.bkn` 为根文件入口。

```
{business_dir}/
├── SKILL.md                     # agentskills.io 标准入口
├── network.bkn                  # 网络根文件（type: network）
├── CHECKSUM                     # 可选，目录级一致性校验
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
│   └── restart_pod_high_risk.bkn # type: risk_type
├── concept_groups/
│   └── k8s.bkn                  # type: concept_group
└── data/                        # 可选，.csv 实例数据
    └── scenario.csv
```

**优势**：
- 每个文件 50-100 行，人类易读
- LLM 处理时无 token 限制问题
- 便于团队协作和版本控制
- 支持按需加载

## 增量导入机制

BKN 的核心特性是支持**动态增量导入**：任何 `.bkn` 文件可直接导入到已有的知识网络。

```mermaid
flowchart LR
    subgraph Existing["已有知识网络"]
        KN["Knowledge Network<br/>Pod, Node, Service"]
    end

    subgraph NewFiles["新增 BKN 文件"]
        NewObject["deployment.bkn<br/>type: object_type"]
        NewRelation["pod_in_deployment.bkn<br/>type: relation_type"]
    end

    NewObject -->|导入| KN
    NewRelation -->|导入| KN

    subgraph Result["导入后"]
        KN2["Knowledge Network<br/>Pod, Node, Service,<br/>Deployment + 关系"]
    end

    KN --> KN2
```

### 导入行为

| 场景 | 行为 |
|------|------|
| ID 不存在 | 新增定义 |
| ID 已存在 | 更新定义（覆盖） |
| 删除定义 | 通过 SDK/CLI 的 delete API 显式执行，不通过 BKN 文件 |

### 支持的文件类型

| type | 说明 | 用途 |
|------|------|------|
| `network` | 完整知识网络 | 初始化或全量导入 |
| `object_type` | 单个对象类型定义 | 增量添加/更新对象 |
| `relation_type` | 单个关系类型定义 | 增量添加/更新关系 |
| `action_type` | 单个行动类型定义 | 增量添加/更新行动 |
| `risk_type` | 单个风险类型定义 | 增量添加/更新风险 |
| `concept_group` | 概念分组 | 增量添加/更新分组 |

### 典型工作流

1. **初始化**: 导入 `network` 类型的完整定义
2. **扩展**: 导入单个 `object_type` / `relation_type` / `action_type` / `risk_type` 文件
3. **修改**: 导入同 ID 的文件，自动覆盖
4. **删除**: 调用 SDK/CLI 的 delete API

## 与 知识网络管理 API 的映射

> 说明：接口路径仅用于表达 BKN 概念与系统 API 的对应关系，具体实现路径以实际部署为准。

| BKN 概念 | API 端点 |
|----------|----------|
| ObjectType | `/api/knowledge-networks/{kn_id}/object-types` |
| RelationType | `/api/knowledge-networks/{kn_id}/relation-types` |
| ActionType | `/api/knowledge-networks/{kn_id}/action-types` |

## 参考

- [BKN 语言规范](./SPECIFICATION.md)
- [BKN vs RESTful API 对比](./BKN_vs_REST_API.md)
- 样例：
  - [K8s 网络](../examples/k8s-network/) - K8s 拓扑知识网络
  - [供应链网络](../examples/supplychain-hd/) - 供应链业务知识网络
