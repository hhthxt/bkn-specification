#!/usr/bin/env python3
"""Convert legacy risk CSV tables into .bknd data files."""

from __future__ import annotations

import argparse
import csv
from pathlib import Path


ENTITY_TABLES: dict[str, str] = {
    "scenario": "scenario",
    "action_option": "action_option",
    "risk": "risk",
    "risk_statement": "risk_statement",
}

RELATION_TABLES: dict[str, str] = {
    "rs_under_scenario": "rs_under_scenario",
    "rs_about_action": "rs_about_action",
    "rs_asserts_risk": "rs_asserts_risk",
}


def _escape_cell(value: str) -> str:
    # Keep Markdown table shape stable when values contain pipe chars.
    return (
        value.replace("|", "\\|")
        .replace("\n", " ")
        .replace("\r", " ")
        .strip()
    )


def _read_csv(path: Path) -> tuple[list[str], list[dict[str, str]]]:
    with path.open("r", encoding="utf-8", newline="") as f:
        reader = csv.DictReader(f)
        fieldnames = list(reader.fieldnames or [])
        rows: list[dict[str, str]] = []
        for row in reader:
            rows.append({k: str(v or "") for k, v in row.items()})
    return fieldnames, rows


def _write_bknd(
    out_path: Path,
    network: str,
    table_id: str,
    is_relation: bool,
    columns: list[str],
    rows: list[dict[str, str]],
    source: str,
) -> None:
    role_line = (
        f"relation: {table_id}" if is_relation else f"entity: {table_id}"
    )
    header = [
        "---",
        "type: data",
        role_line,
        f"network: {network}",
        f"source: {source}",
        "---",
        "",
        f"# {table_id}",
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


def convert(input_dir: Path, output_dir: Path, network: str, source: str) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)

    for table_id in ENTITY_TABLES:
        csv_path = input_dir / f"{table_id}.csv"
        columns, rows = _read_csv(csv_path)
        _write_bknd(
            output_dir / f"{table_id}.bknd",
            network=network,
            table_id=table_id,
            is_relation=False,
            columns=columns,
            rows=rows,
            source=source,
        )

    for table_id in RELATION_TABLES:
        csv_path = input_dir / f"{table_id}.csv"
        columns, rows = _read_csv(csv_path)
        _write_bknd(
            output_dir / f"{table_id}.bknd",
            network=network,
            table_id=table_id,
            is_relation=True,
            columns=columns,
            rows=rows,
            source=source,
        )


def main() -> None:
    repo_root = Path(__file__).resolve().parents[3]
    default_input = repo_root / "examples" / "risk_old" / "data"
    default_output = repo_root / "examples" / "risk" / "data"

    parser = argparse.ArgumentParser(
        description="Convert legacy risk CSV tables to .bknd files."
    )
    parser.add_argument("--input-dir", type=Path, default=default_input)
    parser.add_argument("--output-dir", type=Path, default=default_output)
    parser.add_argument("--network", default="recoverable-network")
    parser.add_argument(
        "--source",
        default="legacy risk csv export",
        help="Data provenance value written to frontmatter source",
    )
    args = parser.parse_args()

    convert(
        input_dir=args.input_dir,
        output_dir=args.output_dir,
        network=args.network,
        source=args.source,
    )

    print(f"Converted CSV files from {args.input_dir} -> {args.output_dir}")


if __name__ == "__main__":
    main()
