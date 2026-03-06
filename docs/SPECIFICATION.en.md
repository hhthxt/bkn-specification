# BKN Language Specification

Version: 1.0.0
spec_version: 1.0.0

## Overview

BKN (Business Knowledge Network) is a declarative modeling language based on Markdown for defining objects, relations, and actions in business knowledge networks. BKN describes model structure and semantics only — runtime capabilities such as validation engines, data pipelines, and workflows are provided by platforms that consume BKN models.

This document defines the complete syntax specification for BKN.

### Glossary

**Core Concepts**

| Term | Meaning |
|------|---------|
| BKN | Business Knowledge Network |
| knowledge_network | The overall collection of a business knowledge network |
| object | Business object type (e.g., Pod/Node/Service) |
| relation | Relationship type connecting two objects (e.g., belongs_to, routes_to) |
| action | Operation definition on an object (can bind to tool or mcp) |

**Object Structure**

| Term | Meaning |
|------|---------|
| data_view | Data view; the data source an object or relation maps to |
| data_properties | Object property definition table; declares field types, primary key, display key, etc. |
| property_override | Property override; special configuration for inherited properties (index, constraint) |
| logic_properties | Logic properties; derived fields from external sources (metric / operator) |
| primary_key | Primary key field; uniquely identifies an instance (marked YES in Data Properties) |
| display_key | Display key field; used for UI display and search (marked YES in Data Properties) |
| constraint | Property value constraint; declares valid ranges for instance data (e.g., `>= 0`, `in(...)`) |
| metric | Logic property type: a measured value from an external data source |
| operator | Logic property type: computed logic based on input parameters |

**Action Structure**

| Term | Meaning |
|------|---------|
| trigger_condition | Trigger condition; defines when an action executes automatically |
| pre-conditions | Pre-conditions; data checks that must pass before execution (blocks if unsatisfied) |
| tool | External tool bound to an action |
| mcp | Model Context Protocol; MCP tool bound to an action |
| schedule | Timing configuration (FIX_RATE or CRON) for periodic execution |
| scope_of_impact | Scope of impact; declares objects affected by an action |

**File Organization**

| Term | Meaning |
|------|---------|
| frontmatter | YAML metadata block (wrapped in `---`) at the top of every .bkn file |
| network | File type `type: network`; top-level container for a complete knowledge network |
| fragment | File type `type: fragment`; mixed snippet containing multiple object/relation/action definitions |
| data | File type `type: data`; instance data file (recommended `.bknd` extension) |
| delete | File type `type: delete`; explicitly declares definitions to be removed |
| patch | File type `type: patch`; incremental modification to an existing file |
| namespace | Namespace; used for large-scale organization and avoiding ID conflicts |
| spec_version | Specification version; identifies which BKN spec version a file conforms to |

### Primitives (Canonical Section and Table Terms)

The table below uses a **unified heading hierarchy** that applies to all file types (network / fragment / object / relation / action / data).

| Level | English (canonical) | Definition | Syntax |
|:-----:|---------------------|------------|--------|
| `#` | Objects | Section: all object definitions in this file | `# Objects` |
| `#` | Relations | Section: all relation definitions | `# Relations` |
| `#` | Actions | Section: all action definitions | `# Actions` |
| `##` | Object | Individual object definition | `## Object: {id}` |
| `##` | Relation | Individual relation definition | `## Relation: {id}` |
| `##` | Action | Individual action definition | `## Action: {id}` |
| `###` | Data Source | The data view this object maps from | `### Data Source` |
| `###` | Data Properties | Explicit list of fields (name, type, PK, index) | `### Data Properties` |
| `###` | Property Override | Per-property overrides (e.g. index config) | `### Property Override` |
| `###` | Logic Properties | Derived fields: metric, operator | `### Logic Properties` |
| `###` | Business Semantics | Human-readable meaning of the object/relation | `### Business Semantics` |
| `###` | Endpoints | Relation endpoints: source, target, type | `### Endpoints` |
| `###` | Mapping Rules | How source/target properties map | `### Mapping Rules` |
| `###` | Mapping View | For data_view relations: the join view | `### Mapping View` |
| `###` | Source Mapping | Map source object props to view | `### Source Mapping` |
| `###` | Target Mapping | Map view to target object props | `### Target Mapping` |
| `###` | Bound Object | Object this action operates on | `### Bound Object` |
| `###` | Trigger Condition | When to run (YAML condition) | `### Trigger Condition` |
| `###` | Pre-conditions | Data conditions required before action execution | `### Pre-conditions` |
| `###` | Tool Configuration | tool or MCP binding | `### Tool Configuration` |
| `###` | Parameter Binding | param name, source, binding | `### Parameter Binding` |
| `###` | Schedule | FIX_RATE or CRON | `### Schedule` |
| `###` | Scope of Impact | What objects are affected | `### Scope of Impact` |
| `####` | {property_name} | Individual logic property sub-section | `#### {name}` |
| — | Primary Key | Field that uniquely identifies an instance | Data Properties table column |
| — | Display Key | Field used for UI label / search display | Data Properties table column |
| — | Action Type | add \| modify \| delete | table column |

