# BKN Language Specification

Version: 1.0.0
spec_version: 1.0.0

## Overview

BKN (Business Knowledge Network) is a Markdown-based modeling language for business knowledge networks, used to describe such networks. This document defines the syntax specification for BKN.

### Reader Roadmap (Where to Start)

- **Business readers**: Start with the examples and tables in "Entity Definition", "Relation Definition", and "Action Definition", then read "Best Practices".
- **Engineering readers**: Start with "Incremental Import Specification (Deterministic Semantics)" and "Validation / Failure Strategy", then review the field types.
- **Agent / LLM**: Prefer reading by "one definition per file" to avoid loading too much content at once.

### Glossary

| Term | Meaning |
|------|---------|
| BKN | Business Knowledge Network |
| knowledge_network / network | The overall collection of a business knowledge network |
| entity | Business object type (e.g., Pod/Node/Service) |
| relation | Relationship type connecting two entities (e.g., belongs_to, routes_to) |
| action | Operation definition executed on an entity (may bind to tool/mcp) |
| data_view | Data view (data source that entities/relations can map directly to) |
| primary_key | Primary key field (for uniquely identifying instances) |
| display_key | Display field (for UI / search display) |
| fragment | Mixed fragment file (may contain multiple entities/relations/actions) |
| delete | Delete marker file (explicitly declares definitions to be removed) |

### Primitives (Canonical Section and Table Terms)

The table below is organized by **heading level**. The Level column shows the canonical depth in network/fragment files.

| Level | English (canonical) | Definition | Syntax |
|:-----:|---------------------|------------|--------|
| `#` | Entities | Section: all entity definitions in this file | `# Entities` |
| `#` | Relations | Section: all relation definitions | `# Relations` |
| `#` | Actions | Section: all action definitions | `# Actions` |
| `##` | Entity | Individual entity definition | `## Entity: {id}` |
| `##` | Relation | Individual relation definition | `## Relation: {id}` |
| `##` | Action | Individual action definition | `## Action: {id}` |
| `###` | Data Source | The data view this entity maps from | `### Data Source` |
| `###` | Data Properties | Explicit list of fields (name, type, PK, index) | `### Data Properties` |
| `###` | Property Override | Per-property overrides (e.g. index config) | `### Property Override` |
| `###` | Logic Properties | Derived fields: metric, operator | `### Logic Properties` |
| `###` | Business Semantics | Human-readable meaning of the entity/relation | `### Business Semantics` |
| `###` | Endpoints | Relation endpoints: source, target, type | `### Endpoints` |
| `###` | Mapping Rules | How source/target properties map | `### Mapping Rules` |
| `###` | Mapping View | For data_view relations: the join view | `### Mapping View` |
| `###` | Source Mapping | Map source entity props to view | `### Source Mapping` |
| `###` | Target Mapping | Map view to target entity props | `### Target Mapping` |
| `###` | Bound Entity | Entity this action operates on | `### Bound Entity` |
| `###` | Trigger Condition | When to run (YAML condition) | `### Trigger Condition` |
| `###` | Tool Configuration | tool or MCP binding | `### Tool Configuration` |
| `###` | Parameter Binding | param name, source, binding | `### Parameter Binding` |
| `###` | Schedule | FIX_RATE or CRON | `### Schedule` |
| `###` | Scope of Impact | What objects are affected | `### Scope of Impact` |
| `####` | {property_name} | Individual logic property sub-section | `#### {name}` |
| — | Primary Key | Field that uniquely identifies an instance | blockquote `**Primary Key**`, table column |
| — | Display Key | Field used for UI label / search display | blockquote `**Display Key**`, table column |
| — | Action Type | add \| modify \| delete | table column |

> In single-file format (`type: entity/relation/action`), all levels shift up by one: `##` becomes `#`, `###` becomes `##`, `####` becomes `###`. The `#` group headings (Entities/Relations/Actions) are not used.

