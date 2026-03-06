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
| Risk | Risk type for structured execution risk modeling |
| data_view | Data source mapping for objects and relations |

### Updating Networks (No-Patch Model)

- **Add/modify**: Edit `.bkn` files and import (upsert by `network`, `type`, `id`).
- **Delete**: Use the SDK/CLI delete API explicitly; deletion is not expressed in BKN files.

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

## License

Apache-2.0
