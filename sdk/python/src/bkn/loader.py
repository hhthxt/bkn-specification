"""Load .bkn files from disk, resolving network includes."""

from __future__ import annotations

import os
from pathlib import Path

from bkn.models import BknDocument, BknNetwork
from bkn.parser import parse


def load(path: str | Path) -> BknDocument:
    """Load and parse a single .bkn file.

    Args:
        path: Path to the .bkn file.

    Returns:
        Parsed BknDocument.
    """
    path = Path(path)
    text = path.read_text(encoding="utf-8")
    return parse(text, source_path=str(path))


def load_network(root_path: str | Path) -> BknNetwork:
    """Load a network .bkn file and recursively resolve its includes.

    Args:
        root_path: Path to the root .bkn file (type: network).

    Returns:
        BknNetwork containing the root document and all included documents.

    Raises:
        ValueError: If a circular include is detected.
    """
    root_path = Path(root_path).resolve()
    root_doc = load(root_path)

    loaded_paths: set[str] = {str(root_path)}
    includes: list[BknDocument] = []

    _resolve_includes(root_doc, root_path.parent, loaded_paths, includes)

    return BknNetwork(root=root_doc, includes=includes)


def _resolve_includes(
    doc: BknDocument,
    base_dir: Path,
    loaded_paths: set[str],
    result: list[BknDocument],
) -> None:
    """Recursively resolve includes from a document's frontmatter."""
    for include_rel in doc.frontmatter.includes:
        include_path = (base_dir / include_rel).resolve()
        path_str = str(include_path)

        if path_str in loaded_paths:
            raise ValueError(
                f"Circular include detected: {include_rel} "
                f"(resolved to {path_str})"
            )

        if not include_path.exists():
            raise FileNotFoundError(
                f"Include file not found: {include_rel} "
                f"(resolved to {path_str})"
            )

        loaded_paths.add(path_str)
        inc_doc = load(include_path)
        result.append(inc_doc)

        _resolve_includes(inc_doc, include_path.parent, loaded_paths, result)
