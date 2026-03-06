"""Verify all examples/*.bkn files load successfully with the SDK."""

from __future__ import annotations

import pytest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
EXAMPLES_DIR = REPO_ROOT / "examples"


def _get_all_bkn():
    """Collect all .bkn files under examples/."""
    return sorted(EXAMPLES_DIR.rglob("*.bkn"))


def _get_all_bknd():
    """Collect all .bknd files under examples/."""
    return sorted(EXAMPLES_DIR.rglob("*.bknd"))


class TestLoadSingleFiles:
    """Every .bkn file must parse with load()."""

    @pytest.mark.parametrize(
        "path",
        _get_all_bkn(),
        ids=lambda p: str(p.relative_to(EXAMPLES_DIR)),
    )
    def test_load_single_file(self, path: Path):
        from bkn import load
        doc = load(path)
        assert doc is not None
        assert doc.frontmatter is not None

    @pytest.mark.parametrize(
        "path",
        _get_all_bknd(),
        ids=lambda p: str(p.relative_to(EXAMPLES_DIR)),
    )
    def test_load_single_data_file(self, path: Path):
        from bkn import load
        doc = load(path)
        assert doc is not None
        assert doc.frontmatter.type == "data"
        assert len(doc.data_tables) >= 1


class TestLoadNetworks:
    """Network entry points must load with load_network()."""

    def test_supplychain_network(self):
        from bkn import load_network
        path = EXAMPLES_DIR / "supplychain-hd" / "supplychain.bkn"
        net = load_network(path)
        assert len(net.all_objects) == 12
        assert len(net.all_relations) == 14

    def test_k8s_topology_single_file(self):
        from bkn import load_network
        path = EXAMPLES_DIR / "k8s-topology.bkn"
        net = load_network(path)
        assert len(net.all_objects) >= 3
        assert len(net.all_relations) >= 2
        assert len(net.all_actions) >= 2

    def test_k8s_network_with_includes(self):
        from bkn import load_network
        path = EXAMPLES_DIR / "k8s-network" / "index.bkn"
        net = load_network(path)
        assert len(net.all_objects) >= 3
        assert len(net.all_relations) >= 2
        assert len(net.all_actions) >= 2

    def test_k8s_modular_with_includes(self):
        """k8s-modular index.bkn includes all object/relation/action files."""
        from bkn import load_network
        path = EXAMPLES_DIR / "k8s-modular" / "index.bkn"
        net = load_network(path)
        assert len(net.all_objects) == 3
        assert len(net.all_relations) == 2
        assert len(net.all_actions) == 2

    def test_risk_network_includes_bknd(self):
        """risk-fragment includes risk_scenario/risk_rule .bknd data files."""
        from bkn import load_network
        path = EXAMPLES_DIR / "risk" / "risk-fragment.bkn"
        net = load_network(path)
        assert len(net.all_data_tables) == 2
        ids = {t.object_or_relation for t in net.all_data_tables}
        assert ids == {"risk_scenario", "risk_rule"}

    def test_md_compat_network(self):
        """index.md + includes objects.md loads (BKN .md carrier compatibility)."""
        from bkn import load_network
        path = EXAMPLES_DIR / "md-compat" / "index.md"
        if not path.exists():
            pytest.skip("examples/md-compat not found")
        net = load_network(path)
        assert net.root.frontmatter.type == "network"
        assert net.root.frontmatter.id == "md-compat-demo"
        assert len(net.all_objects) >= 1
