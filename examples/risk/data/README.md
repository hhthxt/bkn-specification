# Risk Data (`.bknd` + JSON)

本目录已切换到 `examples/risk/risk-fragment.bkn` 的新模型（`risk_scenario` / `risk_rule`）。

## `.bknd` 文件格式

`.bknd` 为 BKN 数据文件，扩展名与 `.bkn`（schema）区分。格式：YAML frontmatter（`type: data`、`entity` 或 `relation`、`network`、可选 `source`）+ Markdown 表格。列名需与 schema 中 Data Properties 一致。

仅当 Entity 的 Data Source 为 `bknd` 时，才使用 `.bknd` 维护数据；Data Source 为 `data_view` 的实体数据来自外部系统，不可用 `.bknd` 编辑。

## 文件列表（新模型）

| 文件 | 类型 | 对应 schema ID |
|------|------|----------------|
| `risk_scenario.bknd` | entity data | `risk_scenario` |
| `risk_rule.bknd` | entity data | `risk_rule` |

## 生成与序列化

从 `security_contract_rules.json` 生成新模型 `.bknd`：

```bash
python examples/risk/scripts/extract_risk_data.py
```

SDK 提供 `to_bknd()` / `to_bknd_from_table()` 将结构化数据序列化回 `.bknd` Markdown，供 LLM 消费或持久化。

## JSON 文件

| 文件 | 说明 |
|------|------|
| `security_contract_rules.json` | 安全契约矩阵规则实例（抽取为 `risk_scenario.bknd`/`risk_rule.bknd`，也可直接传给 `evaluate_risk(..., risk_rules=...)`） |
| `scenario_activation.json` | 场景生效条件（如时间窗口） |

旧模型（`scenario` / `action_option` / `risk` / `risk_statement`）保留在 `examples/risk_old/`。

