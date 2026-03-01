# Risk Design (Tag-Based)

本示例展示基于 **tag** 的风险相关定义与 Action 的 **动态 risk 属性**（allow / not_allow）。

## 设计要点

1. **内置 tag `__risk__`**：`__risk__` 为规范保留 tag，仅用于参与内置风险评估的实体/关系；用户不得使用。通过 `- **Tags**: __risk__` 标记；用户可自定义其他 tag 的风险类及自己的评估函数。
2. **Action 的 risk 属性**：Action 拥有运行时/计算属性 `risk`，取值仅 `allow` | `not_allow`，由 **SDK 风险评估模块** 根据当前场景与带 `risk` tag 的实体/关系数据计算得出，不写入 BKN 文件。
3. **风险评估模块**：SDK 提供 `bkn.risk.evaluate_risk(network, action_id, context)`，在给定场景（context）下判断某 action 是否允许执行。

## 目录结构

```
examples/risk/
├── README.md           # 本说明
├── risk.md             # 风险类设计与交互逻辑
├── index.bkn           # 网络入口（聚合 risk-fragment + actions）
├── risk-fragment.bkn   # 带 Tags: __risk__ 的实体示例（无冗余关系）
├── actions.bkn         # 动作定义（如 restart_erp 重启ERP）
├── data/
│   ├── *.bknd                         # risk_scenario / risk_rule 实例数据
│   ├── security_contract_rules.json   # 增强版安全契约矩阵（表格转 risk_rules）
│   └── scenario_activation.json       # 场景生效条件（如 sec_t_01 时间窗口）
└── scripts/
    ├── eval_risk_demo.py              # 通用 evaluate_risk 演示
    ├── eval_security_contract_demo.py # 安全契约矩阵规则演示
    └── simulate_erp_restart_risk.py   # 模拟：2026-02-28 23:00 重启ERP → not_allow
```

## 增强版安全契约矩阵示例

规则数据来源：[增强版安全契约矩阵](https://docs.google.com/spreadsheets/d/1zdJx6arbu_u7DiC7c9BatTD0Z1nDTkiRD1aDHg8S7Lw/edit?gid=1197712390)（策略类别、策略 ID、作用域、触发条件、风控动作、授权级别）。已转换为 `data/security_contract_rules.json`，供 `evaluate_risk(..., risk_rules=...)` 使用。

运行演示（需在仓库根目录）：

```bash
python examples/risk/scripts/eval_security_contract_demo.py
```

将按表格中的多条策略（如 SEC-T-01 月末封网、SEC-R-01 核心产线防危化、SEC-C-02 雪崩防波及等）输出对应 scenario + action 的 allow/not_allow 结果。

## 模拟：2026-02-28 23:00 重启ERP → not_allow

输入「时间 + 动作」，经时间→场景、动作名→action_id 解析后调用风险判断，得到 Action 结论。

- **输入**：时间 2026-02-28 23:00，动作 重启ERP。
- **业务知识**：`data/scenario_activation.json` 描述 sec_t_01（月末封网）生效时间窗口；`data/security_contract_rules.json` 含 (sec_t_01, restart_erp, allowed=false)；`actions.bkn` 定义 Action: restart_erp。
- **运行**（仓库根目录）：`python examples/risk/scripts/simulate_erp_restart_risk.py`
- **输出**：risk=not_allow（符合 SEC-T-01 月末财务绝对封网）。

## 使用 SDK 风险评估

```python
from bkn.loader import load_network
from bkn.risk import evaluate_risk

network = load_network("examples/risk/index.bkn")
result = evaluate_risk(network, action_id="restore_from_backup", context={"scenario_id": "prod_db"})
# result == "allow" or "not_allow"

# 传入规则实例（如从 security_contract_rules.json 加载）得到门控结果
# risk_rules = [{"scenario_id": "sec_t_01", "action_id": "restart_pod", "allowed": False}, ...]
# evaluate_risk(network, "restart_pod", {"scenario_id": "sec_t_01"}, risk_rules=risk_rules)  # -> "not_allow"
```

风险类的设计与交互逻辑见 [risk.md](risk.md)。
