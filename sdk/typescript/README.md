# BKN TypeScript SDK

Parse, validate, and transform [BKN (Business Knowledge Network)](../../docs/SPECIFICATION.md) files.

## Installation

```bash
npm install
npm run build
```

## Usage

```typescript
import {
  loadNetwork,
  parseBkn,
  validateNetwork,
  generateChecksum,
  verifyChecksum,
} from "bkn";

// Load a network directory (auto-discovers network.bkn)
const network = await loadNetwork("examples/k8s-network");

// Parse a single file
const doc = parseBkn(text, { sourcePath: "path/to/file.bkn" });

// Validate network data tables
const result = validateNetwork(network);
if (!result.ok) {
  console.error(result.errors);
}

// Generate and verify checksum
await generateChecksum("examples/k8s-network");
const { ok, errors } = await verifyChecksum("examples/k8s-network");
```

## API

- `parseBkn(text, options?)` — Parse .bkn/.bknd text into BknDocument
- `loadBknFile(path, options?)` — Load and parse a single file
- `loadNetwork(pathOrDir, options?)` — Load network with includes
- `discoverRootFile(directory)` — Find network root file
- `validateDocument(doc, options?)` — Validate document structure
- `validateNetwork(network, options?)` — Validate all data tables
- `validateDataTable(table, schema?, network?)` — Validate a data table
- `generateChecksum(path, options?)` — Generate CHECKSUM file
- `verifyChecksum(path, options?)` — Verify CHECKSUM file
- `packToTar(sourceDir, outputPath, options?)` — Pack BKN directory to tar (macOS: COPYFILE_DISABLE=1)

## Compatibility

- **strict**: Follow spec canonical forms
- **compat**: Accept example drift (`type: knowledge_network`, `Affect Object` → `Scope of Impact`, etc.)

## Tests

```bash
npm test
```
