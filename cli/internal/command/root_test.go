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
