// // Copyright The kweaver.ai Authors.
// //
// // Licensed under the Apache License, Version 2.0.
// // See the LICENSE file in the project root for details.

package bkn

// import (
// 	"os"
// 	"path/filepath"
// 	"strings"
// 	"testing"
// )

// // --- Unit Tests for Parse ---

// func TestParse_NoFrontmatter(t *testing.T) {
// 	_, err := Parse("# Plain doc\n\nNo frontmatter here.", "")
// 	if err == nil {
// 		t.Fatal("expected error for missing frontmatter")
// 	}
// 	if !strings.Contains(err.Error(), "YAML frontmatter") {
// 		t.Errorf("expected frontmatter error, got %q", err.Error())
// 	}
// }

// func TestParse_NoType(t *testing.T) {
// 	text := "---\nid: x\nname: 测试\n---\n## Object: x"
// 	_, err := Parse(text, "")
// 	if err == nil {
// 		t.Fatal("expected error for missing type")
// 	}
// 	if !strings.Contains(err.Error(), "valid 'type' field") {
// 		t.Errorf("expected type field error, got %q", err.Error())
// 	}
// }

// func TestParse_InvalidType(t *testing.T) {
// 	text := "---\ntype: foo\nid: x\n---\n"
// 	_, err := Parse(text, "")
// 	if err == nil {
// 		t.Fatal("expected error for invalid type")
// 	}
// 	if !strings.Contains(err.Error(), "invalid BKN type") {
// 		t.Errorf("expected invalid type error, got %q", err.Error())
// 	}
// }

// func TestParse_RiskTypeValid(t *testing.T) {
// 	text := `---
// type: risk_type
// id: test_risk
// name: Test Risk
// ---

// ## Risk: test_risk
// **Level**: high

// ### Pre-checks
// - Check 1
// - Check 2

// ### Post-actions
// - Action 1
// `
// 	doc, err := Parse(text, "")
// 	if err != nil {
// 		t.Fatalf("parse: %v", err)
// 	}
// 	if doc.Frontmatter.Type != "risk_type" {
// 		t.Errorf("expected type risk_type, got %q", doc.Frontmatter.Type)
// 	}
// 	if len(doc.Risks) != 1 {
// 		t.Errorf("expected 1 risk, got %d", len(doc.Risks))
// 	}
// }

// // --- Unit Tests for Load ---

// func TestLoad_UnsupportedExtension(t *testing.T) {
// 	f, err := os.CreateTemp("", "bkn-*.txt")
// 	if err != nil {
// 		t.Fatalf("create temp: %v", err)
// 	}
// 	defer os.Remove(f.Name())
// 	f.WriteString("---\ntype: network\nid: x\n---\n")
// 	f.Close()

// 	_, err = Load(f.Name())
// 	if err == nil {
// 		t.Fatal("expected error for unsupported extension")
// 	}
// 	if !strings.Contains(strings.ToLower(err.Error()), "unsupported file extension") {
// 		t.Errorf("expected extension error, got %q", err.Error())
// 	}
// }

// // --- Unit Tests for Delete Operations ---

// func TestPlanDelete_Unit(t *testing.T) {
// 	net := &BknNetwork{
// 		Root: BknDocument{
// 			Frontmatter: Frontmatter{
// 				Type: "network",
// 				ID:   "test-net",
// 				Name: "Test Network",
// 			},
// 		},
// 		Includes: []BknDocument{
// 			{
// 				Frontmatter: Frontmatter{Type: "object_type", ID: "pod", Name: "Pod"},
// 				Objects:     []BknObject{{ID: "pod", Name: "Pod"}},
// 			},
// 			{
// 				Frontmatter: Frontmatter{Type: "object_type", ID: "node", Name: "Node"},
// 				Objects:     []BknObject{{ID: "node", Name: "Node"}},
// 			},
// 		},
// 	}

// 	plan := PlanDelete(net, []DeleteTarget{{Type: "object", ID: "pod"}}, true)
// 	if !plan.OK() {
// 		t.Errorf("expected ok, got not_found=%v", plan.NotFound)
// 	}
// 	if len(plan.Targets) != 1 || plan.Targets[0].ID != "pod" {
// 		t.Errorf("expected 1 target pod, got %v", plan.Targets)
// 	}
// }

