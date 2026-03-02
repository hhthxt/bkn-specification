"""Risk assessment module: compute Action risk (allow/not_allow/unknown) from __risk__-tagged definitions and context."""

from bkn.risk.evaluate import ALLOW, NOT_ALLOW, UNKNOWN, RiskEvaluator, evaluate_risk

__all__ = ["evaluate_risk", "RiskEvaluator", "ALLOW", "NOT_ALLOW", "UNKNOWN"]
