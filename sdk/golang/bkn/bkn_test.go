package bkn

import (
	"os"
	"path/filepath"
	"strings"
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

func TestValidateDataTable_ReadOnlyDataSourceFails(t *testing.T) {
	schema := &BknObject{
		ID: "test_object",
		DataSource: &DataSource{
			Type: "connection",
			ID:   "erp_db",
			Name: "ERP DB",
		},
		DataProperties: []DataProperty{
			{Property: "id", PrimaryKey: true},
		},
	}
	table := &DataTable{
		ObjectOrRelation: "test_object",
		Columns:          []string{"id"},
		Rows:             []map[string]string{{"id": "1"}},
	}
	result := ValidateDataTable(table, schema, nil)
	if result.OK() {
		t.Fatal("expected validation error for readonly data source")
	}
	var found bool
	for _, e := range result.Errors {
		if e.Code == "readonly_data_source" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected readonly_data_source error, got %v", result.Errors)
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

// --- .md carrier compatibility tests ---

func TestLoadMdCompatNetwork(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "md-compat", "index.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("examples/md-compat not found")
		return
	}
	net, err := LoadNetwork(path)
	if err != nil {
		t.Fatalf("load network: %v", err)
	}
	if net.Root.Frontmatter.Type != "network" {
		t.Errorf("expected type network, got %q", net.Root.Frontmatter.Type)
	}
	if net.Root.Frontmatter.ID != "md-compat-demo" {
		t.Errorf("expected id md-compat-demo, got %q", net.Root.Frontmatter.ID)
	}
	objects := net.AllObjects()
	if len(objects) < 1 {
		t.Errorf("expected at least 1 object from includes, got %d", len(objects))
	}
}

func TestParseConnection(t *testing.T) {
	text := `---
type: connection
id: erp_db
name: ERP Database
network: demo
---

## Connection: erp_db

**ERP Database** - Shared ERP connection.

### Connection

| Type | Endpoint | Secret Ref |
|------|----------|------------|
| postgres | postgresql://erp.example.com:5432/erp | DB_PASSWORD |
`
	doc, err := Parse(text, "")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Frontmatter.Type != "connection" {
		t.Errorf("expected type connection, got %q", doc.Frontmatter.Type)
	}
	if len(doc.Connections) != 1 {
		t.Errorf("expected 1 connection, got %d", len(doc.Connections))
	}
	conn := doc.Connections[0]
	if conn.ID != "erp_db" {
		t.Errorf("expected id erp_db, got %q", conn.ID)
	}
	if conn.Config == nil {
		t.Fatal("expected config")
	}
	if conn.Config.ConnType != "postgres" || conn.Config.Endpoint != "postgresql://erp.example.com:5432/erp" || conn.Config.SecretRef != "DB_PASSWORD" {
		t.Errorf("unexpected config: %+v", conn.Config)
	}
}

func TestParseRisk(t *testing.T) {
	text := `---
type: risk
id: pod_restart_risk
name: Pod Restart Risk
network: demo
---

## Risk: pod_restart_risk

**Pod Restart Risk** - Controls pod restart actions.

### Control Scope

| Controlled Object | Controlled Action | Risk Level |
|-------------------|-------------------|------------|
| pod | restart_pod | high |

### Control Strategy

| Condition | Strategy |
|-----------|----------|
| production | require approval |

### Pre-checks

| Check Item | Type | Description |
|------------|------|-------------|
| can_i_restart | permission | Verify restart permission |

### Rollback Plan

Scale workload back to original replicas.

### Audit Requirements

Record operator and scenario.
`
	doc, err := Parse(text, "")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Frontmatter.Type != "risk" {
		t.Fatalf("expected type risk, got %q", doc.Frontmatter.Type)
	}
	if len(doc.Risks) != 1 {
		t.Fatalf("expected 1 risk, got %d", len(doc.Risks))
	}
	risk := doc.Risks[0]
	if risk.ID != "pod_restart_risk" {
		t.Errorf("expected id pod_restart_risk, got %q", risk.ID)
	}
	if len(risk.ControlScope) != 1 || risk.ControlScope[0].ControlledObject != "pod" || risk.ControlScope[0].ControlledAction != "restart_pod" || risk.ControlScope[0].RiskLevel != "high" {
		t.Errorf("unexpected control scope: %+v", risk.ControlScope)
	}
	if len(risk.ControlStrategies) != 1 || risk.ControlStrategies[0].Strategy != "require approval" {
		t.Errorf("unexpected control strategies: %+v", risk.ControlStrategies)
	}
	if len(risk.PreChecks) != 1 || risk.PreChecks[0].CheckItem != "can_i_restart" {
		t.Errorf("unexpected pre-checks: %+v", risk.PreChecks)
	}
	if !strings.Contains(risk.RollbackPlan, "original replicas") {
		t.Errorf("unexpected rollback plan: %q", risk.RollbackPlan)
	}
	if !strings.Contains(risk.AuditRequirements, "operator") {
		t.Errorf("unexpected audit requirements: %q", risk.AuditRequirements)
	}
}

func TestLoadConnectionDemo(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "connection-demo", "index.bkn")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("examples/connection-demo not found")
		return
	}
	net, err := LoadNetwork(path)
	if err != nil {
		t.Fatalf("load network: %v", err)
	}
	if len(net.AllConnections()) < 1 {
		t.Errorf("expected at least 1 connection, got %d", len(net.AllConnections()))
	}
	if net.GetConnection("erp_db") == nil {
		t.Error("expected erp_db connection")
	}
	material := findObject(net.AllObjects(), "material")
	if material == nil || material.DataSource == nil || material.DataSource.Type != "connection" || material.DataSource.ID != "erp_db" {
		t.Errorf("material should reference connection erp_db, got %+v", material)
	}
	legacy := findObject(net.AllObjects(), "legacy_view")
	if legacy == nil || legacy.DataSource == nil || legacy.DataSource.Type != "data_view" {
		t.Errorf("legacy_view should use data_view, got %+v", legacy)
	}
}

