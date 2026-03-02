"""Evaluate whether an action is allowed in a given scenario based on risk-tagged knowledge."""

from __future__ import annotations

from typing import Any, Protocol

from bkn.models import BknNetwork


ALLOW = "allow"
NOT_ALLOW = "not_allow"
UNKNOWN = "unknown"


class RiskEvaluator(Protocol):
    """Protocol for a risk evaluation function. Users may implement this for custom logic."""

    def __call__(
        self,
        network: BknNetwork,
        action_id: str,
        context: dict[str, Any],
        **kwargs: Any,
    ) -> str:
        """Return 'allow', 'not_allow', or 'unknown'."""
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
    with reserved `__risk__` in the network identify risk-related schema.

    Three-state result:
      - "not_allow": at least one matching rule has allowed=False → block.
      - "allow": at least one matching rule exists and all have allowed=True → permit.
      - "unknown": no risk_rules provided or no rule matches → insufficient info.

    Users may replace this with a custom evaluator (same signature) for their own
    risk logic.

    Args:
        network: Loaded BKN network (reserved for API consistency / future use).
        action_id: Action ID to evaluate.
        context: Context for the evaluation, e.g. {"scenario_id": "sec_t_01"}.
        risk_rules: Optional list of rule dicts with keys scenario_id, action_id, allowed.

    Returns:
        "allow", "not_allow", or "unknown".
    """
    if not risk_rules:
        return UNKNOWN
    scenario_id = (context or {}).get("scenario_id")
    matched = False
    for rule in risk_rules:
        if rule.get("action_id") != action_id:
            continue
        if scenario_id is not None and rule.get("scenario_id") != scenario_id:
            continue
        matched = True
        if rule.get("allowed") is False:
            return NOT_ALLOW
    return ALLOW if matched else UNKNOWN
