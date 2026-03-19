# BKN CLI

`bkn` is a Go-based command-line interface for inspecting, validating, and transforming BKN files.

This README assumes your current working directory is `cli/`.

## Run

Show top-level help:

```bash
go run ./cmd/bkn --help
```

All commands support:

```bash
--format text
--format json
```

The default output format is `text`.

## Commands

### Inspect

Inspect a single file:

```bash
go run ./cmd/bkn inspect file ../examples/risk/actions.bkn
```

Inspect a single file with metadata and definition details:

```bash
go run ./cmd/bkn inspect file ../examples/risk/actions.bkn --verbose
```

Inspect a network with includes resolved (path can be a file or directory):

```bash
go run ./cmd/bkn inspect network ../examples/risk/index.bkn
go run ./cmd/bkn inspect network ../examples/k8s-network
```

Inspect a network with metadata, description, includes, and definition lists:

```bash
go run ./cmd/bkn inspect network ../examples/risk/index.bkn --verbose
```

JSON output:

```bash
go run ./cmd/bkn inspect network ../examples/risk/index.bkn --verbose --format json
```

### Validate

Validate all data tables in a network (path can be a file or directory):

```bash
go run ./cmd/bkn validate network ../examples/risk/index.bkn
go run ./cmd/bkn validate network ../examples/k8s-network
```

Validate a single data file against a network schema (`--network` can be file or directory):

```bash
go run ./cmd/bkn validate table ../examples/risk/data/risk_scenario.bknd --network ../examples/risk
```

JSON output:

```bash
go run ./cmd/bkn validate network ../examples/risk/index.bkn --format json
```

Validation failures return a non-zero exit code. With `--format json`, the command still emits clean JSON output.

### Data

Serialize a JSON array to `.bknd` for an object:

```bash
go run ./cmd/bkn data to-bknd --object risk_scenario --network recoverable-network --in rows.json
```

Serialize a JSON array to `.bknd` for a relation:

```bash
go run ./cmd/bkn data to-bknd --relation pod_on_node --network k8s --in rows.json
```

Write the generated result to a file:

```bash
go run ./cmd/bkn data to-bknd --object pod --network demo --in rows.json --out pod.bknd
```

Control column order:

```bash
go run ./cmd/bkn data to-bknd --object pod --network demo --in rows.json --columns id,name,status
```

Example input file:

```json
[
  { "scenario_id": "sec_t_01", "name": "Month-end freeze" },
  { "scenario_id": "sec_c_02", "name": "Change window" }
]
```

### Risk

Evaluate an action with external rule data:

```bash
go run ./cmd/bkn risk eval --network ../examples/risk/index.bkn --action restart_erp --context scenario_id=sec_t_01 --rules rules.json
```

JSON output:

```bash
go run ./cmd/bkn risk eval --network ../examples/risk/index.bkn --action restart_erp --context scenario_id=sec_t_01 --rules rules.json --format json
```

Example rules file:

```json
[
  {
    "scenario_id": "sec_t_01",
    "action_id": "restart_erp",
    "allowed": false,
    "risk_level": 5,
    "reason": "Month-end freeze"
  }
]
```

### Delete

Plan a delete operation:

```bash
go run ./cmd/bkn delete plan --network ../examples/k8s-modular/index.bkn --target object:pod
```

Plan multiple targets:

```bash
go run ./cmd/bkn delete plan --network ../examples/k8s-modular/index.bkn --target object:pod --target action:restart_pod
```

Simulate deletion in memory:

```bash
go run ./cmd/bkn delete simulate --network ../examples/k8s-modular/index.bkn --target object:pod
```

JSON output:

```bash
go run ./cmd/bkn delete simulate --network ../examples/k8s-modular/index.bkn --target object:pod --format json
```

Targets use `type:id` syntax. Supported types are:

- `object`
- `relation`
- `action`

### Checksum

Generate `CHECKSUM` for a business directory. The command validates BKN inputs first and refuses to write `CHECKSUM` if the directory is invalid:

```bash
go run ./cmd/bkn checksum generate ../examples/risk
```

Verify `CHECKSUM`:

```bash
go run ./cmd/bkn checksum verify ../examples/risk
```

JSON output:

```bash
go run ./cmd/bkn checksum verify ../examples/risk --format json
```

Typical test flow:

1. Fix any validation errors until `bkn validate network` passes.
2. Generate the checksum file.
3. Verify it immediately.
4. Modify a `.bkn` or `.bknd` file.
5. Verify again to confirm mismatch detection.

## Help

Each command has built-in help:

```bash
go run ./cmd/bkn inspect --help
go run ./cmd/bkn inspect network --help
go run ./cmd/bkn data to-bknd --help
go run ./cmd/bkn risk eval --help
go run ./cmd/bkn delete simulate --help
```

## SKILL.md Compatibility

The CLI works seamlessly with [agentskills.io](https://agentskills.io) SKILL directories:

- `bkn validate network <dir>` and `bkn inspect network <dir>` auto-discover the BKN root (`network.bkn` > `index.bkn`) in a Skill directory; `SKILL.md` does not interfere.
- `bkn checksum generate <dir>` includes `SKILL.md` in its checksum computation, so Skill description changes are tracked alongside BKN schema changes.
- All `--network` flags accept both file paths and directories.

## Recommended First Checks

If you want to smoke test the CLI quickly from `cli/`, run:

```bash
go run ./cmd/bkn inspect network ../examples/risk/index.bkn --verbose
go run ./cmd/bkn validate network ../examples/risk/index.bkn
go run ./cmd/bkn validate network ../examples/k8s-network
go run ./cmd/bkn checksum generate ../examples/risk
go run ./cmd/bkn checksum verify ../examples/risk
```
