# TypeScript SDK MVP API Reference

Frozen API and compatibility policy for Phase 1. Aligned with Python baseline and spec.

## Compatibility Policy

- **strict**: Follow spec canonical forms (type: network, English section names, etc.)
- **compat**: Accept observed drift in examples:
  - `type: knowledge_network` in addition to `type: network`
  - Section alias `Affect Object` → `Scope of Impact`
  - Checksum filename: `CHECKSUM` (Go/examples) or `checksum.txt` (Python/spec)

## Public API

### Parse

```ts
parseBkn(text: string, options?: ParseOptions): BknDocument
parseFrontmatter(text: string): Frontmatter
parseBody(text: string): { objects, relations, actions, risks, connections }
```

### Load

```ts
loadBknFile(path: string | PathLike, options?: LoadOptions): Promise<BknDocument>
loadNetwork(pathOrDir: string | PathLike, options?: LoadOptions): Promise<BknNetwork>
discoverRootFile(directory: string | PathLike): Promise<string>
```

### Validate

```ts
validateDocument(doc: BknDocument, options?: ValidateOptions): ValidationResult
validateNetwork(network: BknNetwork, options?: ValidateOptions): ValidationResult
validateDataTable(table: DataTable, schema?: BknObject, network?: BknNetwork): ValidationResult
```

### Checksum

```ts
generateChecksum(path: string | PathLike, options?: ChecksumOptions): Promise<string>
verifyChecksum(path: string | PathLike, options?: ChecksumOptions): Promise<VerifyResult>
```

### Tar

```ts
packToTar(sourceDir: string, outputPath: string, options?: PackToTarOptions): Promise<void>
```

Packs a BKN directory into a tar archive. On macOS, sets `COPYFILE_DISABLE=1` when spawning tar to prevent AppleDouble (`._*.bkn`) files that would cause Go SDK parsing errors.

## Data Models (from Python models.py)

- Frontmatter, DataSource, ConnectionConfig, Connection
- DataProperty, PropertyOverride, LogicProperty, LogicPropertyParameter
- Endpoint, MappingRule, ToolConfig, PreCondition, Schedule
- BknObject, Relation, Action, Risk, RiskScope, RiskStrategy, RiskPreCheck
- DataTable, BknDocument, BknNetwork

## ValidationResult

```ts
interface ValidationResult {
  ok: boolean
  errors: ValidationError[]
}
interface ValidationError {
  table: string
  row: number | null
  column: string
  code: string
  message: string
}
```

## Checksum Format (compat with examples)

- Filename: `CHECKSUM` (primary), `checksum.txt` (fallback for read)
- Line format: `{key}  sha256:{16hex}` where key is `object_type:id`, `SKILL.md`, etc.
- Aggregate: `*  sha256:{hash}` (optional)