Table column names (canonical): Type, ID, Name, Property, Display Name, Constraint, Primary Key, Display Key, Index, Index Config, Description; Source, Target, Required, Min, Max; Source Property, Target Property; Parameter, Type, Source, Binding, Description; Bound Object, Action Type; Object, Check, Condition, Message; Object, Impact Description.

## File Format

### File Extension

- `.bkn` — BKN schema/definition file (recommended)
- `.bknd` — BKN data file (instance data, recommended)
- `.md` — Compatible carrier; runtime supports it; content must satisfy BKN frontmatter/type/structure constraints

**`.md` compatibility mode**: BKN content may be saved as `.md` for cross-platform documentation and collaboration. At runtime, `.md` files follow the same parse and validation path as `.bkn`/`.bknd`; missing frontmatter, `type`, or invalid structure will fail immediately. Recommended: use `.bkn` for schema, `.bknd` for data; use `.md` when coexisting with generic Markdown tooling.

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

## Object: object_id

Object definition...

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
| `namespace` | object/relation/action/fragment/delete | Namespace/package name for large-scale organization and conflict avoidance (e.g., `platform.k8s`) |
| `owner` | object/relation/action/fragment/delete | Owner/team (for audit and approval routing) |
| `enabled` | action | Whether enabled (default `false` recommended; import does not imply enablement) |
| `risk_level` | action | Risk level (`low|medium|high` for approval and release strategy) |
| `requires_approval` | action | Whether approval is required to enable/execute |

### File Types (type)

| type | Description | Purpose |
|------|-------------|---------|
| `network` | Full knowledge network | Network file containing multiple definitions |
| `object` | Single object definition | Standalone object file, directly importable |
| `relation` | Single relation definition | Standalone relation file, directly importable |
| `action` | Single action definition | Standalone action file, directly importable |
| `fragment` | Mixed fragment | Contains multiple types of partial definitions |
| `data` | Data file | Carries instance rows for object/relation definitions (recommended `.bknd`) |
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

### Single Object File (type: object)

```yaml
---
type: object                     # Single object definition
id: string                       # Object ID, unique identifier
name: string                     # Object display name
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
risk_level: low | medium | high  # Optional, static risk level
requires_approval: boolean       # Optional, whether approval required
---
```

> **Dynamic risk property**: The Action runtime property `risk` (values `allow` | `not_allow`) is computed by the built-in or a user-provided risk evaluation function from the current scenario and knowledge tagged with `__risk__`; it is not declared in this frontmatter.

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
  - object: pod
  - relation: pod_belongs_node
  - action: restart_pod
---
```

---

## Data Files (`.bknd` / `type: data`)

`.bknd` files use the same Markdown syntax as `.bkn`, but the body carries instance rows instead of object/relation/action definitions.

### Frontmatter

```yaml
---
type: data
network: recoverable-network
object: scenario            # mutually exclusive with relation
source: PFMEA模板.xlsx      # optional data provenance
---
```

- `type` must be `data`
- `object` or `relation` must be set (mutually exclusive), pointing to an ID defined in `.bkn`
- `network` is recommended for consistency with schema files
- `source` is optional provenance metadata

### Body

Use one heading (`#` or `##`) plus one GFM table. Table headers should align with target object Data Properties (or relation mapping fields).

```markdown
# scenario

| scenario_id | name | category | primary_object | description |
|-------------|------|----------|----------------|-------------|
| ops-rm-rf | rm -rf 删除备份存储 | integrity | backup_system | 直接销毁备份数据 |
```

### Constraints

- Column names should align with schema definitions to avoid implicit fields
- A single `type: data` file should contain one table for better versioning and auditing

---

## Object Definition Specification