Table column names (canonical): Type, ID, Name, Property, Display Name, Primary Key, Index, Index Config, Description; Source, Target; Source Property, Target Property; Parameter, Source, Binding; Bound Entity, Action Type; Object, Impact Description.

## File Format

### File Extension

- `.bkn` — BKN file

### File Encoding

- UTF-8

### Basic Structure

Every BKN file consists of two parts:

1. **YAML Frontmatter**: File metadata
2. **Markdown Body**: Knowledge network definition content

```markdown
---
type: network
id: example-network
name: Example Network
version: 1.0.0
---

# Network Title

Network description...

## Entity: entity_id

Entity definition...

## Relation: relation_id

Relation definition...

## Action: action_id

Action definition...
```

---

## Frontmatter Specification

### Engineering Control Fields (Recommended)

To support scalable collaboration, approval, and audit, use the following fields in definition files:

| Field | Applicable type | Description |
|-------|-----------------|-------------|
| `spec_version` | all | Specification version used by this file (inherits document spec_version by default) |
| `namespace` | entity/relation/action/fragment/delete | Namespace/package name for large-scale organization and conflict avoidance (e.g., `platform.k8s`) |
| `owner` | entity/relation/action/fragment/delete | Owner/team (for audit and approval routing) |
| `enabled` | action | Whether enabled (default `false` recommended; import does not imply enablement) |
| `risk_level` | action | Risk level (`low|medium|high` for approval and release strategy) |
| `requires_approval` | action | Whether approval is required to enable/execute |

### File Types (type)

| type | Description | Purpose |
|------|-------------|---------|
| `network` | Full knowledge network | Network file containing multiple definitions |
| `entity` | Single entity definition | Standalone entity file, directly importable |
| `relation` | Single relation definition | Standalone relation file, directly importable |
| `action` | Single action definition | Standalone action file, directly importable |
| `fragment` | Mixed fragment | Contains multiple types of partial definitions |
| `delete` | Delete marker | Marks definitions to be deleted |

### Network File (type: network)

```yaml
---
type: network                    # Full knowledge network
id: string                       # Network ID, unique identifier
name: string                     # Network display name
version: string                  # Version (semver)
tags: [string]                   # Optional, tag list
description: string              # Optional, network description
includes: [string]               # Optional, referenced files
---
```

### Single Entity File (type: entity)

```yaml
---
type: entity                     # Single entity definition
id: string                       # Entity ID, unique identifier
name: string                     # Entity display name
version: string                  # Optional, version
network: string                  # Network ID (recommended required for import determinism)
namespace: string                # Optional, namespace/package
owner: string                    # Optional, owner/team
tags: [string]                   # Optional, tag list
---
```

### Single Relation File (type: relation)

```yaml
---
type: relation                   # Single relation definition
id: string                       # Relation ID, unique identifier
name: string                     # Relation display name
version: string                  # Optional, version
network: string                  # Network ID (recommended required for import determinism)
namespace: string                # Optional, namespace/package
owner: string                    # Optional, owner/team
---
```

### Single Action File (type: action)

```yaml
---
type: action                     # Single action definition
id: string                       # Action ID, unique identifier
name: string                     # Action display name
action_type: add | modify | delete  # Action type
version: string                  # Optional, version
network: string                  # Network ID (recommended required for import determinism)
namespace: string                # Optional, namespace/package
owner: string                    # Optional, owner/team
enabled: boolean                 # Optional, whether enabled (default false recommended)
risk_level: low | medium | high  # Optional, risk level
requires_approval: boolean       # Optional, whether approval required
---
```

### Mixed Fragment (type: fragment)

```yaml
---
type: fragment                   # Mixed fragment
id: string                       # Fragment ID
name: string                     # Fragment name
version: string                  # Optional, version
network: string                  # Target network ID (recommended required for import determinism)
namespace: string                # Optional, namespace/package
owner: string                    # Optional, owner/team
---
```

### Delete Marker (type: delete)

