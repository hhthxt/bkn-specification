#!/usr/bin/env python3
"""Generate new-model risk .bknd data files from security contract rules JSON."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

SCENARIO_COLUMNS = [
    "scenario_id",
    "name",
    "category",
    "primary_object",
    "description",
    "activation_rule",
]

RULE_COLUMNS = [
    "rule_id",
    "scenario_id",
    "action_id",
    "allowed",
    "reason",
]

CATEGORY_MAPPING: dict[str, str] = {
    "时空管控": "availability",
    "爆炸半径": "integrity",
    "高危并发": "availability",
    "成本风控": "dependency",
    "数据隐私": "security",
    "拓扑依赖": "dependency",
    "前置校验": "integrity",
    "零信任权限": "security",
    "供应商管控": "dependency",
    "工控冻结": "availability",
    "流量熔断": "performance",
}


def _escape_cell(value: str) -> str:
    return (
        value.replace("|", "\\|")
        .replace("\n", " ")
        .replace("\r", " ")
        .strip()
    )


def _write_bknd(
    out_path: Path,
    network: str,
    object_id: str,
    columns: list[str],
    rows: list[dict[str, str]],
    source: str,
) -> None:
    header = [
        "---",
        "type: data",
        f"object: {object_id}",
        f"network: {network}",
        f"source: {source}",
        "---",
        "",
        f"# {object_id}",
        "",
    ]
    table_header = "| " + " | ".join(columns) + " |"
    table_sep = "|" + "|".join("-" * (len(c) + 2) for c in columns) + "|"
    table_rows = [
        "| "
        + " | ".join(_escape_cell(r.get(col, "")) for col in columns)
        + " |"
        for r in rows
    ]
    content = "\n".join(header + [table_header, table_sep] + table_rows) + "\n"
    out_path.write_text(content, encoding="utf-8")


def _load_activation_rules(activation_json: Path | None) -> dict[str, str]:
    """Load scenario_id -> activation_rule from scenario_activation.json."""
    if not activation_json or not activation_json.exists():
        return {}
    payload = json.loads(activation_json.read_text(encoding="utf-8"))
    result: dict[str, str] = {}
    for s in payload.get("scenarios", []):
        sid = str(s.get("scenario_id", "")).strip()
        if not sid:
            continue
        # Prefer time_window_rule (structured), fallback to time_window (human-readable)
        rule = str(s.get("time_window_rule", "")).strip() or str(
            s.get("time_window", "")
        ).strip()
        if rule:
            result[sid] = rule
    return result


def _build_rows(
    input_json: Path,
    activation_json: Path | None = None,
) -> tuple[list[dict[str, str]], list[dict[str, str]]]:
    payload = json.loads(input_json.read_text(encoding="utf-8"))
    rules = payload.get("risk_rules", [])
    activation_rules = _load_activation_rules(activation_json)

    scenarios: dict[str, dict[str, str]] = {}
    rule_rows: list[dict[str, str]] = []

    for item in rules:
        scenario_id = str(item.get("scenario_id", "")).strip()
        action_id = str(item.get("action_id", "")).strip()
        if not scenario_id or not action_id:
            continue

        rule_name = str(item.get("rule_name", scenario_id)).strip() or scenario_id
        category_raw = str(item.get("category", "")).strip()
        category = CATEGORY_MAPPING.get(category_raw, "operator")
        primary_object = str(item.get("primary_object", "")).strip()
        trigger_condition = str(item.get("trigger_condition", "")).strip()
        control_action = str(item.get("control_action", "")).strip()
        auth_level = str(item.get("auth_level", "")).strip()

        if scenario_id not in scenarios:
            scenarios[scenario_id] = {
                "scenario_id": scenario_id,
                "name": rule_name,
                "category": category,
                "primary_object": primary_object,
                "description": trigger_condition,
                "activation_rule": activation_rules.get(scenario_id, ""),
            }

        allowed_raw = item.get("allowed", False)
        allowed = "true" if bool(allowed_raw) else "false"
        rule_id = str(item.get("rule_id", "")).strip()
        if not rule_id or "_" not in rule_id:
            rule_id = f"{scenario_id}_{action_id}"
        rule_rows.append(
            {
                "rule_id": rule_id,
                "scenario_id": scenario_id,
                "action_id": action_id,
                "allowed": allowed,
                "reason": (
                    f"{rule_name}; control_action={control_action}; "
                    f"auth_level={auth_level}; trigger={trigger_condition}"
                ),
            }
        )

    scenario_rows = [scenarios[k] for k in sorted(scenarios.keys())]
    return scenario_rows, rule_rows


def convert(
    input_json: Path,
    output_dir: Path,
    network: str,
    source: str,
    activation_json: Path | None = None,
) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)
    # Remove legacy old-model files to avoid mixed schemas under examples/risk/data.
    for legacy_name in (
        "scenario.bknd",
        "action_option.bknd",
        "risk.bknd",
        "risk_statement.bknd",
        "rs_under_scenario.bknd",
        "rs_about_action.bknd",
        "rs_asserts_risk.bknd",
    ):
        legacy_path = output_dir / legacy_name
        if legacy_path.exists():
            legacy_path.unlink()

    scenario_rows, rule_rows = _build_rows(input_json, activation_json)
    _write_bknd(
        out_path=output_dir / "risk_scenario.bknd",
        network=network,
        object_id="risk_scenario",
        columns=SCENARIO_COLUMNS,
        rows=scenario_rows,
        source=source,
    )
    _write_bknd(
        out_path=output_dir / "risk_rule.bknd",
        network=network,
        object_id="risk_rule",
        columns=RULE_COLUMNS,
        rows=rule_rows,
        source=source,
    )


def main() -> None:
    repo_root = Path(__file__).resolve().parents[3]
    default_input = repo_root / "examples" / "risk" / "data" / "security_contract_rules.json"
    default_output = repo_root / "examples" / "risk" / "data"
    default_activation = repo_root / "examples" / "risk" / "data" / "scenario_activation.json"

    parser = argparse.ArgumentParser(
        description="Extract risk_scenario/risk_rule .bknd files from rules JSON."
    )
    parser.add_argument("--input-json", type=Path, default=default_input)
    parser.add_argument("--output-dir", type=Path, default=default_output)
    parser.add_argument(
        "--activation-json",
        type=Path,
        default=default_activation,
        help="scenario_activation.json for activation_rule merge",
    )
    parser.add_argument("--network", default="recoverable-network")
    parser.add_argument(
        "--source",
        default="security_contract_rules.json",
        help="Data provenance value written to frontmatter source",
    )
    args = parser.parse_args()

    convert(
        input_json=args.input_json,
        output_dir=args.output_dir,
        network=args.network,
        source=args.source,
        activation_json=args.activation_json,
    )

    print(f"Generated risk_scenario/risk_rule .bknd files in {args.output_dir}")


if __name__ == "__main__":
    main()