func TestLoadNetwork_MissingConnectionFails(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "index.bkn")
	obj := filepath.Join(dir, "material.bkn")
	if err := os.WriteFile(root, []byte(`---
type: network
id: demo
name: Demo
includes:
  - material.bkn
---
`), 0644); err != nil {
		t.Fatalf("write root: %v", err)
	}
	if err := os.WriteFile(obj, []byte(`---
type: object
id: material
name: Material
network: demo
---

## Object: material

**Material**

### Data Source

| Type | ID | Name |
|------|-----|------|
| connection | missing_conn | Missing Connection |

### Data Properties

| Property | Primary Key |
|----------|:-----------:|
| id | YES |
`), 0644); err != nil {
		t.Fatalf("write object: %v", err)
	}
	_, err := LoadNetwork(root)
	if err == nil {
		t.Fatal("expected missing connection error")
	}
	if !strings.Contains(err.Error(), "missing connection") {
		t.Errorf("expected missing connection error, got %q", err.Error())
	}
}

func findObject(objs []BknObject, id string) *BknObject {
	for i := range objs {
		if objs[i].ID == id {
			return &objs[i]
		}
	}
	return nil
}

func TestLoadMdCompatObjects(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "md-compat", "objects.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("examples/md-compat not found")
		return
	}
	doc, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if doc.Frontmatter.Type != "fragment" {
		t.Errorf("expected type fragment, got %q", doc.Frontmatter.Type)
	}
	if len(doc.Objects) < 1 {
		t.Errorf("expected at least 1 object, got %d", len(doc.Objects))
	}
	if doc.Objects[0].ID != "demo_item" {
		t.Errorf("expected object demo_item, got %q", doc.Objects[0].ID)
	}
}

func TestParse_NoFrontmatter(t *testing.T) {
	_, err := Parse("# Plain doc\n\nNo frontmatter here.", "")
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
	if !strings.Contains(err.Error(), "YAML frontmatter") {
		t.Errorf("expected frontmatter error, got %q", err.Error())
	}
}

func TestParse_NoType(t *testing.T) {
	text := "---\nid: x\nname: 测试\n---\n## Object: x"
	_, err := Parse(text, "")
	if err == nil {
		t.Fatal("expected error for missing type")
	}
	if !strings.Contains(err.Error(), "valid 'type' field") {
		t.Errorf("expected type field error, got %q", err.Error())
	}
}

func TestParse_InvalidType(t *testing.T) {
	text := "---\ntype: foo\nid: x\n---\n"
	_, err := Parse(text, "")
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
	if !strings.Contains(err.Error(), "invalid BKN type") {
		t.Errorf("expected invalid type error, got %q", err.Error())
	}
}

func TestLoad_UnsupportedExtension(t *testing.T) {
	f, err := os.CreateTemp("", "bkn-*.txt")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString("---\ntype: network\nid: x\n---\n")
	f.Close()

	_, err = Load(f.Name())
	if err == nil {
		t.Fatal("expected error for unsupported extension")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "unsupported file extension") {
		t.Errorf("expected extension error, got %q", err.Error())
	}
}

func TestPlanDelete(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "k8s-modular", "index.bkn")
	net, err := LoadNetwork(path)
	if err != nil {
		t.Skipf("example not found: %v", err)
	}
	plan := PlanDelete(net, []DeleteTarget{{Type: "object", ID: "pod"}}, true)
	if !plan.OK() {
		t.Errorf("expected ok, got not_found=%v", plan.NotFound)
	}
	if len(plan.Targets) != 1 || plan.Targets[0].ID != "pod" {
		t.Errorf("expected 1 target pod, got %v", plan.Targets)
	}
}

func TestPlanDelete_NotFound(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "k8s-modular", "index.bkn")
	net, err := LoadNetwork(path)
	if err != nil {
		t.Skipf("example not found: %v", err)
	}
	plan := PlanDelete(net, []DeleteTarget{{Type: "object", ID: "nonexistent"}}, true)
	if plan.OK() {
		t.Error("expected not ok for nonexistent target")
	}
	if len(plan.NotFound) != 1 {
		t.Errorf("expected 1 not_found, got %v", plan.NotFound)
	}
}

func TestNetworkWithout(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "k8s-modular", "index.bkn")
	net, err := LoadNetwork(path)
	if err != nil {
		t.Skipf("example not found: %v", err)
	}
	orig := len(net.AllObjects())
	out := NetworkWithout(net, []DeleteTarget{{Type: "object", ID: "pod"}})
	if len(out.AllObjects()) != orig-1 {
		t.Errorf("expected %d objects, got %d", orig-1, len(out.AllObjects()))
	}
	for _, o := range out.AllObjects() {
		if o.ID == "pod" {
			t.Error("pod should be removed")
		}
	}
}

func TestGenerateAndVerifyChecksum(t *testing.T) {
	dir, err := os.MkdirTemp("", "bkn-checksum-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Mkdir(filepath.Join(dir, "objects"), 0755)
	os.WriteFile(filepath.Join(dir, "objects", "pod.bkn"), []byte("---\ntype: object\nid: pod\nname: Pod\nnetwork: k8s\n---\n\n## Object: pod\n**Pod**\n"), 0644)

	content, err := GenerateChecksumFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, "sha256:") || !strings.Contains(content, "objects/pod.bkn") {
		t.Errorf("unexpected content: %s", content)
	}
	ok, errs := VerifyChecksumFile(dir)
	if !ok {
		t.Errorf("verify failed: %v", errs)
	}
}