```yaml
---
type: delete                     # Delete marker
network: string                  # Target network ID (recommended required for import determinism)
namespace: string                # Optional, namespace/package
owner: string                    # Optional, owner/team
targets:                         # Definitions to delete
  - entity: pod
  - relation: pod_belongs_node
  - action: restart_pod
---
```

---

## Entity Definition Specification

### Syntax

```markdown
## Entity: {entity_id}

**{Display Name}** - {Brief description}

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | {view_id} | {view_name} |

> **Primary Key**: `{primary_key}` | **Display Attribute**: `{display_key}`

### Property Override

(Optional) Declare only properties needing special configuration

| Property | Display Name | Index Config | Description |
|----------|--------------|--------------|-------------|
| ... | ... | ... | ... |

### Logic Properties

#### {property_name}

- **Type**: metric | operator
- **Source**: {source_id} ({source_type})
- **Description**: {description}

| Parameter | Source | Binding |
|-----------|--------|---------|
| ... | property | {property_name} |
| ... | input | - |
```

### Field Reference

| Field | Required | Description |
|-------|:--------:|-------------|
| entity_id | YES | Entity unique ID, lowercase letters, digits, underscores |
| Display Name | YES | Human-readable name |
| Data Source | YES | Mapped data view |
| Primary Key | YES | Primary key property name |
| Display Attribute | YES | Property used for display |
| Property Override | NO | Properties needing special configuration |
| Logic Properties | NO | Extended properties such as metrics, operators |

### Configuration Modes

#### Mode 1: Full Mapping (Simplest)

Map directly to view, inherit all fields automatically:

```markdown
## Entity: node

**Node**

### Data Source

| Type | ID |
|------|-----|
| data_view | view_123 |

> **Primary Key**: `id` | **Display Attribute**: `node_name`
```

#### Mode 2: Mapping + Property Override

Map to view, declare only properties needing special configuration:

```markdown
## Entity: pod

**Pod Instance**

### Data Source

| Type | ID |
|------|-----|
| data_view | view_456 |

> **Primary Key**: `id` | **Display Attribute**: `pod_name`

### Property Override

| Property | Index Config |
|----------|--------------|
| pod_status | fulltext + vector |
```

#### Mode 3: Full Definition

Declare all properties explicitly:

```markdown
## Entity: service

**Service**

### Data Source

| Type | ID |
|------|-----|
| data_view | view_789 |

### Data Properties

| Property | Display Name | Type | Description | PK | Index |
|----------|--------------|------|-------------|:--:|:-----:|
| id | ID | int64 | Primary key | YES | YES |
| service_name | Name | VARCHAR | Service name | | YES |

> **Display Attribute**: `service_name`
```

---

## Relation Definition Specification

### Syntax

```markdown
## Relation: {relation_id}

**{Display Name}** - {Brief description}

| Source | Target | Type |
|--------|--------|------|
| {source_entity} | {target_entity} | direct | data_view |

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| {source_prop} | {target_prop} |

### Business Semantics

(Optional) Description of relation business meaning...
```

> **Single-file format difference**: In a `type: relation` single-file, use `## Association Definition` as the section heading (H2) and `## Mapping Rules`. When embedded in network/fragment, the source/target table follows directly after `**Name**`, and use `### Mapping Rules` for the mapping section.

### Field Reference

| Field | Required | Description |
|-------|:--------:|-------------|
| relation_id | YES | Relation unique identifier |
| Source | YES | Source entity ID |
| Target | YES | Target entity ID |
| Type | YES | `direct` (direct mapping) or `data_view` (view mapping) |
| Mapping Rules | YES | Property mapping relationship |

### Relation Types

#### Direct Mapping (direct)

Associate via property value matching:

```markdown
## Relation: pod_belongs_node

| Source | Target | Type |
|--------|--------|------|
| pod | node | direct |

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| pod_node_name | node_name |
```

#### View Mapping (data_view)

Associate via intermediate view:

