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
	// Walk up to find repo root (contains examples/k8s-modular/network.bkn)
	dir := cwd
	for i := 0; i < 10; i++ {
		p := filepath.Join(dir, "examples", "k8s-modular", "network.bkn")
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

func TestParseActions(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "k8s-modular", "action_types", "restart_pod.bkn")
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

func TestLoadNetwork(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "k8s-modular", "network.bkn")
	net, err := LoadNetwork(path)
	if err != nil {
		t.Fatalf("load network: %v", err)
	}
	objects := net.AllObjects()
	actions := net.AllActions()
	if len(objects) < 2 {
		t.Errorf("expected objects from includes, got %d", len(objects))
	}
	if len(actions) < 1 {
		t.Errorf("expected actions from includes, got %d", len(actions))
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
	if doc.Frontmatter.Type != "object_type" {
		t.Errorf("expected type object_type, got %q", doc.Frontmatter.Type)
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
	path := filepath.Join(root, "examples", "k8s-modular", "network.bkn")
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
	path := filepath.Join(root, "examples", "k8s-modular", "network.bkn")
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
	path := filepath.Join(root, "examples", "k8s-modular", "network.bkn")
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
	os.Mkdir(filepath.Join(dir, "object_types"), 0755)
	os.WriteFile(filepath.Join(dir, "object_types", "pod.bkn"), []byte("---\ntype: object_type\nid: pod\nname: Pod\nnetwork: k8s\n---\n\n## Object: pod\n**Pod**\n"), 0644)

	content, err := GenerateChecksumFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	// New format: object_type:pod  sha256:hash
	if !strings.Contains(content, "sha256:") || !strings.Contains(content, "object_type:pod") {
		t.Errorf("unexpected content: %s", content)
	}
	ok, errs := VerifyChecksumFile(dir)
	if !ok {
		t.Errorf("verify failed: %v", errs)
	}
}

func TestChecksumNormalization_BlankLinesAndWhitespace(t *testing.T) {
	dir, err := os.MkdirTemp("", "bkn-checksum-norm-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseBkn := "---\ntype: object_type\nid: x\n---\n\n## Object: x\n**X**\n"
	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(baseBkn), 0644)

	content, err := GenerateChecksumFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Extract the checksum line for object_type:x (new format)
	var baseHash string
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "object_type:x") {
			parts := strings.SplitN(line, "  ", 2)
			if len(parts) == 2 {
				baseHash = strings.TrimSpace(parts[1])
				break
			}
		}
	}
	if baseHash == "" {
		t.Fatal("could not find object_type:x checksum in generated content")
	}

	// Same semantic content with extra blank lines - checksum should match
	withBlankLines := "---\ntype: object_type\nid: x\n---\n\n\n## Object: x\n\n**X**\n\n"
	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(withBlankLines), 0644)
	content2, _ := GenerateChecksumFile(dir)
	var hash2 string
	for _, line := range strings.Split(content2, "\n") {
		if strings.Contains(line, "object_type:x") {
			parts := strings.SplitN(line, "  ", 2)
			if len(parts) == 2 {
				hash2 = strings.TrimSpace(parts[1])
				break
			}
		}
	}
	if baseHash != hash2 {
		t.Errorf("checksum changed with blank lines only: %q vs %q", baseHash, hash2)
	}

	// Same semantic content with CRLF and trailing spaces - checksum should match
	withCRLF := "---\r\ntype: object_type\r\nid: x\r\n---\r\n\r\n## Object: x\r\n**X**   \r\n"
	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(withCRLF), 0644)
	content3, _ := GenerateChecksumFile(dir)
	var hash3 string
	for _, line := range strings.Split(content3, "\n") {
		if strings.Contains(line, "object_type:x") {
			parts := strings.SplitN(line, "  ", 2)
			if len(parts) == 2 {
				hash3 = strings.TrimSpace(parts[1])
				break
			}
		}
	}
	if baseHash != hash3 {
		t.Errorf("checksum changed with CRLF/trailing space only: %q vs %q", baseHash, hash3)
	}

	// Restore original and verify still passes
	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(baseBkn), 0644)
	GenerateChecksumFile(dir)
	ok, errs := VerifyChecksumFile(dir)
	if !ok {
		t.Errorf("verify failed after round-trip: %v", errs)
	}
}