// func TestPlanDelete_NotFound_Unit(t *testing.T) {
// 	net := &BknNetwork{
// 		Root: BknDocument{
// 			Frontmatter: Frontmatter{Type: "network", ID: "test-net"},
// 		},
// 		Includes: []BknDocument{
// 			{
// 				Frontmatter: Frontmatter{Type: "object_type", ID: "pod"},
// 				Objects:     []BknObject{{ID: "pod"}},
// 			},
// 		},
// 	}

// 	plan := PlanDelete(net, []DeleteTarget{{Type: "object", ID: "nonexistent"}}, true)
// 	if plan.OK() {
// 		t.Error("expected not ok for nonexistent target")
// 	}
// 	if len(plan.NotFound) != 1 {
// 		t.Errorf("expected 1 not_found, got %v", plan.NotFound)
// 	}
// }

// func TestNetworkWithout_Unit(t *testing.T) {
// 	net := &BknNetwork{
// 		Root: BknDocument{
// 			Frontmatter: Frontmatter{Type: "network", ID: "test-net"},
// 		},
// 		Includes: []BknDocument{
// 			{
// 				Frontmatter: Frontmatter{Type: "object_type", ID: "pod"},
// 				Objects:     []BknObject{{ID: "pod", Name: "Pod"}},
// 			},
// 			{
// 				Frontmatter: Frontmatter{Type: "object_type", ID: "node"},
// 				Objects:     []BknObject{{ID: "node", Name: "Node"}},
// 			},
// 		},
// 	}

// 	orig := len(net.AllObjects())
// 	out := NetworkWithout(net, []DeleteTarget{{Type: "object", ID: "pod"}})
// 	if len(out.AllObjects()) != orig-1 {
// 		t.Errorf("expected %d objects, got %d", orig-1, len(out.AllObjects()))
// 	}
// 	for _, o := range out.AllObjects() {
// 		if o.ID == "pod" {
// 			t.Error("pod should be removed")
// 		}
// 	}
// }

// // --- Unit Tests for Checksum ---

// func TestGenerateAndVerifyChecksum(t *testing.T) {
// 	dir, err := os.MkdirTemp("", "bkn-checksum-*")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(dir)
// 	os.Mkdir(filepath.Join(dir, "object_types"), 0755)
// 	os.WriteFile(filepath.Join(dir, "object_types", "pod.bkn"), []byte("---\ntype: object_type\nid: pod\nname: Pod\nnetwork: k8s\n---\n\n## Object: pod\n**Pod**\n"), 0644)

// 	content, err := GenerateChecksumFile(dir)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if !strings.Contains(content, "sha256:") || !strings.Contains(content, "object_type:pod") {
// 		t.Errorf("unexpected content: %s", content)
// 	}
// 	ok, errs := VerifyChecksumFile(dir)
// 	if !ok {
// 		t.Errorf("verify failed: %v", errs)
// 	}
// }

// func TestChecksumNormalization_BlankLinesAndWhitespace(t *testing.T) {
// 	dir, err := os.MkdirTemp("", "bkn-checksum-norm-*")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(dir)

// 	baseBkn := "---\ntype: object_type\nid: x\n---\n\n## Object: x\n**X**\n"
// 	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(baseBkn), 0644)

// 	content, err := GenerateChecksumFile(dir)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	var baseHash string
// 	for _, line := range strings.Split(content, "\n") {
// 		if strings.Contains(line, "object_type:x") {
// 			parts := strings.SplitN(line, "  ", 2)
// 			if len(parts) == 2 {
// 				baseHash = strings.TrimSpace(parts[1])
// 				break
// 			}
// 		}
// 	}
// 	if baseHash == "" {
// 		t.Fatal("could not find object_type:x checksum")
// 	}

// 	withBlankLines := "---\ntype: object_type\nid: x\n---\n\n\n## Object: x\n\n**X**\n\n"
// 	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(withBlankLines), 0644)
// 	content2, _ := GenerateChecksumFile(dir)
// 	var hash2 string
// 	for _, line := range strings.Split(content2, "\n") {
// 		if strings.Contains(line, "object_type:x") {
// 			parts := strings.SplitN(line, "  ", 2)
// 			if len(parts) == 2 {
// 				hash2 = strings.TrimSpace(parts[1])
// 				break
// 			}
// 		}
// 	}
// 	if baseHash != hash2 {
// 		t.Errorf("checksum changed with blank lines: %q vs %q", baseHash, hash2)
// 	}
// }

