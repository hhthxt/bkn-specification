# BKN 规范

**BKN（Business Knowledge Network，业务知识网络）** 是一种基于 Markdown 的业务知识网络建模语言。本仓库托管官方规范与示例。

[English](README.md)

## 规范文档

核心文档为 **BKN 语言规范**：

- **[SPECIFICATION.md](docs/bkn_docs/SPECIFICATION.md)** — 完整规范（中文）
- **[SPECIFICATION.en.md](docs/bkn_docs/SPECIFICATION.en.md)** — 英文版

### 核心概念

| 概念 | 说明 |
|------|------|
| Entity（实体） | 业务对象类型（如 Pod、Node、Service） |
| Relation（关系） | 实体之间的连接 |
| Action（行动） | 对实体的操作定义（可绑定 tool/MCP） |
| data_view | 实体/关系映射的数据来源 |

### 文件结构

```
docs/bkn_docs/
├── SPECIFICATION.md      # 完整规范（中文）
├── SPECIFICATION.en.md    # 完整规范（英文）
├── ARCHITECTURE.md        # 架构概览
├── examples/              # 示例网络（Kubernetes 拓扑）
│   ├── k8s-topology.bkn  # 单文件示例
│   ├── k8s-network/      # 按类型拆分（entities/relations/actions）
│   └── k8s-modular/      # 每定义一文件
└── templates/            # BKN 文件模板
```

## 演示工具

本仓库内含 **BKN Editor**，用于编辑和可视化 BKN 文件的演示 Web 应用：

- 文件树与 Monaco 编辑器
- 实体-关系网络图（React Flow）
- Entity / Relation / Action 模板

```bash
cd bkn_editor
npm install
npm run dev
```

访问 [http://localhost:3000](http://localhost:3000)。演示会加载 `docs/bkn_docs/examples` 下的示例，数据保存在浏览器 localStorage。

### AI 生成

BKN Editor 支持通过 OpenAI 或 Anthropic 的 AI 辅助生成。配置方法：复制示例文件并编辑 `bkn_editor/.env.local`：

```bash
cd bkn_editor
copy .env.local.example .env.local   # Windows
# 或: cp .env.local.example .env.local
```

| 变量 | 说明 | 必填 |
|------|------|------|
| `AI_PROVIDER` | `openai` 或 `anthropic`（默认 `openai`） | 可选 |
| `OPENAI_API_KEY` | [OpenAI API 密钥](https://platform.openai.com/api-keys) | 使用 OpenAI 时 |
| `OPENAI_MODEL` | 模型名（默认 `gpt-4o-mini`） | 可选 |
| `OPENAI_BASE_URL` | 自定义 OpenAI 兼容 API 地址 | 可选 |
| `ANTHROPIC_API_KEY` | [Anthropic API 密钥](https://console.anthropic.com/) | 使用 Anthropic 时 |
| `ANTHROPIC_MODEL` | 模型名 | 可选 |
| `ANTHROPIC_BASE_URL` | 自定义 Anthropic API 地址 | 可选 |

修改 `.env.local` 后需重启 `npm run dev`。

> **说明**：BKN Editor 为**演示工具**，用于理解规范。生产工具需按规范独立实现。

## 许可证

Apache-2.0
