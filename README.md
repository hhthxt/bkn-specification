# BKN Specification

**BKN (Business Knowledge Network)** is a Markdown-based domain modeling language for business knowledge networks. This repository hosts the official specification and examples.

## Specification

The core documentation is the **BKN Language Specification**:

- **[SPECIFICATION.md](docs/ontology/bkn_docs/SPECIFICATION.md)** — Full specification (Chinese)
- **[SPECIFICATION.en.md](docs/ontology/bkn_docs/SPECIFICATION.en.md)** — English edition

### Key Concepts

| Concept | Description |
|---------|-------------|
| Entity | Business object types (e.g. Pod, Node, Service) |
| Relation | Links between entities |
| Action | Operations on entities (with tool/MCP binding) |
| data_view | Data source mapping for entities and relations |

### File Organization

```
docs/ontology/bkn_docs/
├── SPECIFICATION.md      # Full spec (CN)
├── SPECIFICATION.en.md   # Full spec (EN)
├── ARCHITECTURE.md       # Architecture overview
├── examples/             # Example networks (Kubernetes topology)
│   ├── k8s-topology.bkn  # Single-file example
│   ├── k8s-network/      # Split by type (entities, relations, actions)
│   └── k8s-modular/     # One definition per file
└── templates/           # BKN file templates
```

## Demo Tool

This repo includes **BKN Editor**, a demo web app for editing and visualizing BKN files:

- File tree and Monaco editor for `.bkn` files
- Graph view of entities and relations (React Flow)
- Templates for Entity, Relation, Action

```bash
cd bkn_editor
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). The demo loads examples from `docs/ontology/bkn_docs/examples` and stores data in browser localStorage.

> **Note:** BKN Editor is a **demo** for exploring the specification. Production tooling should follow the spec independently.

## License

Apache-2.0
