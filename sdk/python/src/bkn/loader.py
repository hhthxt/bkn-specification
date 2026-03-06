"""Load .bkn/.bknd/.md files from disk, resolving network includes."""

from __future__ import annotations

from pathlib import Path

from bkn.models import BknDocument, BknNetwork
from bkn.parser import parse

# Supported file extensions for BKN content. .md is allowed as a carrier;
# content must still satisfy BKN frontmatter/type/structure requirements.
BKN_SUPPORTED_EXTENSIONS = frozenset({".bkn", ".bknd", ".md"})


def _check_extension(path: Path) -> None:
    """Raise ValueError if path extension is not supported."""
    ext = path.suffix.lower()
    if ext not in BKN_SUPPORTED_EXTENSIONS:
        raise ValueError(
            f"Unsupported file extension: {ext!r}. "
            f"BKN supports: {', '.join(sorted(BKN_SUPPORTED_EXTENSIONS))}"
        )


def load(path: str | Path) -> BknDocument:
    """Load and parse a single .bkn/.bknd/.md file.

    Supported extensions: .bkn, .bknd, .md. Content must satisfy BKN
    frontmatter, type, and structure requirements regardless of extension.

    Args:
        path: Path to the .bkn, .bknd, or .md file.

    Returns:
        Parsed BknDocument.

    Raises:
        ValueError: If extension is unsupported or content is not valid BKN.
    """
    path = Path(path).resolve()
    _check_extension(path)
    text = path.read_text(encoding="utf-8")
    return parse(text, source_path=str(path))


def load_network(root_path: str | Path) -> BknNetwork:
    """Load a network file and recursively resolve its includes.

    Supported extensions: .bkn, .bknd, .md. Root file should be type: network.
    Only files listed in frontmatter `includes` are loaded. .bknd/.md data
    files may be loaded by explicitly including them.

    Args:
        root_path: Path to the root file (e.g. index.bkn or index.md).

    Returns:
        BknNetwork containing the root document and all included documents.

    Raises:
        ValueError: If extension is unsupported, content is not valid BKN,
            or a circular include is detected.
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
