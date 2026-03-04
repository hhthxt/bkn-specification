"""Tests for the risk assessment module."""

from __future__ import annotations

from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[3]
RISK_FRAGMENT = REPO_ROOT / "examples" / "risk" / "risk-fragment.bkn"


@pytest.fixture
def network():
    """Load the risk-tagged fragment as a network (single doc, no includes)."""
    from bkn.loader import load_network
    return load_network(str(RISK_FRAGMENT))


def test_evaluate_risk_no_rules_returns_unknown(network):
    """With no risk_rules, result is unknown (insufficient info)."""
    from bkn.risk import evaluate_risk
    assert evaluate_risk(network, "any_action", {"scenario_id": "any"}) == "unknown"
    assert evaluate_risk(network, "restart_erp", {}) == "unknown"


def test_evaluate_risk_not_allow_when_rule_forbids(network):
    """When a rule has allowed=False for the given scenario+action, return not_allow."""
    from bkn.risk import evaluate_risk
    rules = [
        {"scenario_id": "sec_t_01", "action_id": "restart_erp", "allowed": False},
    ]
    assert evaluate_risk(
        network, "restart_erp", {"scenario_id": "sec_t_01"}, risk_rules=rules
    ) == "not_allow"


def test_evaluate_risk_allow_when_rule_allows(network):
    """When a rule has allowed=True, return allow."""
    from bkn.risk import evaluate_risk
    rules = [
        {"scenario_id": "sec_c_02", "action_id": "batch_restart_nodes", "allowed": True},
    ]
    assert evaluate_risk(
        network, "batch_restart_nodes", {"scenario_id": "sec_c_02"}, risk_rules=rules
    ) == "allow"


def test_evaluate_risk_no_match_returns_unknown(network):
    """When no rule matches the scenario+action, return unknown."""
    from bkn.risk import evaluate_risk
    rules = [
        {"scenario_id": "other_scenario", "action_id": "other_action", "allowed": False},
    ]
    assert evaluate_risk(
        network, "restart_erp", {"scenario_id": "sec_t_01"}, risk_rules=rules
    ) == "unknown"


def test_evaluate_risk_no_scenario_matches_by_action_only(network):
    """Without scenario_id in context, rules match by action_id alone."""
    from bkn.risk import evaluate_risk
    rules = [
        {"action_id": "grant_root_admin", "allowed": False},
    ]
    assert evaluate_risk(network, "grant_root_admin", {}, risk_rules=rules) == "not_allow"
    assert evaluate_risk(network, "grant_root_admin", None, risk_rules=rules) == "not_allow"


def test_evaluate_risk_no_scenario_allow(network):
    """Global rule with allowed=True, no scenario filtering."""
    from bkn.risk import evaluate_risk
    rules = [
        {"action_id": "query_sensitive_data", "allowed": True},
    ]
    assert evaluate_risk(network, "query_sensitive_data", {}, risk_rules=rules) == "allow"


def test_evaluate_risk_scenario_filters_rules(network):
    """With scenario_id in context, only matching rules participate."""
    from bkn.risk import evaluate_risk
    rules = [
        {"scenario_id": "sec_t_01", "action_id": "restart_erp", "allowed": False},
    ]
    assert evaluate_risk(
        network, "restart_erp", {"scenario_id": "sec_c_02"}, risk_rules=rules
    ) == "unknown"


def test_network_has_risk_tagged_objects(network):
    """The risk fragment defines objects with reserved Tags: __risk__."""
    risk_objects = [e for e in network.all_objects if "__risk__" in (e.tags or [])]
    assert len(risk_objects) >= 1
    ids = [e.id for e in risk_objects]
    assert "risk_scenario" in ids or "risk_rule" in ids
