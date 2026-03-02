package bkn

import "fmt"

// EvaluateRisk computes whether the given action is allowed in the current context.
// Returns "not_allow" if any matching rule has allowed=false, otherwise "allow".
// When no riskRules are provided, returns "allow" by default.
func EvaluateRisk(network *BknNetwork, actionID string, context map[string]any, riskRules []map[string]any) string {
	var scenarioID string
	if context != nil {
		if v, ok := context["scenario_id"]; ok {
			scenarioID = fmtString(v)
		}
	}
	if riskRules == nil {
		riskRules = []map[string]any{}
	}
	for _, rule := range riskRules {
		if getString(rule, "action_id") != actionID {
			continue
		}
		if scenarioID != "" && getString(rule, "scenario_id") != scenarioID {
			continue
		}
		if isAllowedFalse(rule) {
			return "not_allow"
		}
	}
	return "allow"
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
