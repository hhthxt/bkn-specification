# BKN Golang SDK

Planned Go SDK for parsing, validating, and transforming BKN files.

## Status

Not yet implemented. The Go SDK will provide feature parity with the [Python SDK](../python/README.md):

- Parse `.bkn` files (YAML frontmatter + Markdown body)
- Structured models for Entity, Relation, Action
- Network loading with `includes` resolution
- Transformers (e.g., kweaver JSON output)

## Planned Structure

```
golang/
├── go.mod
├── bkn/
│   ├── models.go
│   ├── parser.go
│   ├── loader.go
│   └── transformers/
│       └── kweaver.go
└── bkn_test/
    └── parser_test.go
```

Contributions welcome.
