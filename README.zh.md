# BKN 规范

**BKN（Business Knowledge Network，业务知识网络）** 是一种基于 Markdown 的业务知识网络建模语言。本仓库托管官方规范与示例。

[English](README.md)

## 规范文档

核心文档为 **BKN 语言规范**：

- **[SPECIFICATION.md](docs/SPECIFICATION.md)** — 完整规范（中文）
- **[SPECIFICATION.en.md](docs/SPECIFICATION.en.md)** — 英文版

## CLI

仓库同时提供了一个基于 Go 的 CLI，用于查看、校验和转换 BKN 文件：

- **[cli/README.md](cli/README.md)** — 使用方法、命令示例与 smoke test 命令

### 核心概念

| 概念 | 说明 |
|------|------|
| Object（对象） | 业务对象类型（如 Pod、Node、Service） |
| Relation（关系） | 对象之间的连接 |
| Action（行动） | 对对象的操作定义（可绑定 tool/MCP） |
| Risk（风险） | 风险类，对执行风险进行结构化建模 |
| data_view | 对象/关系映射的数据来源 |

### 更新网络（无 patch 模型）

- **新增/修改**：编辑 `.bkn` 文件并导入（按 `network`、`type`、`id` 执行 upsert）。
- **删除**：通过 SDK/CLI 的 delete API 显式执行；删除不通过 BKN 文件表达。

### 文件结构

```
├── docs/
│   ├── SPECIFICATION.md      # 完整规范（中文）
│   ├── SPECIFICATION.en.md   # 完整规范（英文）
│   ├── ARCHITECTURE.md       # 架构概览
│   └── templates/            # BKN 文件模板
└── examples/                 # 示例网络（Kubernetes 拓扑）
    ├── k8s-topology.bkn      # 单文件示例
    ├── k8s-network/          # 按类型拆分（objects/relations/actions）
    └── k8s-modular/          # 每定义一文件
```

## 许可证

Apache-2.0
