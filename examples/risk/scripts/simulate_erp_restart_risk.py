#!/usr/bin/env python3
"""
模拟：输入「时间 2026-02-28 23:00 + 动作 重启ERP」→ 风险判断模块 → not_allow。

对应 SEC-T-01 月末财务绝对封网：每月28日23点至31日4点禁止重启/缩容/DDL。
从仓库根目录运行：python examples/risk/scripts/simulate_erp_restart_risk.py
"""
from __future__ import annotations

import json
from datetime import datetime
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
FRAGMENT_PATH = REPO_ROOT / "examples" / "risk" / "risk-fragment.bkn"
RULES_PATH = REPO_ROOT / "examples" / "risk" / "data" / "security_contract_rules.json"

# 月末封网：当月 28 日 23:00 至 31 日 04:00（含 31 日 4 点前）
def time_to_scenario_id(dt: datetime) -> str | None:
    day = dt.day
    hour = dt.hour
    minute = dt.minute
    # 28 日 23:00 起
    if day == 28 and (hour > 23 or (hour == 23 and minute >= 0)):
        return "sec_t_01"
    if day == 29 or day == 30:
        return "sec_t_01"
    # 31 日 04:00 前
    if day == 31 and (hour < 4 or (hour == 4 and minute == 0)):
        return "sec_t_01"
    return None


ACTION_NAME_TO_ID = {
    "重启ERP": "restart_erp",
    "重启erp": "restart_erp",
}


def action_name_to_id(name: str) -> str | None:
    return ACTION_NAME_TO_ID.get(name.strip()) if name else None


def main() -> None:
    # 输入
    time_str = "2026-02-28 23:00"
    action_name = "重启ERP"
    dt = datetime.strptime(time_str, "%Y-%m-%d %H:%M")

    scenario_id = time_to_scenario_id(dt)
    action_id = action_name_to_id(action_name)
    if not scenario_id or not action_id:
        print(f"time={time_str} action={action_name!r} -> scenario_id={scenario_id!r} action_id={action_id!r}")
        print("Cannot resolve scenario or action; abort.")
        return

    with open(RULES_PATH, encoding="utf-8") as f:
        data = json.load(f)
    risk_rules_raw = data["risk_rules"]
    risk_rules = [
        {"scenario_id": r["scenario_id"], "action_id": r["action_id"], "allowed": r["allowed"]}
        for r in risk_rules_raw
    ]

    from bkn.loader import load_network
    from bkn.risk import evaluate_risk

    network = load_network(str(FRAGMENT_PATH))
    context = {"scenario_id": scenario_id}
    result = evaluate_risk(network, action_id, context, risk_rules=risk_rules)

    print("=== 模拟：2026-02-28 23:00 重启ERP 风险判断 ===")
    print(f"  输入: 时间={time_str}, 动作={action_name!r}")
    print(f"  解析: scenario_id={scenario_id}, action_id={action_id}")
    print(f"  结论: risk={result}")
    if result == "not_allow":
        print("  (符合 SEC-T-01 月末财务绝对封网，预期 not_allow)")
    else:
        print("  (预期 not_allow，请检查规则数据)")


if __name__ == "__main__":
    main()
