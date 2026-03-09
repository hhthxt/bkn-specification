package command

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := cwd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "examples", "risk", "index.bkn")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("repo root not found")
	return ""
}

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if stderr.Len() > 0 {
		return strings.TrimSpace(stderr.String()), err
	}
	return strings.TrimSpace(stdout.String()), err
}

func TestInspectNetwork(t *testing.T) {
	root := repoRoot(t)
	out, err := executeCommand(t, "inspect", "network", filepath.Join(root, "examples", "risk", "index.bkn"))
	if err != nil {
		t.Fatalf("inspect network: %v (%s)", err, out)
	}
	if !strings.Contains(out, "Objects:") {
		t.Fatalf("expected object count in output, got %q", out)
	}
}

func TestInspectNetworkVerbose(t *testing.T) {
	root := repoRoot(t)
	out, err := executeCommand(
		t,
		"inspect", "network",
		filepath.Join(root, "examples", "risk", "index.bkn"),
		"--verbose",
	)
	if err != nil {
		t.Fatalf("inspect network verbose: %v (%s)", err, out)
	}
	if !strings.Contains(out, "Description: 本网络聚合风险相关定义") {
		t.Fatalf("expected network description in output, got %q", out)
	}
	if !strings.Contains(out, "Objects:") || !strings.Contains(out, "risk_scenario") {
		t.Fatalf("expected detailed object list in output, got %q", out)
	}
}

func TestInspectFileVerbose(t *testing.T) {
	root := repoRoot(t)
	out, err := executeCommand(
		t,
		"inspect", "file",
		filepath.Join(root, "examples", "risk", "actions.bkn"),
		"--verbose",
	)
	if err != nil {
		t.Fatalf("inspect file verbose: %v (%s)", err, out)
	}
	if !strings.Contains(out, "Description: 本片段定义与风险规则关联的动作") {
		t.Fatalf("expected file description in output, got %q", out)
	}
	if !strings.Contains(out, "Actions:") || !strings.Contains(out, "restart_erp") {
		t.Fatalf("expected detailed action list in output, got %q", out)
	}
}

func TestValidateNetwork(t *testing.T) {
	root := repoRoot(t)
	out, err := executeCommand(t, "validate", "network", filepath.Join(root, "examples", "risk", "index.bkn"))
	if err != nil {
		t.Fatalf("validate network: %v (%s)", err, out)
	}
	if !strings.Contains(out, "Validation OK") {
		t.Fatalf("expected validation success, got %q", out)
	}
}

func TestValidateNetwork_DirInput(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "examples", "k8s-network")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("examples/k8s-network not found")
		return
	}
	out, err := executeCommand(t, "validate", "network", dir)
	if err != nil {
		t.Fatalf("validate network dir: %v (%s)", err, out)
	}
	if !strings.Contains(out, "Validation OK") {
		t.Fatalf("expected validation success for dir input, got %q", out)
	}
}

func TestInspectNetwork_DirInput(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "examples", "k8s-network")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("examples/k8s-network not found")
		return
	}
	out, err := executeCommand(t, "inspect", "network", dir)
	if err != nil {
		t.Fatalf("inspect network dir: %v (%s)", err, out)
	}
	if !strings.Contains(out, "Objects:") || !strings.Contains(out, "k8s-network") {
		t.Fatalf("expected network inspection output, got %q", out)
	}
}

func TestValidateTable_NetworkDir(t *testing.T) {
	root := repoRoot(t)
	dataFile := filepath.Join(root, "examples", "risk", "data", "risk_scenario.bknd")
	networkDir := filepath.Join(root, "examples", "risk")
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		t.Skip("examples/risk/data/risk_scenario.bknd not found")
		return
	}
	out, err := executeCommand(t, "validate", "table", dataFile, "--network", networkDir)
	if err != nil {
		t.Fatalf("validate table --network dir: %v (%s)", err, out)
	}
	if !strings.Contains(out, "Validation OK") {
		t.Fatalf("expected validation success, got %q", out)
	}
}

