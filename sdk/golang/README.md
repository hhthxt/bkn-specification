# BKN Golang SDK

Go SDK for parsing, validating, and transforming BKN files. Provides feature parity with the [Python SDK](../python/README.md) for core functionality.

## Requirements

- **Go 1.24+**

## Status

Implemented. The Go SDK supports:

- Parse `.bkn`, `.bknd`, and `.md` files (YAML frontmatter + Markdown body). `.md` is a compatible carrier; content must satisfy BKN frontmatter/type/structure. Recommended: schema `.bkn`, data `.bknd`.
- Structured models for `BknObject`, `Relation`, `Action`, `Risk`, `Connection`, and `DataTable`
- Network loading: directory or file input; root discovery (network.bkn > network.md > index.bkn > index.md); `includes` resolution (cycle detection); implicit same-dir when no includes
- Network reference validation for shared `connection` data sources
- Data validation against object schema (not_null, regex, in, range, type checks, PK uniqueness)
- `.bknd` writability guard for object schemas backed by `data_view` or `connection`
- Serialization to `.bknd` format
- Risk evaluation (`EvaluateRisk`)

Transformers (e.g., kweaver) are not yet implemented.

## Structure

```
golang/
├── go.mod
├── bkn/
│   ├── models.go      # Data structures
│   ├── parser.go      # Parse .bkn/.bknd/.md
│   ├── loader.go      # Load, LoadNetwork
│   ├── validator.go   # ValidateDataTable, ValidateNetworkData
│   ├── serializer.go  # ToBknd, ToBkndFromTable
│   ├── risk.go        # EvaluateRisk
│   └── bkn_test.go    # Tests
└── README.md
```

## 如何引入（How to Use）

### 方式一：作为依赖引入

在您的 Go 项目中执行：

```bash
go get github.com/kweaver-ai/bkn-specification/sdk/golang
```

然后导入包：

```go
import "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
```

### 方式二：本地开发 / 私有仓库

若仓库未发布或需使用本地路径，在您的 `go.mod` 中添加：

```go
replace github.com/kweaver-ai/bkn-specification/sdk/golang => /path/to/bkn-specification/sdk/golang
```

### 方式三：复制到项目中

将 `sdk/golang/bkn/` 目录复制到您的项目内，修改 import 路径即可。

---

## Usage

```go
package main

import (
    "fmt"
    "github.com/kweaver-ai/bkn-specification/sdk/golang/bkn"
)

func main() {
    // Load a network (resolves includes). Path can be a file or directory; for dir, root is auto-discovered.
    net, err := bkn.LoadNetwork("examples/risk/index.bkn")
    // Or: net, err := bkn.LoadNetwork("examples/k8s-network")
    if err != nil {
        panic(err)
    }

    // Access objects, relations, actions, data tables
    objects := net.AllObjects()
    tables := net.AllDataTables()

    // Validate data against schema
    result := bkn.ValidateNetworkData(net)
    if !result.OK() {
        for _, e := range result.Errors {
            fmt.Println(e)
        }
    }

    // Risk evaluation
    rules := []map[string]any{
        {"scenario_id": "sec_t_01", "action_id": "restart_erp", "allowed": false, "risk_level": 5, "reason": "月末封网"},
    }
    result := bkn.EvaluateRisk(net, "restart_erp", map[string]any{"scenario_id": "sec_t_01"}, rules)
    fmt.Println(result.Decision)  // "not_allow"
    if result.RiskLevel != nil {
        fmt.Println(*result.RiskLevel) // 5
    }
    fmt.Println(result.Reason) // "月末封网"

    // Custom evaluator
    myEvaluator := func(network *bkn.BknNetwork, actionID string, context map[string]any, riskRules []map[string]any) bkn.RiskResult {
        if actionID == "grant_root_admin" {
            lv := 5
            return bkn.RiskResult{Decision: bkn.NotAllow, RiskLevel: &lv, Reason: "全局禁止提权"}
        }
        return bkn.RiskResult{Decision: bkn.Unknown}
    }
    result2 := bkn.EvaluateRiskWith(myEvaluator, net, "grant_root_admin", map[string]any{}, nil)
    fmt.Println(result2.Decision) // "not_allow"
}
```

### Update model (no-patch)

- **Add/modify**: Import `.bkn` files; each definition is upserted by `(network, type, id)`.
- **Delete**: Use `PlanDelete(...)` / `NetworkWithout(...)` for planning and in-memory simulation; deletion is not expressed in BKN files, and persistence is handled by the consumer.

## API

| Function | Description |
|----------|-------------|
| `Parse(text, sourcePath)` | Parse .bkn/.bknd/.md content into BknDocument |
| `ParseFrontmatter(text)` | Parse YAML frontmatter only |
| `ParseBody(text)` | Parse Markdown body into Object/Relation/Action/Risk/Connection lists |
| `ParseDataTables(text, fm, sourcePath)` | Parse .bknd data tables |
| `Load(path)` | Load single file from disk (.bkn/.bknd/.md) |
| `LoadNetwork(rootPath)` | Load network with includes resolution (.bkn/.bknd/.md) |
| `ValidateDataTable(table, schema, network)` | Validate DataTable against object schema |
| `ValidateNetworkData(network)` | Validate all DataTables in network |
| `ToBknd(opts)` | Serialize rows to .bknd format |
| `ToBkndFromTable(table, network, source)` | Serialize DataTable to .bknd |
| `EvaluateRisk(network, actionID, context, riskRules)` | Return RiskResult (Decision, RiskLevel, Reason) |
| `EvaluateRiskWith(evaluator, network, actionID, context, riskRules)` | Invoke custom evaluator, return RiskResult |
| `PlanDelete(network, targets, dryRun)` | Validate delete targets, return DeletePlan |
| `NetworkWithout(network, targets)` | Return new network with targets removed (in-memory) |
| `GenerateChecksumFile(root)` | Validate BKN inputs, then generate checksum.txt in directory |
| `VerifyChecksumFile(root)` | Verify checksum.txt, return (ok, errors) |

## Tests

Run from `sdk/golang`:

```bash
go test ./bkn/... -v
```
