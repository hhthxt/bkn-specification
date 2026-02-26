"""Tests for BKN parser using real example files."""

from __future__ import annotations

import os
import sys
from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[3]
EXAMPLES_DIR = REPO_ROOT / "docs" / "examples"

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from bkn.parser import parse, parse_frontmatter, parse_body
from bkn.loader import load, load_network
from bkn.models import BknDocument, BknNetwork


# ---------------------------------------------------------------------------
# supplychain-hd network (split by type)
# ---------------------------------------------------------------------------

class TestSupplychainNetwork:
    """Test parsing the supplychain-hd multi-file network."""

    @pytest.fixture
    def network(self) -> BknNetwork:
        root_path = EXAMPLES_DIR / "supplychain-hd" / "supplychain.bkn"
        if not root_path.exists():
            pytest.skip(f"Example file not found: {root_path}")
        return load_network(root_path)

    def test_root_frontmatter(self, network: BknNetwork):
        fm = network.root.frontmatter
        assert fm.type == "network"
        assert fm.id == "supplychain"
        assert fm.name == "HD供应链业务知识网络_v2"
        assert "供应链" in fm.tags
        assert len(fm.includes) == 2

    def test_entity_count(self, network: BknNetwork):
        assert len(network.all_entities) == 12

    def test_relation_count(self, network: BknNetwork):
        assert len(network.all_relations) == 14

    def test_entity_po_parsed(self, network: BknNetwork):
        po = next((e for e in network.all_entities if e.id == "po"), None)
        assert po is not None
        assert po.name == "采购订单"
        assert len(po.data_properties) > 0

        pk_props = [dp for dp in po.data_properties if dp.primary_key]
        assert len(pk_props) >= 1

        dk_props = [dp for dp in po.data_properties if dp.display_key]
        assert len(dk_props) >= 1

    def test_entity_po_data_source(self, network: BknNetwork):
        po = next(e for e in network.all_entities if e.id == "po")
        assert po.data_source is not None
        assert po.data_source.type == "data_view"
        assert po.data_source.name == "erp_purchase_order"

    def test_entity_po_property_overrides(self, network: BknNetwork):
        po = next(e for e in network.all_entities if e.id == "po")
        assert len(po.property_overrides) > 0
        org_override = next(
            (o for o in po.property_overrides if o.property == "org_name"), None
        )
        assert org_override is not None
        assert "fulltext" in org_override.index_config
        assert "vector" in org_override.index_config

    def test_entity_tags(self, network: BknNetwork):
        po = next(e for e in network.all_entities if e.id == "po")
        assert "已审核" in po.tags

    def test_entity_product_logic_properties(self, network: BknNetwork):
        product = next(
            (e for e in network.all_entities if e.id == "product"), None
        )
        if product is None:
            pytest.skip("product entity not found")
        assert len(product.logic_properties) > 0

    def test_relation_product2bom(self, network: BknNetwork):
        rel = next(
            (r for r in network.all_relations if r.id == "product2bom"), None
        )
        assert rel is not None
        assert rel.name == "产品关联产品bom"
        assert len(rel.endpoints) == 1
        ep = rel.endpoints[0]
        assert ep.source == "product"
        assert ep.target == "bom"
        assert ep.type == "direct"

    def test_relation_mapping_rules(self, network: BknNetwork):
        rel = next(r for r in network.all_relations if r.id == "product2bom")
        assert len(rel.mapping_rules) >= 1
        assert rel.mapping_rules[0].source_property == "material_number"


# ---------------------------------------------------------------------------
# k8s-topology single-file network
# ---------------------------------------------------------------------------

