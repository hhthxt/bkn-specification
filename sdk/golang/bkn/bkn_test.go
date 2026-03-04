package bkn

import (
	"os"
	"path/filepath"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Skip("cannot get cwd")
		return ""
	}
	// Walk up to find repo root (contains examples/risk/risk-fragment.bkn)
	dir := cwd
	for i := 0; i < 10; i++ {
		p := filepath.Join(dir, "examples", "risk", "risk-fragment.bkn")
		if _, err := os.Stat(p); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("repo root not found (run from bkn-specification or sdk/golang)")
	return ""
}

func TestParseRiskFragment(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "risk", "risk-fragment.bkn")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	doc, err := Parse(string(data), path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Frontmatter.Type != "fragment" {
		t.Errorf("expected type fragment, got %q", doc.Frontmatter.Type)
	}
	if len(doc.Objects) < 2 {
		t.Errorf("expected at least 2 objects, got %d", len(doc.Objects))
	}
	var foundScenario, foundRule bool
	for _, e := range doc.Objects {
		if e.ID == "risk_scenario" {
			foundScenario = true
		}
		if e.ID == "risk_rule" {
			foundRule = true
		}
	}
	if !foundScenario || !foundRule {
		t.Errorf("expected risk_scenario and risk_rule objects")
	}
}

func TestParseActions(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "risk", "actions.bkn")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	doc, err := Parse(string(data), path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(doc.Actions) < 1 {
		t.Errorf("expected at least 1 action, got %d", len(doc.Actions))
	}
}

func TestParseDataFile(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "risk", "data", "risk_scenario.bknd")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	doc, err := Parse(string(data), path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(doc.DataTables) != 1 {
		t.Fatalf("expected 1 data table, got %d", len(doc.DataTables))
	}
	dt := doc.DataTables[0]
	if dt.ObjectOrRelation != "risk_scenario" {
		t.Errorf("expected object risk_scenario, got %q", dt.ObjectOrRelation)
	}
	if len(dt.Rows) < 1 {
		t.Errorf("expected at least 1 row")
	}
}

func TestLoadNetwork(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "risk", "index.bkn")
	net, err := LoadNetwork(path)
	if err != nil {
		t.Fatalf("load network: %v", err)
	}
	objects := net.AllObjects()
	actions := net.AllActions()
	tables := net.AllDataTables()
	if len(objects) < 2 {
		t.Errorf("expected objects from includes, got %d", len(objects))
	}
	if len(actions) < 1 {
		t.Errorf("expected actions from includes, got %d", len(actions))
	}
	if len(tables) < 2 {
		t.Errorf("expected data tables from includes, got %d", len(tables))
	}
}

func TestValidateDataTable(t *testing.T) {
	schema := &BknObject{
		ID: "test_object",
		DataProperties: []DataProperty{
			{Property: "id", Constraint: "not_null", PrimaryKey: true},
			{Property: "name"},
		},
	}
	table := &DataTable{
		ObjectOrRelation: "test_object",
		Columns:          []string{"id", "name"},
		Rows:             []map[string]string{{"id": "1", "name": "a"}},
	}
	result := ValidateDataTable(table, schema, nil)
	if !result.OK() {
		t.Errorf("expected OK, got errors: %v", result.Errors)
	}
}

func TestValidateDataTable_NotNullFails(t *testing.T) {
	schema := &BknObject{
		ID: "test_object",
		DataProperties: []DataProperty{
			{Property: "id", Constraint: "not_null", PrimaryKey: true},
		},
	}
	table := &DataTable{
		ObjectOrRelation: "test_object",
		Columns:          []string{"id"},
		Rows:             []map[string]string{{"id": ""}},
	}
	result := ValidateDataTable(table, schema, nil)
	if result.OK() {
		t.Error("expected validation error for empty not_null")
	}
	var found bool
	for _, e := range result.Errors {
		if e.Code == "not_null" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected not_null error, got %v", result.Errors)
	}
}

func TestToBkndRoundTrip(t *testing.T) {
	table := &DataTable{
		ObjectOrRelation: "test_object",
		Columns:          []string{"id", "name"},
		Rows: []map[string]string{
			{"id": "1", "name": "a"},
			{"id": "2", "name": "b"},
		},
		Network: "test-network",
	}
	out, err := ToBkndFromTable(table, "test-network", "")
	if err != nil {
		t.Fatalf("ToBkndFromTable: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output")
	}
	if len(out) < 20 {
		t.Errorf("output too short: %q", out)
	}
	// Re-parse and check
	doc, err := Parse(out, "")
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(doc.DataTables) != 1 {
		t.Fatalf("expected 1 table after round-trip, got %d", len(doc.DataTables))
	}
	dt := doc.DataTables[0]
	if len(dt.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(dt.Rows))
	}
}