### Syntax

```markdown
## Object: {object_id}

**{Display Name}** - {Brief description}

- **Tags**: {tag1}, {tag2}     (optional, definition-level tags)
- **Owner**: {owner}           (optional, owner/team)

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | {view_id} | {view_name} |

### Data Properties

| Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index |
|----------|--------------|------|------------|-------------|:-----------:|:-----------:|:-----:|
| {prop} | {name} | {type} | | {desc} | YES | | YES |
| {prop} | {name} | {type} | | {desc} | | YES | |

- `Primary Key`: Property marked `YES` uniquely identifies an instance; at least one required
- `Display Key`: Property marked `YES` is used for UI display and search; at least one required
- `Constraint` column is optional; declares valid value ranges for instance data. Leave empty for no constraint. See "Constraint Column Syntax" below for details

### Property Override

(Optional) Declare only properties needing special configuration

| Property | Display Name | Index Config | Constraint | Description |
|----------|--------------|--------------|------------|-------------|
| ... | ... | ... | ... | ... |

#### Index Config Syntax

The `Index Config` column supports a combined syntax; multiple index types are joined with ` + `. Optional parameters may be passed in parentheses:

| Type | Syntax | Description |
|------|--------|-------------|
| keyword | `keyword` | Basic keyword index |
| keyword | `keyword(max_len)` | Keyword index with ignore_above_len |
| fulltext | `fulltext` | Full-text index, default analyzer |
| fulltext | `fulltext(analyzer)` | Full-text index with specific analyzer (e.g. standard, ik_max_word) |
| vector | `vector` | Vector index, default model |
| vector | `vector(model_id)` | Vector index with specified embedding model ID |

Example: `keyword(1024) + fulltext(standard) + vector(1951511856216674304)`

### Logic Properties

#### {property_name}

- **Type**: metric | operator
- **Source**: {source_id} ({source_type})
- **Description**: {description}

| Parameter | Type | Source | Binding | Description |
|-----------|------|--------|---------|-------------|
| ... | string | property | {property_name} | Bind from object property |
| ... | array | input | - | Runtime user input |
| ... | string | const | {value} | Constant value |
```

- `Type`: Parameter data type (e.g. string, number, boolean, array)
- `Source`: Value source — `property` (object property) / `input` (user input) / `const` (constant)
- `Binding`: When Source is property, the property name; when const, the constant value; when input, `-`

### Definition-Level Metadata

In the header of a `## Object:` or `## Relation:` definition (before `### Data Source` or `### Endpoints`), optional inline metadata lines may be used:

- **Tags**: Tag list for this definition (comma-separated), for categorization, filtering, and audit
- **Owner**: Owner or team, for approval routing and audit

In fragment or network files, multiple objects or relations may each have different tags and owner.

### Risk-Related Definitions

- **Reserved tag**: **`__risk__`** is a built-in reserved tag used only for objects and relations that participate in the built-in risk assessment. **Users must not use `__risk__` for custom purposes**, to avoid conflicting with built-in behavior.
- Add `- **Tags**: __risk__` in the definition header of objects and relations that participate in the built-in risk evaluation. AI applications and the built-in evaluator use this tag to identify risk-related definitions.
- Actions have a **runtime/computed property** `risk` (see Action definition section), with values `allow` | `not_allow`, computed by the built-in or a user-provided risk evaluation function from the current scenario and data tagged with `__risk__`; it is **not written in BKN files**.

**Openness**: Users may define **their own risk-like classes** (using **non-reserved** tags, e.g. `compliance`, `audit`) and **their own risk evaluation functions**; the built-in `__risk__` and default evaluator are one optional implementation and do not preclude extension or replacement.

### Field Reference

| Field | Required | Description |
|-------|:--------:|-------------|
| {object_id} | YES | Object unique ID, lowercase letters, digits, underscores |
| {display_name} | YES | Human-readable name |
| Data Source | NO | Mapped data view; managed by the platform automatically when omitted |
| Data Properties | YES | Property definitions; must mark Primary Key and Display Key |
| Property Override | NO | Properties needing special configuration (index, constraints) |
| Logic Properties | NO | Extended properties such as metrics, operators |

### Data Types

The `Type` column in Data Properties tables uses the following standard types. Type names are case-insensitive; the canonical forms below are recommended.

