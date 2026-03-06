"""Tests for BKN .md carrier compatibility."""

from __future__ import annotations

from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[3]
EXAMPLES_DIR = REPO_ROOT / "examples"
MD_COMPAT_DIR = EXAMPLES_DIR / "md-compat"


class TestMdCompatPositive:
    """Positive: .md files with valid BKN content load successfully."""

    def test_load_index_md_network(self):
        """index.md + includes objects.md loads."""
        if not (MD_COMPAT_DIR / "index.md").exists():
            pytest.skip("examples/md-compat not found")
        from bkn.loader import load_network

        network = load_network(MD_COMPAT_DIR / "index.md")
        assert network.root.frontmatter.type == "network"
        assert network.root.frontmatter.id == "md-compat-demo"
        assert len(network.all_objects) >= 1

    def test_load_objects_md_fragment(self):
        """objects.md as fragment loads."""
        if not (MD_COMPAT_DIR / "objects.md").exists():
            pytest.skip("examples/md-compat not found")
        from bkn.loader import load

        doc = load(MD_COMPAT_DIR / "objects.md")
        assert doc.frontmatter.type == "fragment"
        assert len(doc.objects) >= 1
        assert doc.objects[0].id == "demo_item"


class TestMdCompatNegative:
    """Negative: invalid content fails with clear errors."""

    def test_no_frontmatter_raises(self):
        """Plain .md without frontmatter raises."""
        from bkn.parser import parse

        with pytest.raises(ValueError, match="YAML frontmatter"):
            parse("# Plain doc\n\nNo frontmatter here.")

    def test_no_type_raises(self):
        """Frontmatter without type raises."""
        from bkn.parser import parse

        text = "---\nid: x\nname: 测试\n---\n## Object: x"
        with pytest.raises(ValueError, match="valid 'type' field"):
            parse(text)

    def test_invalid_type_raises(self):
        """Invalid type value raises."""
        from bkn.parser import parse

        text = "---\ntype: foo\nid: x\n---\n"
        with pytest.raises(ValueError, match="Invalid BKN type"):
            parse(text)

    def test_unsupported_extension_raises(self):
        """Unsupported file extension raises."""
        import tempfile

        from bkn.loader import load

        with tempfile.NamedTemporaryFile(suffix=".txt", delete=False) as f:
            f.write(b"---\ntype: network\nid: x\n---\n")
            p = Path(f.name)
        try:
            with pytest.raises(ValueError, match="Unsupported file extension"):
                load(p)
        finally:
            p.unlink(missing_ok=True)
