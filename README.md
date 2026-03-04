# BKN Specification

**BKN (Business Knowledge Network)** is a Markdown-based domain modeling language for business knowledge networks. This repository hosts the official specification and examples.

[中文](README.zh.md)

## Specification

The core documentation is the **BKN Language Specification**:

- **[SPECIFICATION.md](docs/SPECIFICATION.md)** — Full specification (Chinese)
- **[SPECIFICATION.en.md](docs/SPECIFICATION.en.md)** — English edition

### Key Concepts

| Concept | Description |
|---------|-------------|
| Object | Business object types (e.g. Pod, Node, Service) |
| Relation | Links between objects |
| Action | Operations on objects (with tool/MCP binding) |
| data_view | Data source mapping for objects and relations |

### File Organization

```
├── docs/
│   ├── SPECIFICATION.md      # Full spec (CN)
│   ├── SPECIFICATION.en.md   # Full spec (EN)
│   ├── ARCHITECTURE.md       # Architecture overview
│   └── templates/            # BKN file templates
└── examples/                 # Example networks (Kubernetes topology)
    ├── k8s-topology.bkn      # Single-file example
    ├── k8s-network/          # Split by type (objects, relations, actions)
    └── k8s-modular/          # One definition per file
```

## Demo Tool

This repo includes **BKN Editor**, a demo web app for editing and visualizing BKN files:

- File tree and Monaco editor for `.bkn` files
- Graph view of objects and relations (React Flow)
- Templates for Object, Relation, Action

```bash
cd bkn_editor
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). The demo loads examples from `examples` and stores data in browser localStorage.

### AI Generation

The BKN Editor supports AI-assisted generation via OpenAI or Anthropic. Configure by copying the example and editing `.env.local`:

```bash
cd bkn_editor
copy .env.local.example .env.local   # Windows
# or: cp .env.local.example .env.local
```

| Variable | Description | Required |
|----------|-------------|----------|
| `AI_PROVIDER` | `openai` or `anthropic` (default: `openai`) | Optional |
| `OPENAI_API_KEY` | [OpenAI API key](https://platform.openai.com/api-keys) | When using OpenAI |
| `OPENAI_MODEL` | Model name (default: `gpt-4o-mini`) | Optional |
| `OPENAI_BASE_URL` | Custom OpenAI-compatible API base URL | Optional |
| `ANTHROPIC_API_KEY` | [Anthropic API key](https://console.anthropic.com/) | When using Anthropic |
| `ANTHROPIC_MODEL` | Model name | Optional |
| `ANTHROPIC_BASE_URL` | Custom Anthropic API base URL | Optional |

Restart `npm run dev` after changing `.env.local`.

> **Note:** BKN Editor is a **demo** for exploring the specification. Production tooling should follow the spec independently.

## License

Apache-2.0