| Type | Description | JSON Mapping | SQL Mapping |
|------|-------------|-------------|-------------|
| int32 | 32-bit signed integer | number | INT / INTEGER |
| int64 | 64-bit signed integer | number | BIGINT |
| integer | Generic integer (precision unspecified) | number | Platform-dependent (typically int64) |
| float32 | 32-bit floating point | number | FLOAT / REAL |
| float64 | 64-bit floating point | number | DOUBLE / DOUBLE PRECISION |
| float | Generic floating point (precision unspecified) | number | Platform-dependent (typically float64) |
| decimal(p,s) | Exact decimal; p = precision, s = scale | string / number | DECIMAL(p,s) / NUMERIC(p,s) |
| decimal | Generic exact decimal (precision unspecified) | string / number | Platform-dependent |
| bool | Boolean | boolean | BOOLEAN |
| VARCHAR | Variable-length string | string | VARCHAR / TEXT |
| TEXT | Long text | string | TEXT / CLOB |
| DATE | Date (no time) | string (ISO 8601) | DATE |
| TIME | Time (no date) | string (ISO 8601) | TIME |
| TIMESTAMP | Date and time (with timezone) | string (ISO 8601) | TIMESTAMP |
| JSON | JSON structured data | object / array | JSON / JSONB |
| BINARY | Binary data | string (base64) | BLOB / BYTEA |

> When the data source uses a type not listed above, the source-native type name may be used directly (e.g. `ARRAY<VARCHAR>`). Parsers should pass through unrecognized types as-is.

### Configuration Modes

#### Mode 1: Mapping + Minimal Properties

Map to view, declare only primary key and display key:

```markdown
## Object: node

**Node**

### Data Source

| Type | ID |
|------|-----|
| data_view | view_123 |

### Data Properties

| Property | Primary Key | Display Key |
|----------|:-----------:|:-----------:|
| id | YES | |
| node_name | | YES |
```

#### Mode 2: Mapping + Property Override

Map to view, declare keys and configure properties needing special treatment:

```markdown
## Object: pod

**Pod Instance**

### Data Source

| Type | ID |
|------|-----|
| data_view | view_456 |

### Data Properties

| Property | Primary Key | Display Key |
|----------|:-----------:|:-----------:|
| id | YES | |
| pod_name | | YES |

### Property Override

| Property | Index Config | Constraint | Description |
|----------|--------------|------------|-------------|
| pod_status | fulltext(standard) + vector | in(Running,Pending,Failed,Unknown) | Full-text and semantic search |
```

#### Mode 3: Full Definition

Declare all properties explicitly (with types, constraints, indexes):

```markdown
## Object: service

**Service**

### Data Source

| Type | ID |
|------|-----|
| data_view | view_789 |

### Data Properties

| Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index |
|----------|--------------|------|------------|-------------|:-----------:|:-----------:|:-----:|
| id | ID | int64 | | Primary key | YES | | YES |
| service_name | Name | VARCHAR | not_null | Service name | | YES | YES |
| service_type | Service Type | VARCHAR | in(ClusterIP,NodePort,LoadBalancer) | Service type | | | |
```

---

## Relation Definition Specification

### Syntax

```markdown
## Relation: {relation_id}

**{Display Name}** - {Brief description}

- **Tags**: {tag1}, {tag2}     (optional, definition-level tags)
- **Owner**: {owner}           (optional, owner/team)

| Source | Target | Type | Required | Min | Max |
|--------|--------|------|----------|-----|-----|
| {source} | {target} | direct \| data_view | YES \| NO | 0 | - |

- `Required`: YES/NO, whether at least one relation must exist (from Source side)
- `Min`: Minimum relation count, default 0
- `Max`: Maximum relation count, `-` means unlimited
- Required / Min / Max are optional columns; omit to apply no constraint

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| {source_prop} | {target_prop} |

### Business Semantics

(Optional) Description of relation business meaning...
```

### Field Reference

| Field | Required | Description |
|-------|:--------:|-------------|
| {relation_id} | YES | Relation unique identifier |
| Source | YES | Source object ID |
| Target | YES | Target object ID |
| Type | YES | `direct` (direct mapping) or `data_view` (view mapping) |
| Mapping Rules | YES | Property mapping relationship |
| Required | NO | Whether at least one relation must exist (from Source side) |
| Min | NO | Minimum relation count |
| Max | NO | Maximum relation count, `-` means unlimited |

### Relation Types

#### Direct Mapping (direct)

Associate via property value matching:

