"""Risk assessment module: compute Action risk (allow/not_allow) from __risk__-tagged definitions and context."""

from bkn.risk.evaluate import RiskEvaluator, evaluate_risk

__all__ = ["evaluate_risk", "RiskEvaluator"]
