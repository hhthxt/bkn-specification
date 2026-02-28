"""Verify all examples/*.bkn files load successfully with the SDK."""

from __future__ import annotations

import pytest
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
EXAMPLES_DIR = REPO_ROOT / "examples"


def _get_all_bkn():
    """Collect all .bkn files under examples/."""
    return sorted(EXAMPLES_DIR.rglob("*.bkn"))


class TestLoadSingleFiles:
    """Every .bkn file must parse with load()."""

    @pytest.mark.parametrize("path", _get_all_bkn(), ids=lambda p: str(p.relative_to(EXAMPLES_DIR)))
    def test_load_single_file(self, path: Path):
        from bkn import load
        doc = load(path)
        assert doc is not None
        assert doc.frontmatter is not None


class TestLoadNetworks:
    """Network entry points must load with load_network()."""

    def test_supplychain_network(self):
        from bkn import load_network
        path = EXAMPLES_DIR / "supplychain-hd" / "supplychain.bkn"
        net = load_network(path)
        assert len(net.all_entities) == 12
        assert len(net.all_relations) == 14

    def test_k8s_topology_single_file(self):
        from bkn import load_network
        path = EXAMPLES_DIR / "k8s-topology.bkn"
        net = load_network(path)
        assert len(net.all_entities) >= 3
        assert len(net.all_relations) >= 2
        assert len(net.all_actions) >= 2

    def test_k8s_network_with_includes(self):
        from bkn import load_network
        path = EXAMPLES_DIR / "k8s-network" / "index.bkn"
        net = load_network(path)
        assert len(net.all_entities) >= 3
        assert len(net.all_relations) >= 2
        assert len(net.all_actions) >= 2

    def test_k8s_modular_with_includes(self):
        """k8s-modular index.bkn includes all entity/relation/action files."""
        from bkn import load_network
        path = EXAMPLES_DIR / "k8s-modular" / "index.bkn"
        net = load_network(path)
        assert len(net.all_entities) == 3
        assert len(net.all_relations) == 2
        assert len(net.all_actions) == 2
