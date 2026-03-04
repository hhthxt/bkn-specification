package bkn

import "fmt"

const (
	Allow    = "allow"
	NotAllow = "not_allow"
	Unknown  = "unknown"
)

// RiskResult is the structured output of risk evaluation.
type RiskResult struct {
	Decision  string // "allow", "not_allow", or "unknown"
	RiskLevel *int   // nil when unknown; 0–5 recommended
	Reason    string
}

// RiskEvaluatorFunc is the signature for custom risk evaluators.
type RiskEvaluatorFunc func(
	network *BknNetwork,
	actionID string,
	context map[string]any,
	riskRules []map[string]any,
) RiskResult

// EvaluateRisk computes whether the given action is allowed in the current context.
//
// Three-state result:
//   - "not_allow": at least one matching rule has allowed=false.
//   - "allow": at least one matching rule exists and all have allowed=true.
//   - "unknown": no riskRules provided or no rule matches the action+scenario.
func EvaluateRisk(network *BknNetwork, actionID string, context map[string]any, riskRules []map[string]any) RiskResult {
	if len(riskRules) == 0 {
		return RiskResult{Decision: Unknown}
	}
	var scenarioID string
	if context != nil {
		if v, ok := context["scenario_id"]; ok {
			scenarioID = fmtString(v)
		}
	}
	var matchedRules []map[string]any
	for _, rule := range riskRules {
		if getString(rule, "action_id") != actionID {
			continue
		}
		if scenarioID != "" && getString(rule, "scenario_id") != scenarioID {
			continue
		}
		matchedRules = append(matchedRules, rule)
	}
	if len(matchedRules) == 0 {
		return RiskResult{Decision: Unknown}
	}
	// not_allow priority: take first rule with allowed=false
	for _, rule := range matchedRules {
		if isAllowedFalse(rule) {
			return RiskResult{
				Decision:  NotAllow,
				RiskLevel: getIntPtr(rule, "risk_level"),
				Reason:    getString(rule, "reason"),
			}
		}
	}
	// all allow: take highest risk_level
	best := matchedRules[0]
	bestLevel := getIntPtr(best, "risk_level")
	bestVal := 0
	if bestLevel != nil {
		bestVal = *bestLevel
	}
	for _, rule := range matchedRules[1:] {
		lv := getIntPtr(rule, "risk_level")
		v := 0
		if lv != nil {
			v = *lv
		}
		if v > bestVal {
			bestLevel = lv
			bestVal = v
			best = rule
		}
	}
	return RiskResult{
		Decision:  Allow,
		RiskLevel: bestLevel,
		Reason:    getString(best, "reason"),
	}
}

// EvaluateRiskWith invokes the given evaluator for custom risk logic.
func EvaluateRiskWith(
	evaluator RiskEvaluatorFunc,
	network *BknNetwork,
	actionID string,
	context map[string]any,
	riskRules []map[string]any,
) RiskResult {
	return evaluator(network, actionID, context, riskRules)
}

func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		return fmtString(v)
	}
	return ""
}

func fmtString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	default:
		return fmt.Sprintf("%v", v)
	}
}

func isAllowedFalse(rule map[string]any) bool {
	v, ok := rule["allowed"]
	if !ok {
		return false
	}
	switch x := v.(type) {
	case bool:
		return !x
	default:
		return false
	}
}

func getIntPtr(m map[string]any, key string) *int {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch x := v.(type) {
	case int:
		return &x
	case int64:
		i := int(x)
		return &i
	case float64:
		i := int(x)
		return &i
	case string:
		var i int
		if _, err := fmt.Sscanf(x, "%d", &i); err == nil {
			return &i
		}
		return nil
	default:
		return nil
	}
}
