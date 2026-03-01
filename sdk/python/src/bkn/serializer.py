"""Serialize structured data to .bknd Markdown format."""

from __future__ import annotations

from bkn.models import DataTable


def _escape_cell(val: str | int | float | None) -> str:
    """Escape a cell value for Markdown table (avoid breaking pipes)."""
    if val is None:
        return ""
    s = str(val).strip()
    if "|" in s:
        s = s.replace("|", "\\|")
    return s


def to_bknd(
    entity_id: str | None = None,
    relation_id: str | None = None,
    rows: list[dict[str, str | int | float | None]] | None = None,
    network: str = "",
    source: str | None = None,
    columns: list[str] | None = None,
) -> str:
    """Serialize structured data to .bknd Markdown format.

    Args:
        entity_id: Entity ID (use with relation_id=None for entity data).
        relation_id: Relation ID (use with entity_id=None for relation data).
        rows: Data rows as list of dicts. Keys must match column names.
        network: Network ID.
        source: Optional provenance (e.g. source file).
        columns: Column order. If None, inferred from first row keys.

    Returns:
        Full .bknd file content (frontmatter + Markdown table).
    """
    if entity_id and relation_id:
        raise ValueError("Specify either entity_id or relation_id, not both")
    if not entity_id and not relation_id:
        raise ValueError("Specify either entity_id or relation_id")
    if not rows:
        rows = []

    target_id = relation_id if relation_id else entity_id
    is_relation = bool(relation_id)

    if columns is None and rows:
        columns = list(rows[0].keys())
    elif columns is None:
        columns = []

    fm_lines = [
        "---",
        "type: data",
        f"network: {network}",
    ]
    if is_relation:
        fm_lines.append(f"relation: {relation_id}")
    else:
        fm_lines.append(f"entity: {entity_id}")
    if source:
        fm_lines.append(f"source: {source}")
    fm_lines.append("---")
    fm_lines.append("")
    fm_lines.append(f"# {target_id}")
    fm_lines.append("")

    if not columns:
        return "\n".join(fm_lines)

    header = "| " + " | ".join(columns) + " |"
    sep = "|" + "|".join(["---"] * len(columns)) + "|"
    table_lines = [header, sep]
    for row in rows:
        cells = [_escape_cell(row.get(c, "")) for c in columns]
        table_lines.append("| " + " | ".join(cells) + " |")

    fm_lines.append("\n".join(table_lines))
    return "\n".join(fm_lines)


def to_bknd_from_table(
    table: DataTable,
    network: str | None = None,
    source: str | None = None,
) -> str:
    """Serialize a DataTable to .bknd Markdown format.

    Args:
        table: Parsed DataTable from a .bknd file.
        network: Override network (uses table.network if not provided).
        source: Optional source override.

    Returns:
        Full .bknd file content.
    """
    net = network or table.network
    cols = table.columns if table.columns else (list(table.rows[0].keys()) if table.rows else [])
    if relation_id := (table.entity_or_relation if table.is_relation else None):
        return to_bknd(
            relation_id=relation_id,
            rows=table.rows,
            network=net,
            source=source,
            columns=cols,
        )
    return to_bknd(
        entity_id=table.entity_or_relation,
        rows=table.rows,
        network=net,
        source=source,
        columns=cols,
    )