```markdown
## Relation: user_likes_post

| Source | Target | Type |
|--------|--------|------|
| user | post | data_view |

### Mapping View

| Type | ID |
|------|-----|
| data_view | user_post_likes_view |

### Source Mapping

| Source Property | View Property |
|-----------------|---------------|
| user_id | uid |

### Target Mapping

| View Property | Target Property |
|---------------|-----------------|
| pid | post_id |
```

---

## Action Definition Specification

### Syntax

```markdown
## Action: {action_id}

**{Display Name}** - {Brief description}

| Bound Entity | Action Type |
|--------------|--------------|
| {entity_id} | add | modify | delete |

### Trigger Condition

```yaml
field: {property_name}
operation: == | != | > | < | >= | <= | in | not_in | exist | not_exist
value: {value}
```

### Tool Configuration

| Type | Tool ID |
|------|--------|
| tool | {tool_id} |

or

| Type | MCP |
|------|-----|
| mcp | {mcp_id}/{tool_name} |

### Parameter Binding

| Parameter | Source | Binding |
|-----------|--------|---------|
| {param_name} | property | {property_name} |
| {param_name} | input | - |
| {param_name} | const | {value} |

### Schedule Configuration

(Optional)

| Type | Expression |
|------|------------|
| FIX_RATE | {interval} |
| CRON | {cron_expr} |

### Execution Description

(Optional) Detailed execution flow...
```

> **Single-file format difference**: In a `type: action` single-file, use `## Bound Entity` as the section heading (H2) and other sections one level down. When embedded in network/fragment, the binding table follows directly after `**Name**`, and sections use `###`.

### Governance Requirements (Strongly Recommended)

Action definitions connect to execution surface (tool/mcp). For stability and security, explicitly document the following in each Action and enforce through governance:

1. **Trigger**: When to trigger (manual/scheduled/conditional), whether conditions are reproducible
2. **Scope of impact**: Which objects, scope boundary, expected side effects
3. **Permissions and prerequisites**: Who can import/enable/execute, approval requirements, required credentials
4. **Rollback / failure strategy**: Failure handling, retry policy, circuit breaker/rate limit, reversibility

> Recommended practice: Import does not imply enablement; enablement and execution require separate permissions and audit logs, traceable to the corresponding BKN definition version.

### Field Reference

| Field | Required | Description |
|-------|:--------:|-------------|
| action_id | YES | Action unique identifier |
| Bound Entity | YES | Target entity ID |
| Action Type | YES | `add` / `modify` / `delete` |
| Trigger Condition | NO | Conditions for automatic trigger |
| Tool Configuration | YES | Tool or MCP to execute |
| Parameter Binding | YES | Parameter source configuration |
| Schedule Configuration | NO | Scheduled execution configuration |

### Trigger Condition Operators

| Operator | Description | Example |
|----------|-------------|---------|
| == | Equal | `value: Running` |
| != | Not equal | `value: Running` |
| > | Greater than | `value: 100` |
| < | Less than | `value: 100` |
| >= | Greater than or equal | `value: 100` |
| <= | Less than or equal | `value: 100` |
| in | In set | `value: [A, B, C]` |
| not_in | Not in set | `value: [A, B, C]` |
| exist | Exists | (no value needed) |
| not_exist | Does not exist | (no value needed) |
| range | In range | `value: [0, 100]` |

### Parameter Sources

| Source | Description |
|--------|-------------|
| property | From entity property |
| input | Runtime user input |
| const | Constant value |

---

## Common Syntax Elements

### Table Format

Use standard Markdown tables:

```markdown
| Col1 | Col2 | Col3 |
|------|------|------|
| Val1 | Val2 | Val3 |
```

Center alignment (for booleans):

```markdown
| Col1 | Col2 |
|------|:----:|
| Val1 | YES |
```

### YAML Code Blocks

For complex structures (e.g., conditional expressions):

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

### Mermaid Diagrams

For visualizing relations:

```markdown
```mermaid
graph LR
    A --> B
    B --> C
`` `
```

### Blockquote

For key information:

```markdown
> **Primary Key**: `id` | **Display Attribute**: `name`
```

### Heading Levels

