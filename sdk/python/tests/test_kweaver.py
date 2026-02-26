"""Tests for the kweaver transformer."""

from __future__ import annotations

import json
import sys
import tempfile
from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[3]
EXAMPLES_DIR = REPO_ROOT / "docs" / "bkn_docs" / "examples"

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from bkn.loader import load_network
from bkn.transformers.kweaver import KweaverTransformer, _parse_index_config, _map_type


# ---------------------------------------------------------------------------
# Index Config parsing
# ---------------------------------------------------------------------------

class TestIndexConfigParsing:
    def test_empty(self):
        cfg = _parse_index_config("")
        assert cfg["keyword_config"]["enabled"] is False
        assert cfg["fulltext_config"]["enabled"] is False
        assert cfg["vector_config"]["enabled"] is False

    def test_keyword_only(self):
        cfg = _parse_index_config("keyword")
        assert cfg["keyword_config"]["enabled"] is True
        assert cfg["keyword_config"]["ignore_above_len"] == 1024

    def test_keyword_with_len(self):
        cfg = _parse_index_config("keyword(2048)")
        assert cfg["keyword_config"]["enabled"] is True
        assert cfg["keyword_config"]["ignore_above_len"] == 2048

    def test_fulltext_with_analyzer(self):
        cfg = _parse_index_config("fulltext(standard)")
        assert cfg["fulltext_config"]["enabled"] is True
        assert cfg["fulltext_config"]["analyzer"] == "standard"

    def test_vector_with_model(self):
        cfg = _parse_index_config("vector(1951511856216674304)")
        assert cfg["vector_config"]["enabled"] is True
        assert cfg["vector_config"]["model_id"] == "1951511856216674304"

    def test_combined(self):
        cfg = _parse_index_config(
            "keyword(1024) + fulltext(standard) + vector(1951511856216674304)"
        )
        assert cfg["keyword_config"]["enabled"] is True
        assert cfg["fulltext_config"]["enabled"] is True
        assert cfg["fulltext_config"]["analyzer"] == "standard"
        assert cfg["vector_config"]["enabled"] is True
        assert cfg["vector_config"]["model_id"] == "1951511856216674304"

    def test_fulltext_plus_vector(self):
        cfg = _parse_index_config("fulltext(standard) + vector(1951511856216674304)")
        assert cfg["keyword_config"]["enabled"] is False
        assert cfg["fulltext_config"]["enabled"] is True
        assert cfg["vector_config"]["enabled"] is True


# ---------------------------------------------------------------------------
# Type mapping
# ---------------------------------------------------------------------------

class TestTypeMapping:
    def test_varchar(self):
        assert _map_type("VARCHAR") == "string"

    def test_int64(self):
        assert _map_type("int64") == "integer"

    def test_float64(self):
        assert _map_type("float64") == "float"

    def test_decimal(self):
        assert _map_type("decimal") == "decimal"

    def test_decimal_with_precision(self):
        assert _map_type("decimal(10,2)") == "decimal"

    def test_date(self):
        assert _map_type("DATE") == "date"

    def test_timestamp(self):
        assert _map_type("TIMESTAMP") == "timestamp"

    def test_empty(self):
        assert _map_type("") == "string"


# ---------------------------------------------------------------------------
# Supplychain network transformation
# ---------------------------------------------------------------------------

class TestSupplychainTransform:
    """Test transforming the supplychain-hd network to kweaver JSON."""

    @pytest.fixture
    def network(self):
        root_path = EXAMPLES_DIR / "supplychain-hd" / "supplychain.bkn"
        if not root_path.exists():
            pytest.skip(f"Example file not found: {root_path}")
        return load_network(root_path)

    @pytest.fixture
    def transformer(self):
        return KweaverTransformer(
            branch="main",
            base_version="",
            id_prefix="supplychain_hd0202_",
        )

    @pytest.fixture
    def payload(self, network, transformer):
        return transformer.to_json(network)

    def test_knowledge_network(self, payload):
        kn = payload["knowledge_network"]
        assert kn["name"] == "HD供应链业务知识网络_v2"
        assert kn["branch"] == "main"
        assert "供应链" in kn.get("tags", [])

    def test_object_types_count(self, payload):
        assert len(payload["object_types"]) == 12

    def test_relation_types_count(self, payload):
        assert len(payload["relation_types"]) == 14

    def test_po_object_type(self, payload):
        po = next(
            (ot for ot in payload["object_types"] if ot.get("id", "").endswith("_po")),
            None,
        )
        assert po is not None
        assert po["name"] == "采购订单"
        assert "primary_keys" in po
        assert len(po["primary_keys"]) >= 1
        assert po["display_key"] != ""
        assert len(po["data_properties"]) > 0

    def test_po_data_property_types(self, payload):
        po = next(ot for ot in payload["object_types"] if ot.get("id", "").endswith("_po"))
        type_set = {dp["type"] for dp in po["data_properties"]}
        assert "string" in type_set

    def test_relation_has_source_target(self, payload):
        for rt in payload["relation_types"]:
            assert "source_object_type_id" in rt
            assert "target_object_type_id" in rt

    def test_relation_mapping_rules(self, payload):
        product2bom = next(
            (rt for rt in payload["relation_types"] if "product2bom" in rt.get("id", "")),
            None,
        )
        assert product2bom is not None
        assert len(product2bom.get("mapping_rules", [])) >= 1
        rule = product2bom["mapping_rules"][0]
        assert "source_property" in rule
        assert "name" in rule["source_property"]

    def test_id_prefix_applied(self, payload):
        for ot in payload["object_types"]:
            assert ot["id"].startswith("supplychain_hd0202_")

    def test_to_files(self, network, transformer):
        with tempfile.TemporaryDirectory() as tmpdir:
            files = transformer.to_files(network, tmpdir)
            assert len(files) == 3
            for f in files:
                assert f.exists()
                data = json.loads(f.read_text(encoding="utf-8"))
                assert data is not None


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