func TestChecksumNormalization_SemanticChangeAltersChecksum(t *testing.T) {
	dir, err := os.MkdirTemp("", "bkn-checksum-semantic-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseBkn := "---\ntype: object_type\nid: x\n---\n\n## Object: x\n**X**\n"
	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(baseBkn), 0644)
	content, _ := GenerateChecksumFile(dir)
	var baseHash string
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "object_type:x") {
			parts := strings.SplitN(line, "  ", 2)
			if len(parts) == 2 {
				baseHash = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	// Semantic change: different object name
	modifiedBkn := "---\ntype: object_type\nid: x\n---\n\n## Object: x\n**Y**\n"
	os.WriteFile(filepath.Join(dir, "test.bkn"), []byte(modifiedBkn), 0644)
	content2, _ := GenerateChecksumFile(dir)
	var hash2 string
	for _, line := range strings.Split(content2, "\n") {
		if strings.Contains(line, "object_type:x") {
			parts := strings.SplitN(line, "  ", 2)
			if len(parts) == 2 {
				hash2 = strings.TrimSpace(parts[1])
				break
			}
		}
	}
	if baseHash == hash2 {
		t.Error("checksum should change when semantic content changes")
	}
}

func TestChecksumNormalization_BkndWhitespace(t *testing.T) {
	dir, err := os.MkdirTemp("", "bkn-checksum-bknd-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	baseBknd := "---\ntype: data\nobject: x\n---\n\n## Data\n\n| a | b |\n|---|---|\n| 1 | 2 |\n"
	os.WriteFile(filepath.Join(dir, "data.bknd"), []byte(baseBknd), 0644)
	content, _ := GenerateChecksumFile(dir)
	var baseHash string
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "data.bknd") {
			parts := strings.SplitN(line, "  ", 2)
			if len(parts) == 2 {
				baseHash = strings.TrimSpace(parts[0])
				break
			}
		}
	}

	// Same table with extra blank lines and padding
	withWhitespace := "---\ntype: data\nobject: x\n---\n\n## Data\n\n\n|  a  |  b  |\n|-----|-----|\n|  1  |  2  |\n\n"
	os.WriteFile(filepath.Join(dir, "data.bknd"), []byte(withWhitespace), 0644)
	content2, _ := GenerateChecksumFile(dir)
	var hash2 string
	for _, line := range strings.Split(content2, "\n") {
		if strings.Contains(line, "data.bknd") {
			parts := strings.SplitN(line, "  ", 2)
			if len(parts) == 2 {
				hash2 = strings.TrimSpace(parts[0])
				break
			}
		}
	}
	if baseHash != hash2 {
		t.Errorf("checksum changed with bknd whitespace only: %q vs %q", baseHash, hash2)
	}
}

// --- Loader dir discovery tests ---

func TestLoadNetwork_DirDiscoversRoot(t *testing.T) {
	root := repoRoot(t)
	dir := filepath.Join(root, "examples", "k8s-network")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Skip("examples/k8s-network not found")
		return
	}
	net, err := LoadNetwork(dir)
	if err != nil {
		t.Fatalf("load network dir: %v", err)
	}
	if net.Root.Frontmatter.ID != "k8s-network" {
		t.Errorf("expected id k8s-network, got %q", net.Root.Frontmatter.ID)
	}
	if len(net.AllObjects()) < 3 {
		t.Errorf("expected at least 3 objects, got %d", len(net.AllObjects()))
	}
}

func TestDiscoverRootFile_NetworkBknPriority(t *testing.T) {
	dir, err := os.MkdirTemp("", "bkn-root-discovery-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "network.bkn"), []byte("---\ntype: network\nid: net-priority\n---\n"), 0644)
	os.WriteFile(filepath.Join(dir, "index.bkn"), []byte("---\ntype: network\nid: index-priority\n---\n"), 0644)

	root, err := DiscoverRootFile(dir)
	if err != nil {
		t.Fatalf("discover root: %v", err)
	}
	if filepath.Base(root) != "network.bkn" {
		t.Errorf("expected network.bkn, got %s", filepath.Base(root))
	}
}

func TestLoadNetwork_ImplicitSameDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "bkn-implicit-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "network.bkn"), []byte("---\ntype: network\nid: implicit-demo\n---\n"), 0644)
	os.WriteFile(filepath.Join(dir, "objects.bkn"), []byte("---\ntype: object_type\nid: obj1\n---\n## Object: obj1\n"), 0644)
	os.WriteFile(filepath.Join(dir, "relations.bkn"), []byte("---\ntype: relation_type\nid: rel1\n---\n## Relation: rel1\n"), 0644)

	net, err := LoadNetwork(dir)
	if err != nil {
		t.Fatalf("load network: %v", err)
	}
	if len(net.AllObjects()) != 1 {
		t.Errorf("expected 1 object, got %d", len(net.AllObjects()))
	}
	if len(net.AllRelations()) != 1 {
		t.Errorf("expected 1 relation, got %d", len(net.AllRelations()))
	}
}

func TestDiscoverRootFile_MultipleNetworksFails(t *testing.T) {
	dir, err := os.MkdirTemp("", "bkn-multi-root-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.WriteFile(filepath.Join(dir, "a.bkn"), []byte("---\ntype: network\nid: a\n---\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.bkn"), []byte("---\ntype: network\nid: b\n---\n"), 0644)

	_, err = DiscoverRootFile(dir)
	if err == nil {
		t.Error("expected error for multiple network roots")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "multiple network roots") {
		t.Errorf("expected 'multiple network roots' in error, got %q", err.Error())
	}
}