class TestK8sTopology:
    """Test parsing the k8s-topology single-file example."""

    @pytest.fixture
    def doc(self) -> BknDocument:
        path = EXAMPLES_DIR / "k8s-topology.bkn"
        if not path.exists():
            pytest.skip(f"Example file not found: {path}")
        return load(path)

    def test_frontmatter(self, doc: BknDocument):
        fm = doc.frontmatter
        assert fm.type == "network"
        assert fm.id == "k8s-topology"
        assert fm.version == "1.0.0"

    def test_has_entities(self, doc: BknDocument):
        assert len(doc.entities) >= 3
        entity_ids = {e.id for e in doc.entities}
        assert "pod" in entity_ids
        assert "node" in entity_ids
        assert "service" in entity_ids

    def test_has_relations(self, doc: BknDocument):
        assert len(doc.relations) >= 2

    def test_has_actions(self, doc: BknDocument):
        assert len(doc.actions) >= 1

    def test_pod_entity(self, doc: BknDocument):
        pod = next(e for e in doc.entities if e.id == "pod")
        assert "Pod" in pod.name
        assert len(pod.data_properties) > 0

    def test_action_parsed(self, doc: BknDocument):
        action = doc.actions[0]
        assert action.id
        assert action.bound_entity


# ---------------------------------------------------------------------------
# Parser unit tests
# ---------------------------------------------------------------------------

class TestParserUnits:
    """Unit tests for low-level parser functions."""

    def test_simple_frontmatter(self):
        text = "---\ntype: entity\nid: test\nname: Test\n---\n\n## Entity: test\n"
        fm = parse_frontmatter(text)
        assert fm.type == "entity"
        assert fm.id == "test"
        assert fm.name == "Test"

    def test_no_frontmatter(self):
        text = "# Just markdown\n\nSome content.\n"
        fm = parse_frontmatter(text)
        assert fm.type == ""

    def test_frontmatter_with_tags(self):
        text = "---\ntype: network\nid: n1\ntags: [a, b, c]\n---\n"
        fm = parse_frontmatter(text)
        assert fm.tags == ["a", "b", "c"]

    def test_frontmatter_with_includes(self):
        text = "---\ntype: network\nid: n1\nincludes:\n  - entities.bkn\n  - relations.bkn\n---\n"
        fm = parse_frontmatter(text)
        assert fm.includes == ["entities.bkn", "relations.bkn"]

    def test_parse_entity_block(self):
        text = """---
type: fragment
id: test
---

## Entity: my_entity

**My Entity** - A test entity

- **Tags**: tag1, tag2

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | 123 | my_view |

### Data Properties

| Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index |
|----------|--------------|------|------------|-------------|:-----------:|:-----------:|:-----:|
| id | ID | int64 | | Primary key | YES | | YES |
| name | Name | VARCHAR | not_null | Entity name | | YES | YES |
"""
        entities, relations, actions = parse_body(text)
        assert len(entities) == 1
        e = entities[0]
        assert e.id == "my_entity"
        assert e.name == "My Entity"
        assert e.description == "A test entity"
        assert e.tags == ["tag1", "tag2"]
        assert e.data_source is not None
        assert e.data_source.id == "123"
        assert len(e.data_properties) == 2
        assert e.data_properties[0].primary_key is True
        assert e.data_properties[1].display_key is True

    def test_parse_relation_block(self):
        text = """---
type: fragment
id: test
---

## Relation: a_to_b

**A relates to B** - Test relation

- **Tags**: tested

### Endpoints

| Source | Target | Type | Required | Min | Max |
|--------|--------|------|----------|-----|-----|
| entity_a | entity_b | direct | NO | 0 | - |

### Mapping Rules

| Source Property | Target Property |
|-----------------|-----------------|
| a_id | b_ref_id |
"""
        _, relations, _ = parse_body(text)
        assert len(relations) == 1
        r = relations[0]
        assert r.id == "a_to_b"
        assert r.name == "A relates to B"
        assert r.tags == ["tested"]
        assert len(r.endpoints) == 1
        assert r.endpoints[0].source == "entity_a"
        assert len(r.mapping_rules) == 1
        assert r.mapping_rules[0].source_property == "a_id"


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