```markdown
## Relation: pod_belongs_node

| Source | Target | Type | Required | Min | Max |
|--------|--------|------|----------|-----|-----|
| pod | node | direct | YES | 1 | 1 |

Each Pod must belong to exactly one Node.

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| pod_node_name | node_name |
```

#### View Mapping (data_view)

Associate via intermediate view:

```markdown
## Relation: user_likes_post

| Source | Target | Type | Required | Min | Max |
|--------|--------|------|----------|-----|-----|
| user | post | data_view | NO | 0 | - |

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

| Bound Object | Action Type |
|--------------|--------------|
| {object_id} | add | modify | delete |

### Trigger Condition

```yaml
field: {property_name}
operation: == | != | > | < | >= | <= | in | not_in | exist | not_exist
value: {value}
```

### Pre-conditions

(Optional) Data conditions required before execution; if not satisfied, action is blocked

| Object | Check | Condition | Message |
|--------|-------|-----------|---------|
| {object_id} | relation:{relation_id} | exist | Violation message |
| {object_id} | property:{property_name} | {op} {value} | Violation message |

- `Check`: `property:{name}` or `relation:{id}`, specifies what to check
- `Condition`: Reuses Trigger Condition operator syntax
- Trigger determines "when to run"; Pre-conditions determine "whether execution is allowed"

### Tool Configuration

| Type | Tool ID |
|------|--------|
| tool | {tool_id} |

or

| Type | MCP |
|------|-----|
| mcp | {mcp_id}/{tool_name} |

### Parameter Binding

| Parameter | Type | Source | Binding | Description |
|-----------|------|--------|---------|-------------|
| {param_name} | string | property | {property_name} | {description} |
| {param_name} | string | input | - | {description} |
| {param_name} | string | const | {value} | {description} |

### Schedule Configuration

(Optional)

| Type | Expression |
|------|------------|
| FIX_RATE | {interval} |
| CRON | {cron_expr} |

### Execution Description

(Optional) Detailed execution flow...
```

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
| {action_id} | YES | Action unique identifier |
| Bound Object | YES | Target object ID |
| Action Type | YES | `add` / `modify` / `delete` |
| Trigger Condition | NO | Conditions for automatic trigger |
| Pre-conditions | NO | Data conditions required before execution |
| Tool Configuration | YES | Tool or MCP to execute |
| Parameter Binding | YES | Parameter source configuration |
| Schedule Configuration | NO | Scheduled execution configuration |
| risk (computed) | - | Runtime property: `allow` \| `not_allow`, computed by built-in or user-provided evaluator from scenario and data tagged with `__risk__`; not written in BKN |

### Trigger Condition Operators

These operators apply to Trigger Condition, Pre-conditions, and the Constraint column in Data Properties / Property Override tables:

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
| not_null | Not null | (no value needed, constraint-specific) |
| regex | Regex match | `value: "^[a-z]+$"` (constraint-specific) |

### Constraint Column Syntax

The `Constraint` column appears in the **Data Properties** and **Property Override** tables within an Object definition. It declares valid value ranges that instance data must satisfy. The column is optional; an empty cell means no constraint.

#### Format

Each constraint is written in a single table cell using the format **`operator`**, **`operator(args)`**, or **`operator value`**.

| Category | Syntax | Meaning | Applicable Types | Example |
|----------|--------|---------|------------------|---------|
| Comparison | `== value` | Equal to fixed value | Numeric, String | `== 1` |
| Comparison | `!= value` | Not equal to fixed value | Numeric, String | `!= 0` |
| Comparison | `> value` | Greater than | Numeric | `> 0` |
| Comparison | `< value` | Less than | Numeric | `< 1000` |
| Comparison | `>= value` | Greater than or equal | Numeric | `>= 0` |
| Comparison | `<= value` | Less than or equal | Numeric | `<= 100` |
| Range | `range(min,max)` | Closed interval [min, max] | Numeric | `range(0,100)` |
| Enumeration | `in(v1,v2,…)` | Value must be one of the list | String, Numeric | `in(Running,Pending,Failed)` |
| Enumeration | `not_in(v1,v2,…)` | Value must not be in the list | String, Numeric | `not_in(Deleted,Archived)` |
| Existence | `not_null` | Value must not be null | Any | `not_null` |
| Existence | `exist` | Property must exist | Any | `exist` |
| Existence | `not_exist` | Property must not exist | Any | `not_exist` |
| Pattern | `regex:pattern` | Value must match regex | String | `regex:^[a-z0-9_]+$` |