// func TestChecksumNormalization_SemanticChangeAltersChecksum(t *testing.T) {
// 	dir, err := os.MkdirTemp("", "bkn-checksum-semantic-*")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(dir)

// 	baseBkn := "---\ntype: object_type\nid: x\n---\n\n## Object: x\n**X**\n"
// 	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(baseBkn), 0644)
// 	content, _ := GenerateChecksumFile(dir)
// 	var baseHash string
// 	for _, line := range strings.Split(content, "\n") {
// 		if strings.Contains(line, "object_type:x") {
// 			parts := strings.SplitN(line, "  ", 2)
// 			if len(parts) == 2 {
// 				baseHash = strings.TrimSpace(parts[1])
// 				break
// 			}
// 		}
// 	}

// 	modifiedBkn := "---\ntype: object_type\nid: x\n---\n\n## Object: x\n**Y**\n"
// 	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(modifiedBkn), 0644)
// 	content2, _ := GenerateChecksumFile(dir)
// 	var hash2 string
// 	for _, line := range strings.Split(content2, "\n") {
// 		if strings.Contains(line, "object_type:x") {
// 			parts := strings.SplitN(line, "  ", 2)
// 			if len(parts) == 2 {
// 				hash2 = strings.TrimSpace(parts[1])
// 				break
// 			}
// 		}
// 	}
// 	if baseHash == hash2 {
// 		t.Error("checksum should change when semantic content changes")
// 	}
// }

// // --- Unit Tests for Loader ---

// func TestDiscoverRootFile_NetworkBknPriority(t *testing.T) {
// 	dir, err := os.MkdirTemp("", "bkn-root-discovery-*")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(dir)

// 	os.WriteFile(filepath.Join(dir, "network.bkn"), []byte("---\ntype: network\nid: net-priority\n---\n"), 0644)
// 	os.WriteFile(filepath.Join(dir, "index.bkn"), []byte("---\ntype: network\nid: index-priority\n---\n"), 0644)

// 	root, err := DiscoverRootFile(dir)
// 	if err != nil {
// 		t.Fatalf("discover root: %v", err)
// 	}
// 	if filepath.Base(root) != "network.bkn" {
// 		t.Errorf("expected network.bkn, got %s", filepath.Base(root))
// 	}
// }

// func TestDiscoverRootFile_MultipleNetworksFails(t *testing.T) {
// 	dir, err := os.MkdirTemp("", "bkn-multiple-networks-*")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(dir)

// 	// Create multiple network files in the same directory (should fail)
// 	os.WriteFile(filepath.Join(dir, "network1.bkn"), []byte("---\ntype: network\nid: net1\n---\n"), 0644)
// 	os.WriteFile(filepath.Join(dir, "network2.bkn"), []byte("---\ntype: network\nid: net2\n---\n"), 0644)

// 	_, err = DiscoverRootFile(dir)
// 	if err == nil {
// 		t.Error("expected error for multiple network files in same directory")
// 	}
// }

// func TestLoadNetwork_ImplicitSameDir(t *testing.T) {
// 	dir, err := os.MkdirTemp("", "bkn-implicit-*")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(dir)

// 	os.WriteFile(filepath.Join(dir, "network.bkn"), []byte("---\ntype: network\nid: implicit-demo\n---\n"), 0644)
// 	os.WriteFile(filepath.Join(dir, "objects.bkn"), []byte("---\ntype: object_type\nid: obj1\n---\n## Object: obj1\n"), 0644)
// 	os.WriteFile(filepath.Join(dir, "relations.bkn"), []byte("---\ntype: relation_type\nid: rel1\n---\n## Relation: rel1\n"), 0644)

// 	net, err := LoadNetwork(dir)
// 	if err != nil {
// 		t.Fatalf("load network: %v", err)
// 	}
// 	if len(net.AllObjects()) != 1 {
// 		t.Errorf("expected 1 object, got %d", len(net.AllObjects()))
// 	}
// 	if len(net.AllRelations()) != 1 {
// 		t.Errorf("expected 1 relation, got %d", len(net.AllRelations()))
// 	}
// }

