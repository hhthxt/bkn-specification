package bkn

import "fmt"

const (
	Allow    = "allow"
	NotAllow = "not_allow"
	Unknown  = "unknown"
)

// EvaluateRisk computes whether the given action is allowed in the current context.
//
// Three-state result:
//   - "not_allow": at least one matching rule has allowed=false.
//   - "allow": at least one matching rule exists and all have allowed=true.
//   - "unknown": no riskRules provided or no rule matches the action+scenario.
func EvaluateRisk(network *BknNetwork, actionID string, context map[string]any, riskRules []map[string]any) string {
	if len(riskRules) == 0 {
		return Unknown
	}
	var scenarioID string
	if context != nil {
		if v, ok := context["scenario_id"]; ok {
			scenarioID = fmtString(v)
		}
	}
	matched := false
	for _, rule := range riskRules {
		if getString(rule, "action_id") != actionID {
			continue
		}
		if scenarioID != "" && getString(rule, "scenario_id") != scenarioID {
			continue
		}
		matched = true
		if isAllowedFalse(rule) {
			return NotAllow
		}
	}
	if matched {
		return Allow
	}
	return Unknown
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