#### Combining Constraints

When a property requires multiple constraints, separate them with `; ` (semicolon + space):

```
not_null; >= 0
not_null; regex:^[a-z_]+$
>= 0; <= 100
not_null; in(ClusterIP,NodePort,LoadBalancer)
```

Combined constraints use **logical AND** — all constraints must be satisfied simultaneously.

#### Full Example

| Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index |
|----------|--------------|------|------------|-------------|:-----------:|:-----------:|:-----:|
| id | ID | int64 | not_null | Primary key | YES | | YES |
| name | Name | VARCHAR | not_null; regex:^[a-z0-9_]+$ | Unique identifier | | YES | YES |
| quantity | Quantity | int32 | >= 0 | No negatives allowed | | | |
| status | Status | VARCHAR | in(Active,Inactive,Archived) | Enum values | | | YES |
| score | Score | float64 | range(0,100) | Percentage | | | |
| priority | Priority | int32 | not_null; range(1,5) | Level 1–5 | | | |

### Parameter Sources

| Source | Description |
|--------|-------------|
| property | From object property |
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
> **Note**: This object requires an approval workflow
```

### Heading Levels

Heading levels are consistent across all file types:

- `#` - Document/group heading (for example network title, or `# Objects` / `# Relations` / `# Actions`)
- `##` - Definition heading (`## Object:` / `## Relation:` / `## Action:`)
- `###` - In-definition sections (Data Source, Data Properties, Mapping Rules, Trigger Condition, etc.)
- `####` - Sub-items (for example logic property names)

> Rule: There is no longer a “single-file level shift”; all definitions use the hierarchy above.
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

## Object: object1
...

## Object: object2
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
  - objects.bkn
  - relations.bkn
  - actions.bkn
---

# My Network

Network description...
```

### Pattern 3: One Definition per File (Large Networks, Recommended)

Each object, relation, and action in its own file:

```
k8s-network/
├── index.bkn                    # type: network
├── objects/
│   ├── pod.bkn                  # type: object
│   ├── node.bkn                 # type: object
│   └── service.bkn              # type: object
├── relations/
│   ├── pod_belongs_node.bkn     # type: relation
│   └── service_routes_pod.bkn   # type: relation
└── actions/
    ├── restart_pod.bkn          # type: action
    └── cordon_node.bkn          # type: action
```

**Single object file example** (`pod.bkn`):

```markdown
---
type: object
id: pod
name: Pod Instance
network: k8s-network
---

## Object: pod

**Pod Instance**

Minimal deployable unit in Kubernetes.

## Data Source

| Type | ID |
|------|-----|
| data_view | view_123 |

## Data Properties

| Property | Primary Key | Display Key |
|----------|:-----------:|:-----------:|
| id | YES | |
| pod_name | | YES |
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

**Scenario: Add new object to existing network**

Create `deployment.bkn`:

```markdown
---
type: object
id: deployment
name: Deployment
network: k8s-network
---

## Object: deployment

**Deployment**

Kubernetes deployment controller.

## Data Source

| Type | ID |
|------|-----|
| data_view | deployment_view |

## Data Properties

| Property | Primary Key | Display Key |
|----------|:-----------:|:-----------:|
| id | YES | |
| deployment_name | | YES |
```

After import, `k8s-network` will include the new `deployment` object.

**Scenario: Update existing object**

Create a file with the same ID; import will overwrite:

```markdown
---
type: object
id: pod
name: Pod Instance (Updated)
network: k8s-network
---

## Object: pod

**Pod Instance (Updated)**

Updated definition...
```

**Scenario: Delete definition**

```markdown
---
type: delete
network: k8s-network
targets:
  - object: deprecated_object
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

Add monitoring-related objects and actions.

## Object: alert

**Alert**

### Data Source

| Type | ID |
|------|-----|
| data_view | alert_view |

### Data Properties

| Property | Primary Key | Display Key |
|----------|:-----------:|:-----------:|
| id | YES | |
| alert_name | | YES |

---

## Action: send_alert

**Send Alert**

| Bound Object | Action Type |
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

Add after `### Logic Properties` in `## Object: pod`:

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
3. Object definitions first, then relations and actions
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
  - [Split by type](./examples/k8s-network/) — Objects, relations, actions in separate files
  - [One definition per file](./examples/k8s-modular/) — Each definition in its own file (recommended for large-scale use)