// // --- Unit Tests for Serializer ---

// func TestSerialize_RoundTrip(t *testing.T) {
// 	doc := BknDocument{
// 		Frontmatter: Frontmatter{
// 			Type:    "object_type",
// 			ID:      "pod",
// 			Name:    "Pod",
// 			Version: "1.0.0",
// 			Tags:    []string{"k8s", "workload"},
// 		},
// 		Objects: []BknObject{
// 			{
// 				ID:          "pod",
// 				Name:        "Pod",
// 				Description: "Kubernetes Pod",
// 				DataProperties: []DataProperty{
// 					{Property: "image", Type: "string"},
// 					{Property: "replicas", Type: "integer"},
// 				},
// 			},
// 		},
// 	}

// 	serialized := Serialize(&doc)
// 	reparsed, err := Parse(serialized, "")
// 	if err != nil {
// 		t.Fatalf("re-parse: %v", err)
// 	}

// 	if reparsed.Frontmatter.ID != doc.Frontmatter.ID {
// 		t.Errorf("ID mismatch: %q vs %q", reparsed.Frontmatter.ID, doc.Frontmatter.ID)
// 	}
// 	if reparsed.Frontmatter.Type != doc.Frontmatter.Type {
// 		t.Errorf("type mismatch: %q vs %q", reparsed.Frontmatter.Type, doc.Frontmatter.Type)
// 	}
// 	if len(reparsed.Objects) != len(doc.Objects) {
// 		t.Errorf("objects count mismatch: %d vs %d", len(reparsed.Objects), len(doc.Objects))
// 	}
// }

// func TestSerialize_Risk(t *testing.T) {
// 	doc := BknDocument{
// 		Frontmatter: Frontmatter{
// 			Type: "risk_type",
// 			ID:   "test_risk",
// 			Name: "Test Risk",
// 		},
// 		Risks: []Risk{
// 			{
// 				ID:           "test_risk",
// 				Name:         "Test Risk",
// 				ControlScope: "production",
// 				PreChecks: []PreCondition{
// 					{Check: "check1"},
// 					{Check: "check2"},
// 				},
// 			},
// 		},
// 	}

// 	serialized := Serialize(&doc)
// 	if !strings.Contains(serialized, "Risk: test_risk") {
// 		t.Error("serialized should contain Risk section")
// 	}
// 	if !strings.Contains(serialized, "Control Scope") {
// 		t.Error("serialized should contain Control Scope")
// 	}
// }

// // --- Unit Tests for Frontmatter ---

// func TestFrontmatter_NewFields(t *testing.T) {
// 	text := `---
// type: object_type
// id: test_obj
// name: Test Object
// version: 1.2.3
// tags: [tag1, tag2]
// description: A test object
// namespace: test-ns
// author: test-author
// created_at: 2024-01-01
// updated_at: 2024-01-02
// ---

// ## Object: test_obj
// Test content
// `
// 	doc, err := Parse(text, "")
// 	if err != nil {
// 		t.Fatalf("parse: %v", err)
// 	}

// 	fm := doc.Frontmatter
// 	if fm.Version != "1.2.3" {
// 		t.Errorf("version: expected 1.2.3, got %q", fm.Version)
// 	}
// 	if len(fm.Tags) != 2 || fm.Tags[0] != "tag1" {
// 		t.Errorf("tags: expected [tag1, tag2], got %v", fm.Tags)
// 	}
// 	if fm.Description != "A test object" {
// 		t.Errorf("description: expected 'A test object', got %q", fm.Description)
// 	}
// 	if fm.Namespace != "test-ns" {
// 		t.Errorf("namespace: expected 'test-ns', got %q", fm.Namespace)
// 	}
// 	if fm.Author != "test-author" {
// 		t.Errorf("author: expected 'test-author', got %q", fm.Author)
// 	}
// 	if !strings.Contains(fm.CreatedAt, "2024-01-01") {
// 		t.Errorf("created_at: expected to contain '2024-01-01', got %q", fm.CreatedAt)
// 	}
// 	if !strings.Contains(fm.UpdatedAt, "2024-01-02") {
// 		t.Errorf("updated_at: expected to contain '2024-01-02', got %q", fm.UpdatedAt)
// 	}
// }
