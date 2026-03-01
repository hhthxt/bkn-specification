# Risk Data (`.bknd` + JSON)

本目录同时包含两类数据：

1. **旧模型 `.bknd` 样例**：用于演示 `type: data` 数据文件格式。对应 schema 定义于 `examples/risk_old/risk-fragment.bkn`（scenario / action_option / risk_statement 等）。**不参与当前 `examples/risk/risk-fragment.bkn` 网络**（该网络使用 risk_scenario / risk_rule 新模型）。
2. **安全契约规则样例（`.json`）**：用于 `evaluate_risk(..., risk_rules=...)` 演示脚本。

## `.bknd` 文件格式

`.bknd` 为 BKN 数据文件，扩展名与 `.bkn`（schema）区分。格式：YAML frontmatter（`type: data`、`entity` 或 `relation`、`network`、可选 `source`）+ Markdown 表格。列名需与 schema 中 Data Properties 一致。

仅当 Entity 的 Data Source 为 `bknd` 时，才使用 `.bknd` 维护数据；Data Source 为 `data_view` 的实体数据来自外部系统，不可用 `.bknd` 编辑。

## 文件列表

| 文件 | 类型 | 对应 schema ID |
|------|------|----------------|
| `scenario.bknd` | entity data | `scenario` |
| `action_option.bknd` | entity data | `action_option` |
| `risk.bknd` | entity data | `risk` |
| `risk_statement.bknd` | entity data | `risk_statement` |
| `rs_under_scenario.bknd` | relation data | `rs_under_scenario` |
| `rs_about_action.bknd` | relation data | `rs_about_action` |
| `rs_asserts_risk.bknd` | relation data | `rs_asserts_risk` |

## 生成与序列化

从 CSV 转换生成 `.bknd`：

```bash
python examples/risk/scripts/extract_risk_data.py
```

SDK 提供 `to_bknd()` / `to_bknd_from_table()` 将结构化数据序列化回 `.bknd` Markdown，供 LLM 消费或持久化。

## JSON 文件

| 文件 | 说明 |
|------|------|
| `security_contract_rules.json` | 安全契约矩阵规则实例（传给 `risk_rules`） |
| `scenario_activation.json` | 场景生效条件（如时间窗口） |

