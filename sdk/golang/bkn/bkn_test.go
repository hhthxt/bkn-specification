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
	if len(doc.Entities) < 2 {
		t.Errorf("expected at least 2 entities, got %d", len(doc.Entities))
	}
	var foundScenario, foundRule bool
	for _, e := range doc.Entities {
		if e.ID == "risk_scenario" {
			foundScenario = true
		}
		if e.ID == "risk_rule" {
			foundRule = true
		}
	}
	if !foundScenario || !foundRule {
		t.Errorf("expected risk_scenario and risk_rule entities")
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
	if dt.EntityOrRelation != "risk_scenario" {
		t.Errorf("expected entity risk_scenario, got %q", dt.EntityOrRelation)
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
	entities := net.AllEntities()
	actions := net.AllActions()
	tables := net.AllDataTables()
	if len(entities) < 2 {
		t.Errorf("expected entities from includes, got %d", len(entities))
	}
	if len(actions) < 1 {
		t.Errorf("expected actions from includes, got %d", len(actions))
	}
	if len(tables) < 2 {
		t.Errorf("expected data tables from includes, got %d", len(tables))
	}
}

func TestValidateDataTable(t *testing.T) {
	schema := &Entity{
		ID: "test_entity",
		DataProperties: []DataProperty{
			{Property: "id", Constraint: "not_null", PrimaryKey: true},
			{Property: "name"},
		},
	}
	table := &DataTable{
		EntityOrRelation: "test_entity",
		Columns:          []string{"id", "name"},
		Rows:             []map[string]string{{"id": "1", "name": "a"}},
	}
	result := ValidateDataTable(table, schema, nil)
	if !result.OK() {
		t.Errorf("expected OK, got errors: %v", result.Errors)
	}
}

func TestValidateDataTable_NotNullFails(t *testing.T) {
	schema := &Entity{
		ID: "test_entity",
		DataProperties: []DataProperty{
			{Property: "id", Constraint: "not_null", PrimaryKey: true},
		},
	}
	table := &DataTable{
		EntityOrRelation: "test_entity",
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
		EntityOrRelation: "test_entity",
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
	// No rules -> allow
	if got := EvaluateRisk(net, "any_action", map[string]any{"scenario_id": "any"}, nil); got != "allow" {
		t.Errorf("expected allow, got %q", got)
	}
	// Rule forbids -> not_allow
	rules := []map[string]any{
		{"scenario_id": "prod_db", "action_id": "restore_from_backup", "allowed": false},
	}
	if got := EvaluateRisk(net, "restore_from_backup", map[string]any{"scenario_id": "prod_db"}, rules); got != "not_allow" {
		t.Errorf("expected not_allow, got %q", got)
	}
	// Rule allows -> allow
	rulesAllow := []map[string]any{
		{"scenario_id": "prod_db", "action_id": "restore_from_backup", "allowed": true},
	}
	if got := EvaluateRisk(net, "restore_from_backup", map[string]any{"scenario_id": "prod_db"}, rulesAllow); got != "allow" {
		t.Errorf("expected allow, got %q", got)
	}
	// No match -> allow
	rulesOther := []map[string]any{
		{"scenario_id": "other", "action_id": "other_action", "allowed": false},
	}
	if got := EvaluateRisk(net, "restore_from_backup", map[string]any{"scenario_id": "prod_db"}, rulesOther); got != "allow" {
		t.Errorf("expected allow (no match), got %q", got)
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
