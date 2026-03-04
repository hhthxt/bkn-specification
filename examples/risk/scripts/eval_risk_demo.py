#!/usr/bin/env python3
"""
Demo: call SDK risk assessment to get RiskResult (decision, risk_level, reason) for an action in a given context.
Usage (from repo root):
  python examples/risk/scripts/eval_risk_demo.py
  python examples/risk/scripts/eval_risk_demo.py --action restart_erp --scenario sec_t_01
  python examples/risk/scripts/eval_risk_demo.py --custom
"""
from __future__ import annotations

import argparse
from pathlib import Path

# Network path: examples/risk/index.bkn (relative to repo root)
REPO_ROOT = Path(__file__).resolve().parents[3]
FRAGMENT_PATH = REPO_ROOT / "examples" / "risk" / "index.bkn"


def _rules_from_network(network) -> list[dict]:
    """Extract risk_rules from network's risk_rule data table."""
    for table in network.all_data_tables:
        if table.object_or_relation == "risk_rule":
            rules = []
            for row in table.rows:
                r = dict(row)
                # Convert allowed from "true"/"false" string to bool
                a = r.get("allowed", "")
                r["allowed"] = str(a).lower() in ("true", "1", "yes")
                # risk_level: keep as-is (may be str from bknd, _to_int handles it)
                rules.append(r)
            return rules
    return []


def main() -> None:
    parser = argparse.ArgumentParser(description="Evaluate risk for an action in a scenario")
    parser.add_argument("--action", default="restart_erp", help="Action ID")
    parser.add_argument("--scenario", default="sec_t_01", help="Scenario ID (context)")
    parser.add_argument("--bkn", type=Path, default=FRAGMENT_PATH, help="Path to BKN fragment/network")
    parser.add_argument("--custom", action="store_true", help="Use custom evaluator demo")
    args = parser.parse_args()

    from bkn.loader import load_network
    from bkn.risk import RiskResult, evaluate_risk

    network = load_network(str(args.bkn))
    context = {"scenario_id": args.scenario}

    if args.custom:
        # Custom evaluator demo
        def my_evaluator(network, action_id, context, risk_rules=None, **kwargs):
            if action_id == "grant_root_admin":
                return RiskResult(decision="not_allow", risk_level=5, reason="全局禁止提权")
            return RiskResult(decision="unknown")

        result = evaluate_risk(
            network, "grant_root_admin", {}, evaluator=my_evaluator
        )
        print("=== Custom evaluator demo (action=grant_root_admin) ===")
    else:
        # Built-in: load rules from network data
        risk_rules = _rules_from_network(network)
        result = evaluate_risk(network, args.action, context, risk_rules=risk_rules)
        print(f"=== Built-in evaluator: action={args.action} scenario={args.scenario} ===")

    print(f"  decision   = {result.decision}")
    print(f"  risk_level = {result.risk_level}")
    print(f"  reason    = {result.reason}")


if __name__ == "__main__":
    main()