Heading levels depend on file type, following Markdown hierarchy:

#### In network / fragment files (multiple embedded definitions)

- `#` — Network/fragment title
- `##` — Type definition (Entity:/Relation:/Action:)
- `###` — Section within definition (Data Source, Property Override, Logic Properties, etc.)
- `####` — Sub-item (logic property name)

#### In single-definition files (type: entity / relation / action)

- `#` — Definition title (entity/relation/action name)
- `##` — Section (Data Source, Property Override, Logic Properties, etc.)
- `###` — Sub-item (logic property name)

> Rule: Section semantics remain unchanged; heading level shifts with nesting depth. Parsers must support both patterns.

---

## File Organization

### Pattern 1: Single File (Small Networks)

All definitions in one `.bkn` file:

```markdown
---
type: network
id: my-network
---

# My Network

## Entity: entity1
...

## Entity: entity2
...

## Relation: rel1
...

## Action: action1
...
```

### Pattern 2: Split by Type (Medium Networks)

Use `index.bkn` to reference other files:

```markdown
---
type: network
id: my-network
includes:
  - entities.bkn
  - relations.bkn
  - actions.bkn
---

# My Network

Network description...
```

### Pattern 3: One Definition per File (Large Networks, Recommended)

Each entity, relation, and action in its own file:

```
k8s-network/
├── index.bkn                    # type: network
├── entities/
│   ├── pod.bkn                  # type: entity
│   ├── node.bkn                 # type: entity
│   └── service.bkn              # type: entity
├── relations/
│   ├── pod_belongs_node.bkn     # type: relation
│   └── service_routes_pod.bkn   # type: relation
└── actions/
    ├── restart_pod.bkn          # type: action
    └── cordon_node.bkn          # type: action
```

**Single entity file example** (`pod.bkn`):

```markdown
---
type: entity
id: pod
name: Pod Instance
network: k8s-network
---

# Pod Instance

Minimal deployable unit in Kubernetes.

## Data Source

| Type | ID |
|------|-----|
| data_view | view_123 |

> **Primary Key**: `id` | **Display Attribute**: `pod_name`
```

---

## Incremental Import Specification

BKN supports dynamically importing any `.bkn` file into an existing knowledge network.

### Importer Capability Requirements (for Engineering Control)

Implement a **BKN Importer** that converts BKN files into system changes with these capabilities (all required):

| Capability | Description | Purpose |
|------------|-------------|---------|
| `validate` | Structure/table/YAML block validation, reference integrity, parameter binding check | Prevent errors from entering the system |
| `diff` | Compute change set (add/update/delete) and impact scope | Make changes explainable and auditable |
| `dry_run` | Execute validate + diff without applying | Pre-deployment rehearsal |
| `apply` | Execute changes (per deterministic semantics and conflict strategy) | Controlled execution |
| `export` | Export knowledge network state to BKN (round-trip capable) | Prevent drift, rollback, reproducibility |

> Requirement: All import operations must record audit information (operator, timestamp, input file fingerprint, change set, result).

### Import Determinism (Required)

For multi-user collaboration and replay, import semantics must be **deterministic**:

- Same set of input files (ignoring filesystem order) yields same result
- Repeated import of the same file yields same result (idempotent)
- Conflicts must be explainable: either fail explicitly (fail-fast) or follow a defined rule (e.g., last-wins); no implicit merging

### Unique Key and Scope

The unique key for each definition is recommended as:

- `key = (network_id, type, id)`

Where `network_id` comes from:

- Prefer frontmatter `network`
- If missing, use import target network (import command parameter or `type: network` `id`)

### Update Semantics (replace vs merge)

**Replace (full overwrite)** is recommended by default:

- When `key` already exists, replace the old definition with the imported definition
- **Missing field does not mean delete**: Only means the field is not in this definition; deletion must be explicit (see `type: delete`)

If needed, **merge-by-section** may be supported under control, with:

- Merge limited to additive sections (e.g., Property Override, Logic Properties)
- Conflicts must be controllable: fail-fast or last-wins for same-name logic properties/field configs (configurable)
- Merge strategy must be explicitly configured in the importer and recorded in audit logs

### Conflict and Priority

When the same `key` is declared by multiple files in one import batch:

- Default: **fail-fast** (recommended for stability)
- Optional: Explicit priority ordering (e.g., command-line order or `priority` field); otherwise not recommended

### Import Behavior

| Scenario | Behavior |
|----------|----------|
| ID does not exist | Create new definition |
| ID exists | Update definition (overwrite) |
| Using `type: delete` | Delete specified definition |

### Import Examples

**Scenario: Add new entity to existing network**

Create `deployment.bkn`:

```markdown
---
type: entity
id: deployment
name: Deployment
network: k8s-network
---

# Deployment

Kubernetes deployment controller.

## Data Source

| Type | ID |
|------|-----|
| data_view | deployment_view |

> **Primary Key**: `id` | **Display Attribute**: `deployment_name`
```

After import, `k8s-network` will include the new `deployment` entity.

**Scenario: Update existing entity**

Create a file with the same ID; import will overwrite:

```markdown
---
type: entity
id: pod
name: Pod Instance (Updated)
network: k8s-network
---

# Pod Instance

Updated definition...
```

**Scenario: Delete definition**

```markdown
---
type: delete
network: k8s-network
targets:
  - entity: deprecated_entity
  - relation: old_relation
---

# Delete Deprecated Definitions

Clean up unused definitions.
```

**Scenario: Batch import (fragment)**

```markdown
---
type: fragment
id: monitoring-extension
name: Monitoring Extension
network: k8s-network
---

# Monitoring Extension

Add monitoring-related entities and actions.

## Entity: alert

**Alert**

### Data Source

| Type | ID |
|------|-----|
| data_view | alert_view |

> **Primary Key**: `id` | **Display Attribute**: `alert_name`

---

## Action: send_alert

**Send Alert**

| Bound Entity | Action Type |
|--------------|-------------|
| alert | add |

### Tool Configuration

| Type | Tool ID |
|------|---------|
| tool | alert_sender |
```

---

## Patch Specification (File Level)

### Add Operation

```markdown
---
type: patch
id: 2026-01-31-add-metric
target: k8s-topology.bkn
operation: add
---

# Add CPU Metric

Add after `### Logic Properties` in `## Entity: pod`:

#### cpu_usage

- **Type**: metric
- **Source**: cpu_metric
```

### Modify Operation

```markdown
---
type: patch
id: 2026-01-31-update-condition
target: k8s-topology.bkn
operation: modify
---

# Update Trigger Condition

Modify the trigger condition of `## Action: restart_pod` to:

```yaml
field: pod_status
operation: in
value: [Unknown, Failed, CrashLoopBackOff]
`` `
```

### Delete Operation

```markdown
---
type: patch
id: 2026-01-31-remove-action
target: k8s-topology.bkn
operation: delete
---

# Delete Deprecated Action

Delete `## Action: deprecated_action`
```

---

## Best Practices

### Naming Conventions

- **ID**: Lowercase letters, digits, underscores (e.g., `pod_belongs_node`)
- **Display name**: Concise and clear (e.g., "Pod belongs to Node")
- **Tags**: Use a consistent tag system

### Document Structure

1. Put network description at the beginning
2. Use mermaid diagrams for overall topology
3. Entity definitions first, then relations and actions
4. Group related definitions together

### Simplicity

- Prefer full mapping mode
- Declare property overrides only when needed
- Avoid duplicate information

### Readability

- Use tables for structured data
- Add business semantics
- Use mermaid diagrams when helpful

---

## References

- [Architecture Design](./ARCHITECTURE.md)
- Examples:
  - [Single-file mode](./examples/k8s-topology.bkn) — All definitions in one file
  - [Split by type](./examples/k8s-network/) — Entities, relations, actions in separate files
  - [One definition per file](./examples/k8s-modular/) — Each definition in its own file (recommended for large-scale use)