func TestRiskEvalJSON(t *testing.T) {
	root := repoRoot(t)
	rulesPath := filepath.Join(t.TempDir(), "rules.json")
	if err := os.WriteFile(rulesPath, []byte(`[
  {"scenario_id":"sec_t_01","action_id":"restart_erp","allowed":false,"risk_level":5,"reason":"blocked"}
]`), 0644); err != nil {
		t.Fatalf("write rules: %v", err)
	}
	out, err := executeCommand(
		t,
		"risk", "eval",
		"--network", filepath.Join(root, "examples", "risk", "index.bkn"),
		"--action", "restart_erp",
		"--context", "scenario_id=sec_t_01",
		"--rules", rulesPath,
		"--format", "json",
	)
	if err != nil {
		t.Fatalf("risk eval: %v (%s)", err, out)
	}
	var payload struct {
		Result struct {
			Decision string `json:"decision"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal output: %v (%s)", err, out)
	}
	if payload.Result.Decision != "not_allow" {
		t.Fatalf("expected not_allow, got %q", payload.Result.Decision)
	}
}

func TestDeletePlanJSON(t *testing.T) {
	root := repoRoot(t)
	out, err := executeCommand(
		t,
		"delete", "plan",
		"--network", filepath.Join(root, "examples", "k8s-modular", "index.bkn"),
		"--target", "object:pod",
		"--format", "json",
	)
	if err != nil {
		t.Fatalf("delete plan: %v (%s)", err, out)
	}
	var payload struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal output: %v (%s)", err, out)
	}
	if !payload.OK {
		t.Fatal("expected delete plan ok=true")
	}
}

func TestDeleteSimulateJSON(t *testing.T) {
	root := repoRoot(t)
	out, err := executeCommand(
		t,
		"delete", "simulate",
		"--network", filepath.Join(root, "examples", "k8s-modular", "index.bkn"),
		"--target", "object:pod",
		"--format", "json",
	)
	if err != nil {
		t.Fatalf("delete simulate: %v (%s)", err, out)
	}
	var payload struct {
		After map[string]int `json:"after"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal output: %v (%s)", err, out)
	}
	if payload.After["objects"] != 2 {
		t.Fatalf("expected 2 objects after simulation, got %d", payload.After["objects"])
	}
}

func TestChecksumGenerateAndVerify(t *testing.T) {
	dir := t.TempDir()
	objectsDir := filepath.Join(dir, "objects")
	if err := os.Mkdir(objectsDir, 0755); err != nil {
		t.Fatalf("mkdir objects: %v", err)
	}
	if err := os.WriteFile(filepath.Join(objectsDir, "pod.bkn"), []byte(`---
type: object
id: pod
name: Pod
network: demo
---

## Object: pod

**Pod**
`), 0644); err != nil {
		t.Fatalf("write pod.bkn: %v", err)
	}
	if _, err := executeCommand(t, "checksum", "generate", dir); err != nil {
		t.Fatalf("checksum generate: %v", err)
	}
	out, err := executeCommand(t, "checksum", "verify", dir)
	if err != nil {
		t.Fatalf("checksum verify: %v (%s)", err, out)
	}
	if !strings.Contains(out, "Checksum OK") {
		t.Fatalf("expected checksum ok, got %q", out)
	}
}

func TestChecksumGenerateFailsOnInvalidNetwork(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "objects"), 0755); err != nil {
		t.Fatalf("mkdir objects: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "connections"), 0755); err != nil {
		t.Fatalf("mkdir connections: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "data"), 0755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.bkn"), []byte(`---
type: network
id: demo
name: Demo
includes:
  - connections/erp.bkn
  - objects/pod.bkn
  - data/pod.bknd
---

# Demo
`), 0644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "objects", "pod.bkn"), []byte(`---
type: object
id: pod
network: demo
---

## Object: pod

### Data Source

| Type | ID | Name |
|------|----|------|
| connection | erp | ERP |

### Data Properties

| Property | Primary Key |
|----------|-------------|
| id | YES |
`), 0644); err != nil {
		t.Fatalf("write object: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "connections", "erp.bkn"), []byte(`---
type: connection
id: erp
network: demo
---

## Connection: erp

**ERP**
`), 0644); err != nil {
		t.Fatalf("write connection: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "pod.bknd"), []byte(`---
type: data
object: pod
---

## Data

| id |
|----|
| pod-1 |
`), 0644); err != nil {
		t.Fatalf("write data: %v", err)
	}

	out, err := executeCommand(t, "checksum", "generate", dir)
	if err == nil {
		t.Fatalf("expected checksum generate to fail, output=%q", out)
	}
	if !strings.Contains(err.Error(), "checksum validation failed") {
		t.Fatalf("expected validation failure, got err=%q output=%q", err.Error(), out)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "checksum.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("expected checksum.txt not to be written, got %v", statErr)
	}
}

func TestDataToBkndJSON(t *testing.T) {
	rowsPath := filepath.Join(t.TempDir(), "rows.json")
	if err := os.WriteFile(rowsPath, []byte(`[{"id":"1","name":"pod-a"}]`), 0644); err != nil {
		t.Fatalf("write rows: %v", err)
	}
	out, err := executeCommand(
		t,
		"data", "to-bknd",
		"--object", "pod",
		"--network", "demo",
		"--in", rowsPath,
		"--format", "json",
	)
	if err != nil {
		t.Fatalf("data to-bknd: %v (%s)", err, out)
	}
	var payload struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal output: %v (%s)", err, out)
	}
	if !strings.Contains(payload.Content, "type: data") || !strings.Contains(payload.Content, "object: pod") {
		t.Fatalf("unexpected bknd content: %q", payload.Content)
	}
}