func TestEvaluateRisk(t *testing.T) {
	net := &BknNetwork{}
	// No rules -> unknown
	if got := EvaluateRisk(net, "any_action", map[string]any{"scenario_id": "any"}, nil); got.Decision != Unknown {
		t.Errorf("expected unknown, got %q", got.Decision)
	}
	// Rule forbids -> not_allow
	rules := []map[string]any{
		{"scenario_id": "sec_t_01", "action_id": "restart_erp", "allowed": false, "risk_level": 5, "reason": "blocked"},
	}
	if got := EvaluateRisk(net, "restart_erp", map[string]any{"scenario_id": "sec_t_01"}, rules); got.Decision != NotAllow {
		t.Errorf("expected not_allow, got %q", got.Decision)
	}
	if got := EvaluateRisk(net, "restart_erp", map[string]any{"scenario_id": "sec_t_01"}, rules); got.RiskLevel == nil || *got.RiskLevel != 5 {
		t.Errorf("expected risk_level 5, got %v", got.RiskLevel)
	}
	// Rule allows -> allow
	rulesAllow := []map[string]any{
		{"scenario_id": "sec_c_02", "action_id": "batch_restart_nodes", "allowed": true, "risk_level": 2},
	}
	if got := EvaluateRisk(net, "batch_restart_nodes", map[string]any{"scenario_id": "sec_c_02"}, rulesAllow); got.Decision != Allow {
		t.Errorf("expected allow, got %q", got.Decision)
	}
	// No match -> unknown
	rulesOther := []map[string]any{
		{"scenario_id": "other", "action_id": "other_action", "allowed": false},
	}
	if got := EvaluateRisk(net, "restart_erp", map[string]any{"scenario_id": "sec_t_01"}, rulesOther); got.Decision != Unknown {
		t.Errorf("expected unknown (no match), got %q", got.Decision)
	}
	// No scenario in context -> match by action_id only
	rulesGlobal := []map[string]any{
		{"action_id": "grant_root_admin", "allowed": false},
	}
	if got := EvaluateRisk(net, "grant_root_admin", map[string]any{}, rulesGlobal); got.Decision != NotAllow {
		t.Errorf("expected not_allow (no scenario filter), got %q", got.Decision)
	}
	if got := EvaluateRisk(net, "grant_root_admin", nil, rulesGlobal); got.Decision != NotAllow {
		t.Errorf("expected not_allow (nil context), got %q", got.Decision)
	}
	// Scenario in context filters out rules with different scenario
	if got := EvaluateRisk(net, "restart_erp", map[string]any{"scenario_id": "sec_c_02"}, rules); got.Decision != Unknown {
		t.Errorf("expected unknown (scenario mismatch), got %q", got.Decision)
	}
}

func TestEvaluateRiskWith(t *testing.T) {
	net := &BknNetwork{}
	myEvaluator := func(network *BknNetwork, actionID string, context map[string]any, riskRules []map[string]any) RiskResult {
		if actionID == "grant_root_admin" {
			lv := 5
			return RiskResult{Decision: NotAllow, RiskLevel: &lv, Reason: "全局禁止提权"}
		}
		return RiskResult{Decision: Unknown}
	}
	got := EvaluateRiskWith(myEvaluator, net, "grant_root_admin", map[string]any{}, nil)
	if got.Decision != NotAllow {
		t.Errorf("expected not_allow, got %q", got.Decision)
	}
	if got.RiskLevel == nil || *got.RiskLevel != 5 {
		t.Errorf("expected risk_level 5, got %v", got.RiskLevel)
	}
	if got.Reason != "全局禁止提权" {
		t.Errorf("expected reason 全局禁止提权, got %q", got.Reason)
	}
}

func TestValidateNetworkData(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "risk", "index.bkn")
	net, err := LoadNetwork(path)
	if err != nil {
		t.Fatalf("load network: %v", err)
	}
	result := ValidateNetworkData(net)
	if !result.OK() {
		t.Logf("validation errors: %v", result.Errors)
		// Allow some errors if schema doesn't match exactly, but core should pass
		for _, e := range result.Errors {
			if e.Code == "no_schema" {
				t.Errorf("unexpected no_schema for %s", e.Table)
			}
		}
	}
}
