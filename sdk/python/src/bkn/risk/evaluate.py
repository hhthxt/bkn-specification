"""Evaluate whether an action is allowed in a given scenario based on risk-tagged knowledge."""

from __future__ import annotations

from typing import Any, Protocol

from bkn.models import BknNetwork


class RiskEvaluator(Protocol):
    """Protocol for a risk evaluation function. Users may implement this for custom logic."""

    def __call__(
        self,
        network: BknNetwork,
        action_id: str,
        context: dict[str, Any],
        **kwargs: Any,
    ) -> str:
        """Return 'allow' or 'not_allow'. Optional kwargs (e.g. risk_rules) are implementation-defined."""
        ...


def evaluate_risk(
    network: BknNetwork,
    action_id: str,
    context: dict[str, Any],
    risk_rules: list[dict[str, Any]] | None = None,
) -> str:
    """
    Compute whether the given action is allowed in the current context (e.g. scenario).

    Built-in implementation: uses instance data (risk_rules) only; definitions tagged
    with reserved `__risk__` in the network identify risk-related schema. Returns
    "not_allow" if any matching rule has allowed=False, otherwise "allow". When no
    risk_rules are provided, returns "allow" by default. Users may replace this with
    a custom evaluator (same signature) for their own risk logic.

    Args:
        network: Loaded BKN network (reserved for API consistency / future use).
        action_id: Action ID to evaluate.
        context: Context for the evaluation, e.g. {"scenario_id": "prod_db"}.
        risk_rules: Optional list of rule dicts with keys scenario_id, action_id, allowed.
                    Each allowed is bool; False means not_allow for that scenario+action.

    Returns:
        "allow" or "not_allow".
    """
    scenario_id = (context or {}).get("scenario_id")
    if risk_rules is None:
        risk_rules = []
    for rule in risk_rules:
        if rule.get("action_id") != action_id:
            continue
        if scenario_id is not None and rule.get("scenario_id") != scenario_id:
            continue
        if rule.get("allowed") is False:
            return "not_allow"
    return "allow"
