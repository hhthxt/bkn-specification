#!/usr/bin/env python3
"""
演示：用「增强版安全契约矩阵」表格中的规则调用 SDK 风险评估。

数据来源：https://docs.google.com/spreadsheets/d/1zdJx6arbu_u7DiC7c9BatTD0Z1nDTkiRD1aDHg8S7Lw/edit?gid=1197712390

从仓库根目录运行：
  python examples/risk/scripts/eval_security_contract_demo.py
"""
from __future__ import annotations

import json
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
FRAGMENT_PATH = REPO_ROOT / "examples" / "risk" / "risk-fragment.bkn"
RULES_PATH = REPO_ROOT / "examples" / "risk" / "data" / "security_contract_rules.json"


def main() -> None:
    with open(RULES_PATH, encoding="utf-8") as f:
        data = json.load(f)
    risk_rules = data["risk_rules"]

    from bkn.loader import load_network
    from bkn.risk import evaluate_risk

    network = load_network(str(FRAGMENT_PATH))

    # 只保留 evaluate_risk 需要的键（scenario_id, action_id, allowed）；其余供展示
    rules_for_eval = [
        {"scenario_id": r["scenario_id"], "action_id": r["action_id"], "allowed": r["allowed"]}
        for r in risk_rules
    ]

    # 表格中的几条典型评估
    cases = [
        ("sec_t_01", "restart_pod", "月末封网期尝试重启 → 应 not_allow"),
        ("sec_t_01", "ddl_alter", "月末封网期尝试 DDL → 应 not_allow"),
        ("sec_r_01", "execute_sql", "核心产线高危 SQL → 应 not_allow"),
        ("sec_c_02", "batch_restart_nodes", "K8s 批量重启（智能灰度）→ 应 allow"),
        ("sec_n_01", "open_firewall_rule", "跨网段开白名单 → 应 not_allow"),
        ("sec_l_02", "high_freq_api_call", "API 高频调用（限流）→ 应 allow"),
    ]

    print("=== 增强版安全契约矩阵 - evaluate_risk 演示 ===\n")
    for scenario_id, action_id, desc in cases:
        context = {"scenario_id": scenario_id}
        result = evaluate_risk(network, action_id, context, risk_rules=rules_for_eval)
        print(f"  scenario={scenario_id} action={action_id}")
        print(f"    -> risk={result}  ({desc})")
        print()


if __name__ == "__main__":
    main()
