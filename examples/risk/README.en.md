# Risk Design (Tag-Based)

**中文**: [README.md](README.md)

This example demonstrates **tag-based** risk definitions and the **dynamic risk attribute** (allow / not_allow) of Actions.

## Design Highlights

1. **Built-in tag `__risk__`**: `__risk__` is a reserved tag for entities/relations participating in built-in risk assessment; users must not use it. Marked via `- **Tags**: __risk__`; users may define custom risk classes with other tags and their own evaluation functions.
2. **Action risk attribute**: Actions have a runtime/computed attribute `risk`, with values `allow` | `not_allow`, computed by the **SDK risk module** from the current scenario and risk-tagged entity/relation data; it is not stored in BKN files.
3. **Risk evaluation module**: The SDK provides `bkn.risk.evaluate_risk(network, action_id, context)` to determine whether an action is allowed in a given scenario (context).

## Directory Structure

```
examples/risk/
├── README.md           # This readme (Chinese)
├── README.en.md        # This readme (English)
├── risk.md             # Risk design and interaction logic
├── index.bkn           # Network entry (aggregates risk-fragment + actions)
├── risk-fragment.bkn   # Entity examples with Tags: __risk__
├── actions.bkn         # Action definitions (e.g. restart_erp)
├── data/
│   ├── *.bknd                         # risk_scenario / risk_rule instance data
│   ├── security_contract_rules.json   # Security contract matrix (rules for evaluate_risk)
│   └── scenario_activation.json       # Scenario activation (e.g. sec_t_01 time window)
└── scripts/
    ├── eval_risk_demo.py              # Generic evaluate_risk demo
    ├── eval_security_contract_demo.py # Security contract matrix demo
    └── simulate_erp_restart_risk.py   # Simulate: 2026-02-28 23:00 restart ERP → not_allow
```

## Security Contract Matrix Example

Rule data source: [Security Contract Matrix](https://docs.google.com/spreadsheets/d/1zdJx6arbu_u7DiC7c9BatTD0Z1nDTkiRD1aDHg8S7Lw/edit?gid=1197712390) (policy category, policy ID, scope, trigger conditions, control actions, auth level). Converted to `data/security_contract_rules.json` for use with `evaluate_risk(..., risk_rules=...)`.

Run the demo (from repo root):

```bash
python examples/risk/scripts/eval_security_contract_demo.py
```

Outputs allow/not_allow for each scenario + action per the matrix (e.g. SEC-T-01 month-end lockdown, SEC-R-01 core production hazard prevention, SEC-C-02 avalanche mitigation).

## Simulate: 2026-02-28 23:00 Restart ERP → not_allow

Input "time + action"; after resolving time→scenario and action name→action_id, the risk evaluation returns the Action result.

- **Input**: Time 2026-02-28 23:00, action "Restart ERP".
- **Business context**: `data/scenario_activation.json` describes sec_t_01 (month-end lockdown) time window; `data/security_contract_rules.json` contains (sec_t_01, restart_erp, allowed=false); `actions.bkn` defines Action: restart_erp.
- **Run** (from repo root): `python examples/risk/scripts/simulate_erp_restart_risk.py`
- **Output**: risk=not_allow (per SEC-T-01 month-end finance absolute lockdown).

## Using the SDK for Risk Evaluation

```python
from bkn.loader import load_network
from bkn.risk import evaluate_risk

network = load_network("examples/risk/index.bkn")

# No rules -> default allow
result = evaluate_risk(network, action_id="restart_erp", context={"scenario_id": "sec_t_01"})
# result == "allow"

# Pass rule instances (e.g. from security_contract_rules.json) for gating
risk_rules = [{"scenario_id": "sec_t_01", "action_id": "restart_erp", "allowed": False}]
evaluate_risk(network, "restart_erp", {"scenario_id": "sec_t_01"}, risk_rules=risk_rules)
# -> "not_allow"
```

See [risk.md](risk.md) for risk design and interaction logic.
